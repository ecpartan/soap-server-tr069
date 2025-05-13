package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/server"
	"github.com/ecpartan/soap-server-tr069/tasks"

	//"github.com/globusdigital/soap"
	dac "github.com/xinsnake/go-http-digest-auth-client"
)

// FooResponse a simple response
type FooResponse struct {
	Bar string
}

// RunServer run a little demo server
func RunServer() {
	soapServer := server.NewServer()

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
				log.log.Fatalln(err)
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

	log.Println("exiting with error", err)
}

func main() {
	//soap.InitCache(100)
	log.InitLogger(os.Stdout)
	log.LogDebug("Starting server")

	tasks.InitTasks()
	RunServer()

}
