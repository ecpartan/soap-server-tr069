package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/server"
	"github.com/ecpartan/soap-server-tr069/tasks"
	//"github.com/globusdigital/soap"
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

			return
		},
	)

	err := http.ListenAndServe(":8089", soapServer)

	log.Println("exiting with error", err)
}

func main() {
	//soap.InitCache(100)

	logger.InitLogger(os.Stdout)
	logger.LogDebug("Starting server")

	tasks.InitTasks()
	RunServer()

}
