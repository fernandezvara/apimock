package apimock

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	A string `json:"a" xml:"a"`
	B int    `json:"b,omitempty" xml:"b,omitempty"`
}

var helloMock = &URIMock{
	Method:   "GET",
	URI:      "/hello",
	Response: "world",
}

var structMock = &URIMock{
	Method: "GET",
	URI:    "/struct1",
	Response: testStruct{
		A: "aaa",
		B: 111,
	},
}

var structMock2 = &URIMock{
	Method: "GET",
	URI:    "/struct2",
	Response: testStruct{
		A: "aaa",
	},
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
	assert.NotEmpty(t, api.URL, "it must have an URL")

	response, res := httpCall("GET", fmt.Sprintf("%s/hello", api.URL))
	assert.Equal(t, "\"world\"\n", string(response), "It must return the expected result")
	assert.Equal(t, 200, res.StatusCode, "It must have status code 200")
	assert.Equal(t, "*", res.Header.Get("Access-Control-Allow-Origin"), "It must have CORS Headers")
	api.Stop()
	assert.Equal(t, "", api.URL, "URL must be empty")
}

func TestClientNoCORS(t *testing.T) {
	api := NewAPIMock(false, logrus.New(), "json")
	api.AddMock(helloMock)
	assert.Len(t, api.URIMocks, 1, "It must have 1 URIMocks defined")
	api.Start()
	response, res := httpCall("GET", fmt.Sprintf("%s/hello", api.URL))
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
	response, res := httpCall("GET", fmt.Sprintf("%s/hello", api.URL))
	assert.Equal(t, "<string>world</string>", string(response), "It must return the expected result")
	assert.Equal(t, 200, res.StatusCode, "It must have status code 200")
	api.Stop()
}

func TestClientOptions(t *testing.T) {
	api := NewAPIMock(true, logrus.New(), "json")
	api.AddMock(helloMock)
	assert.Len(t, api.URIMocks, 1, "It must have 1 URIMocks defined")
	api.Start()
	_, res := httpCall("OPTIONS", fmt.Sprintf("%s/hello", api.URL))
	assert.Equal(t, 200, res.StatusCode, "It must have status code 200")
	api.Stop()
}

func TestClientHTTPNotFound(t *testing.T) {
	api := NewAPIMock(true, logrus.New(), "json")
	api.AddMock(helloMock)
	assert.Len(t, api.URIMocks, 1, "It must have 1 URIMocks defined")
	api.Start()
	_, res := httpCall("GET", fmt.Sprintf("%s/hello-not-found", api.URL))
	assert.Equal(t, 404, res.StatusCode, "It must have status code 404")
	api.Stop()
}

func TestClientStruct(t *testing.T) {
	var err error
	types := []string{"json", "xml"}
	for _, _type := range types {

		api := NewAPIMock(true, logrus.New(), _type)
		api.AddMock(helloMock)
		api.AddMock(structMock)
		api.AddMock(structMock2)
		assert.Len(t, api.URIMocks, 3, "It must have 1 URIMocks defined")
		api.Start()

		var s testStruct
		if _type == "json" {
			response11, _ := httpCall("GET", fmt.Sprintf("%s/struct1", api.URL))
			assert.Equal(t, "{\"a\":\"aaa\",\"b\":111}\n", string(response11), "Response not expected")

			response12, _ := httpCall("GET", fmt.Sprintf("%s/struct2", api.URL))
			assert.Equal(t, "{\"a\":\"aaa\"}\n", string(response12), "Response not expected")
			err = json.Unmarshal(response11, &s)
		} else {
			response11, _ := httpCall("GET", fmt.Sprintf("%s/struct1", api.URL))
			assert.Equal(t, "<testStruct><a>aaa</a><b>111</b></testStruct>", string(response11), "Response not expected")

			response12, _ := httpCall("GET", fmt.Sprintf("%s/struct2", api.URL))
			assert.Equal(t, "<testStruct><a>aaa</a></testStruct>", string(response12), "Response not expected")
			err = xml.Unmarshal(response11, &s)
		}
		assert.Nil(t, err, "It unmarshals without error")
		assert.Equal(t, "aaa", s.A, "Data not unmarshalled correctly")
		assert.Equal(t, 111, s.B, "Data not unmarshalled correctly")
		api.Stop()
	}
}

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
