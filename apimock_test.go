package apimock

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	A string `json:"a" xml:"a"`
	B int    `json:"b,omitempty" xml:"b,omitempty"`
}

var helloMock = &URIMock{
	Method:     "GET",
	URI:        "/hello",
	StatusCode: http.StatusOK,
	Response:   "world",
}

var helloPostMock = &URIMock{
	Method:     "POST",
	URI:        "/hello",
	StatusCode: 201,
	Response:   "",
}

var structMock = &URIMock{
	Method:     "GET",
	URI:        "/struct1",
	StatusCode: http.StatusOK,
	Response: testStruct{
		A: "aaa",
		B: 111,
	},
}

var structMock2 = &URIMock{
	Method:     "GET",
	URI:        "/struct2",
	StatusCode: http.StatusOK,
	Response: testStruct{
		A: "aaa",
	},
}

var structRAW = &URIMock{
	Method:     "GET",
	URI:        "/raw",
	StatusCode: http.StatusOK,
	Response:   []byte(`{"name":"someName","surname":"someSurname"}`),
}

func TestClient(t *testing.T) {
	assert.Panics(t, func() {
		NewAPIMock(true, logrus.New(), "panic!")
	}, "It must panic if wrong type")

	api := NewAPIMock(true, logrus.New(), "json")
	assert.IsType(t, api, new(APIMock), "It must instance an APIMock struct")
	assert.Equal(t, api.CORSEnabled, true, "Cors must be TRUE")
	assert.Equal(t, api.Log, logrus.New(), "Logger must be set correctly")
	assert.Equal(t, api.Type, "json", "Type must be 'json'")
}

func TestClientStartStop(t *testing.T) {
	api := NewAPIMock(true, logrus.New(), "json")
	api.AddMock(helloMock)
	assert.Len(t, api.URIMocks, 1, "It must have 1 URIMocks defined")
	api.Start()
	assert.NotEmpty(t, api.URL(), "it must have an URL")

	response, res := httpCall("GET", fmt.Sprintf("%s/hello", api.URL()))
	assert.Equal(t, "\"world\"\n", string(response), "It must return the expected result")
	assert.Equal(t, 200, res.StatusCode, "It must have status code 200")
	assert.Equal(t, "*", res.Header.Get("Access-Control-Allow-Origin"), "It must have CORS Headers")

	// Test Port() &  Protocol()
	assert.Equal(t, api.URL(), fmt.Sprintf("%s://127.0.0.1:%d", api.Protocol(), api.Port()))

	api.Stop()
}

func TestClientNoCORS(t *testing.T) {
	api := NewAPIMock(false, logrus.New(), "json")
	api.AddMock(helloMock)
	assert.Len(t, api.URIMocks, 1, "It must have 1 URIMocks defined")
	api.Start()
	response, res := httpCall("GET", fmt.Sprintf("%s/hello", api.URL()))
	assert.Equal(t, "\"world\"\n", string(response), "It must return the expected result")
	assert.Equal(t, 200, res.StatusCode, "It must have status code 200")
	assert.Empty(t, res.Header.Get("Access-Control-Allow-Origin"), "Header must be nil if CORS disabled")
	api.Stop()
}

func TestClientXML(t *testing.T) {
	api := NewAPIMock(true, logrus.New(), "xml")
	api.AddMock(helloMock)
	assert.Len(t, api.URIMocks, 1, "It must have 1 URIMocks defined")
	api.Start()
	assert.Equal(t, api.Type, "xml", "Type must be 'xml'")
	response, res := httpCall("GET", fmt.Sprintf("%s/hello", api.URL()))
	assert.Equal(t, "<string>world</string>", string(response), "It must return the expected result")
	assert.Equal(t, 200, res.StatusCode, "It must have status code 200")
	api.Stop()
}

func TestClientOptions(t *testing.T) {
	api := NewAPIMock(true, logrus.New(), "json")
	api.AddMock(helloMock)
	assert.Len(t, api.URIMocks, 1, "It must have 1 URIMocks defined")
	api.Start()
	_, res := httpCall("OPTIONS", fmt.Sprintf("%s/hello", api.URL()))
	assert.Equal(t, 200, res.StatusCode, "It must have status code 200")
	api.Stop()
}

func TestClientHTTPNotFound(t *testing.T) {
	api := NewAPIMock(true, logrus.New(), "json")
	api.AddMock(helloMock)
	assert.Len(t, api.URIMocks, 1, "It must have 1 URIMocks defined")
	api.Start()
	_, res := httpCall("GET", fmt.Sprintf("%s/hello-not-found", api.URL()))
	assert.Equal(t, 404, res.StatusCode, "It must have status code 404")
	api.Stop()
}

func TestClientStruct(t *testing.T) {
	var err error
	types := []string{"json", "xml"}
	for _, _type := range types {

		api := NewAPIMock(true, logrus.New(), _type)
		api.AddMock(helloMock)
		api.AddMock(helloPostMock)
		api.AddMock(structMock)
		api.AddMock(structMock2)
		assert.Len(t, api.URIMocks, 4, "It must have 4 URIMocks defined")
		api.Start()

		var s testStruct
		switch _type {
		case "json":
			response11, _ := httpCall("GET", fmt.Sprintf("%s/struct1", api.URL()))
			assert.Equal(t, "{\"a\":\"aaa\",\"b\":111}\n", string(response11), "Response not expected")

			response12, _ := httpCall("GET", fmt.Sprintf("%s/struct2", api.URL()))
			assert.Equal(t, "{\"a\":\"aaa\"}\n", string(response12), "Response not expected")
			err = json.Unmarshal(response11, &s)
		case "xml":
			response11, _ := httpCall("GET", fmt.Sprintf("%s/struct1", api.URL()))
			assert.Equal(t, "<testStruct><a>aaa</a><b>111</b></testStruct>", string(response11), "Response not expected")

			response12, _ := httpCall("GET", fmt.Sprintf("%s/struct2", api.URL()))
			assert.Equal(t, "<testStruct><a>aaa</a></testStruct>", string(response12), "Response not expected")
			err = xml.Unmarshal(response11, &s)
		}

		assert.Nil(t, err, "It unmarshals without error")
		assert.Equal(t, "aaa", s.A, "Data not unmarshalled correctly")
		assert.Equal(t, 111, s.B, "Data not unmarshalled correctly")

		_, res := httpCall("POST", fmt.Sprintf("%s/hello", api.URL()))
		assert.Equal(t, http.StatusCreated, res.StatusCode, "StatusCode mismatch")

		api.Stop()
	}
}

func TestClientRAW(t *testing.T) {
	type localStruct struct {
		Name    string `json:"name"`
		Surname string `json:"surname"`
	}

	api := NewAPIMock(true, logrus.New(), "json")
	api.AddMock(structRAW)
	assert.Len(t, api.URIMocks, 1, "It must have 4 URIMocks defined")
	api.Start()

	response, _ := httpCall("GET", fmt.Sprintf("%s/raw", api.URL()))

	var l localStruct
	err := json.Unmarshal(response, &l)
	assert.Nil(t, err)
	assert.Equal(t, "someName", l.Name)
	assert.Equal(t, "someSurname", l.Surname)

}

func TestAdd(t *testing.T) {
	api := NewAPIMock(true, logrus.New(), "json")
	api.Add("GET", "/hi", 200, []byte("ho"))
	assert.Len(t, api.URIMocks, 1)
	api.Start()

	bytes, res := httpCall("GET", fmt.Sprintf("%s/hi", api.URL()))
	assert.Equal(t, "ho", string(bytes))
	assert.Equal(t, 200, res.StatusCode)
}

// Test helpers
func httpCall(_type, uri string) ([]byte, *http.Response) {
	buf := new(bytes.Buffer)
	httpClient := new(http.Client)

	req, err := http.NewRequest(_type, uri, buf)
	isErr(err)
	res, err := httpClient.Do(req)
	isErr(err)
	defer res.Body.Close()
	objectByte, err := ioutil.ReadAll(res.Body)
	isErr(err)

	return objectByte, res
}

func isErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
