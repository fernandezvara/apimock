package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/fernandezvara/apimock"
)

type tt struct {
	Name string
	Age  int
}

func main() {
	a := apiMock() // setup the test
	a.Start()      // start local test server
	defer a.Stop() // defer stop

	fmt.Println(string(httpCall(fmt.Sprintf("%s%s", a.URL, "/hello"))))

	// Unmarshal response
	var t1, t2 tt

	err := json.Unmarshal(httpCall(fmt.Sprintf("%s%s", a.URL, "/struct")), &t1)
	assert(err)
	fmt.Println("t1:", t1)

	err = json.Unmarshal(httpCall(fmt.Sprintf("%s%s", a.URL, "/function")), &t2)
	assert(err)
	fmt.Println("t2:", t2)

}

func apiMock() *apimock.APIMock {
	a := apimock.NewAPIMock(true, logrus.New(), "json")

	route1 := apimock.URIMock{
		Method:     "GET",
		URI:        "/hello",
		StatusCode: http.StatusOK,
		Response:   "world",
	}

	route2 := apimock.URIMock{
		Method:     "GET",
		URI:        "/struct",
		StatusCode: http.StatusOK,
		Response: tt{
			Name: "TestName",
			Age:  39,
		},
	}

	route3 := apimock.URIMock{
		Method:     "GET",
		URI:        "/function",
		StatusCode: http.StatusOK,
		Response:   testFunction(),
	}

	// add routes
	a.AddMock(&route1)
	a.AddMock(&route2)
	a.AddMock(&route3)

	return a
}

func testFunction() *tt {
	return &tt{
		Name: "TestFunctionName",
		Age:  390,
	}
}

func httpCall(uri string) []byte {
	buf := new(bytes.Buffer)
	httpClient := new(http.Client)

	req, err := http.NewRequest("GET", uri, buf)
	assert(err)
	res, err := httpClient.Do(req)
	assert(err)
	defer res.Body.Close()
	objectByte, err := ioutil.ReadAll(res.Body)
	assert(err)

	return objectByte
}

func assert(err error) {
	if err != nil {
		fmt.Println("err:", err)
	}
}
