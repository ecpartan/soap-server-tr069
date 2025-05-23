package main

import (
	"net/http"
	"os"

	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/server"
	"github.com/ecpartan/soap-server-tr069/tasks"
)

// RunServer run the server
func RunServer() {
	soapServer := server.NewServer()

	soapServer.RegisterHandler("/addtask", soapServer.PerformConReq)
	soapServer.RegisterHandler("/", soapServer.MainHandler)

	err := http.ListenAndServe(":8089", soapServer)

	logger.LogDebug("exiting with error", err)
}
func main() {

	logger.InitLogger(os.Stdout)
	logger.LogDebug("Starting server")

	tasks.InitTasks()
	RunServer()

}
