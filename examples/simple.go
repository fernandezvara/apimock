package main

import (
	"bytes"
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
	a := apimock.NewAPIMock(true, logrus.New(), "xml")

	route1 := apimock.URIMock{
		Method:   "GET",
		URI:      "/hello",
		Response: "world",
	}

	route2 := apimock.URIMock{
		Method: "GET",
		URI:    "/struct",
		Response: tt{
			Name: "TestName",
			Age:  39,
		},
	}

	route3 := apimock.URIMock{
		Method:   "GET",
		URI:      "/function",
		Response: testFunction(),
	}

	a.AddMock(&route1)
	a.AddMock(&route2)
	a.AddMock(&route3)
	a.Start()
	defer a.Stop()

	httpCall(fmt.Sprintf("%s%s", a.URL, "/hello"))
	httpCall(fmt.Sprintf("%s%s", a.URL, "/struct"))
	httpCall(fmt.Sprintf("%s%s", a.URL, "/function"))
}

func testFunction() *tt {
	return &tt{
		Name: "TestFunctionName",
		Age:  390,
	}
}

func httpCall(uri string) {
	buf := new(bytes.Buffer)
	httpClient := new(http.Client)

	req, err := http.NewRequest("GET", uri, buf)
	assert(err)
	res, err := httpClient.Do(req)
	assert(err)
	defer res.Body.Close()
	objectByte, err := ioutil.ReadAll(res.Body)
	assert(err)

	fmt.Println(string(objectByte))
}

func assert(err error) {
	if err != nil {
		fmt.Println("err:", err)
	}
}
