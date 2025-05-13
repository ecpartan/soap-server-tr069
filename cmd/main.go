package main

import (
	"fmt"
	"log"
	"net/http"
	"soap"

	//"github.com/globusdigital/soap"
	dac "github.com/xinsnake/go-http-digest-auth-client"
)

// FooResponse a simple response
type FooResponse struct {
	Bar string
}

// RunServer run a little demo server
func RunServer() {
	soapServer := soap.NewServer()
	soapServer.Log = log.Println

	soapServer.RegisterHandler(
		"/",
		"", // SOAPAction
		"", // tagname of soap body content
		// RequestFactoryFunc - give the server sth. to unmarshal the request into
		func() interface{} {
			return &FooResponse{}
		},
		// OperationHandlerFunc - do something
		func(request interface{}, w http.ResponseWriter, httpRequest *http.Request) (response interface{}, err error) {
			fmt.Println("exiting")

			fooResponse := &FooResponse{
				Bar: "Hello",
			}
			response = fooResponse
			return
		},
	)
	soapServer.RegisterHandler(
		"/request",
		"", // SOAPAction
		"", // tagname of soap body content
		// RequestFactoryFunc - give the server sth. to unmarshal the request into
		func() interface{} {
			return &FooResponse{}
		},
		// OperationHandlerFunc - do something
		func(request interface{}, w http.ResponseWriter, httpRequest *http.Request) (response interface{}, err error) {
			fmt.Println("exiting")

			t := dac.NewTransport("", "")
			req, err := http.NewRequest("POST", "http://localhost:8999/", nil)

			if err != nil {
				log.Fatalln(err)
			}

			resp, err := t.RoundTrip(req)
			if err != nil {
				log.Fatalln(err)
			}

			defer resp.Body.Close()

			fmt.Println(resp)
			response = nil
			return
		},
	)

	err := http.ListenAndServe(":8089", soapServer)

	fmt.Println("exiting with error", err)
}

func main() {
	//soap.InitCache(100)
	soap.InitTasks()
	soap.RedisStart()
	RunServer()

}
