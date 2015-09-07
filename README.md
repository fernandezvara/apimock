# apimock
Simple API mocker helper for tests

This library allows to mock easily any API to use in our test.

Normally when creating an API client we need to access many times, so instancing a local api mock will allow us to override problems like rate limiting, double asset creations or just submit any data incorrectly.

API mocks allows to pass any interface decodeable by json or xml unmarshallers.

## Simple usage:

```
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/fernandezvara/apimock"
)

func main() {
	a := apimock.NewAPIMock(true, logrus.New(), "json")

	route1 := apimock.URIMock{
		Method:     "GET",
		URI:        "/hello",
		StatusCode: http.StatusOK,
		Response:   "world",
	}

	a.AddMock(&route1)
	a.Start()
	defer a.Stop()

	b, r := httpCall("GET", fmt.Sprintf("%s/hello", a.URL))
	fmt.Println("response:", string(b))
	fmt.Println("status  :", r.StatusCode)
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
```

*Some more samples on the `/examples` folder.
