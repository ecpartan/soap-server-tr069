package devsoap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ecpartan/soap-server-tr069/internal/apperror"
	p "github.com/ecpartan/soap-server-tr069/internal/parsemap"
	logger "github.com/ecpartan/soap-server-tr069/log"
	repository "github.com/ecpartan/soap-server-tr069/repository/cache"
	"github.com/ecpartan/soap-server-tr069/server/handlers"
	"github.com/ecpartan/soap-server-tr069/tasks/scripter"
	"github.com/ecpartan/soap-server-tr069/tasks/tasker"
	"github.com/julienschmidt/httprouter"
	dac "github.com/xinsnake/go-http-digest-auth-client"
)

type handlerCR struct {
	Cache     *repository.Cache
	execTasks *tasker.Tasker
}

func NewHandlerCR(Cache *repository.Cache, execTasks *tasker.Tasker) handlers.Handler {
	return &handlerCR{
		Cache:     Cache,
		execTasks: execTasks,
	}
}

func (h *handlerCR) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, "/addtask", apperror.Middleware(h.PerformConReq))
}

/*
func ExecuteCR(bytes []byte, cache *repository.Cache) error {

	var getScript map[string]any
	err := json.Unmarshal(bytes, &getScript)
	if err != nil {
		return fmt.Errorf("failed unmarshal task CR: %v", err)
	}

	logger.LogDebug("body_task", getScript)

	script := p.GetXML(getScript, "Script")
	serial := p.GetXML(script, "Serial")
	logger.LogDebug("sn", serial, reflect.TypeOf(serial))

	sn := serial.(string)
	if sn == "" {
		return fmt.Errorf("failed SN in CR: %v", err)
	}
	err = tasks.AddToScripter(sn, script.(map[string]any))

	if err != nil {
		return fmt.Errorf("failed add task CR: %v", err)
	}

	tree := cache.Get(sn)

	if tree == nil {
		logger.LogDebug("mp is nil")
		return fmt.Errorf("failed no found SN in DB: %v", err)
	}

	logger.LogDebug("mp", tree)
	url := p.GetXML(tree, "InternetGatewayDevice.ManagementServer.ConnectionRequestURL.Value")

	if crURL, ok := url.(string); ok {
		logger.LogDebug("crURL", crURL)
		dr := dac.NewRequest("", "", "GET", crURL, "")
		_, err := dr.Execute()

		if err != nil {
			return fmt.Errorf("error in execute connection request: %v", err)
		}
	} else {
		return fmt.Errorf("no found addres for this device by SN: %v", err)
	}

	return nil
}*/

// Connect to the server
// @Summary Perfform a CR
// @Tags SOAP
// @Success 200 {object} tasks.Task
// @Router  /addtask [post]
func (h *handlerCR) PerformConReq(w http.ResponseWriter, r *http.Request) error {

	logger.LogDebug("addtask")
	soapRequestBytes, err := io.ReadAll(r.Body)

	if err != nil {
		return fmt.Errorf("could not read POST: %v", err)
	}

	var getScript map[string]any
	err = json.Unmarshal(soapRequestBytes, &getScript)
	if err != nil {
		return fmt.Errorf("failed unmarshal task CR: %v", err)
	}

	logger.LogDebug("body_task", getScript)

	sn := p.GetSnScript(getScript)
	if sn == "" {
		return fmt.Errorf("failed SN in CR: %v", err)
	}
	err = scripter.AddToScripter(sn, getScript)

	if err != nil {
		return fmt.Errorf("failed add task CR: %v", err)
	}

	tree := h.Cache.Get(sn)

	if tree == nil {
		logger.LogDebug("mp is nil")
		return fmt.Errorf("failed no found SN in DB: %v", err)
	}

	logger.LogDebug("mp", tree)
	url := p.GetXML(tree, "InternetGatewayDevice.ManagementServer.ConnectionRequestURL.Value")

	if crURL, ok := url.(string); ok {
		logger.LogDebug("crURL", crURL)
		dr := dac.NewRequest("", "", "GET", crURL, "")
		_, err := dr.Execute()

		if err != nil {
			return fmt.Errorf("error in execute connection request: %v", err)
		}
	} else {
		return fmt.Errorf("no found addres for this device by SN: %v", err)
	}

	return nil

}
