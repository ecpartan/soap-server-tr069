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
	"github.com/ecpartan/soap-server-tr069/tasks"
	"github.com/julienschmidt/httprouter"
	dac "github.com/xinsnake/go-http-digest-auth-client"
)

type handlerCR struct {
	Cache *repository.Cache
}

func NewHandlerCR(Cache *repository.Cache) handlers.Handler {
	return &handlerCR{
		Cache: Cache,
	}
}

func (h *handlerCR) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, "/addtask", apperror.Middleware(h.PerformConReq))
}

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

	var serial string
	serial, err = tasks.AddToScripter(getScript)
	if err != nil || serial == "" {
		return fmt.Errorf("failed SN in CR: %v", err)
	}

	mp := h.Cache.Get(serial)

	if mp == nil {
		logger.LogDebug("mp is nil")
		return fmt.Errorf("failed no found SN in DB: %v", err)
	}
	logger.LogDebug("mp", mp)
	url := p.GetXML(mp, "InternetGatewayDevice.ManagementServer.ConnectionRequestURL.Value")

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
