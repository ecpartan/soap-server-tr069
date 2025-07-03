package methods

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	p "github.com/ecpartan/soap-server-tr069/internal/parsemap"
	"github.com/ecpartan/soap-server-tr069/jrpc2/methods/response"
	logger "github.com/ecpartan/soap-server-tr069/log"
	repository "github.com/ecpartan/soap-server-tr069/repository/cache"
	"github.com/ecpartan/soap-server-tr069/tasks"
	dac "github.com/xinsnake/go-http-digest-auth-client"
)

func Get(ctx context.Context, req map[string]any) ([]byte, error) {

	if sn, ok := req["Serial"].(string); !ok {
		return nil, ErrNotfoundSerial
	} else {
		if sn == "" {
			return nil, ErrNotfoundSerial
		}
		tree := repository.GetCache().Get(sn)

		logger.LogDebug("tree", tree)

		if tree == nil {
			return nil, ErrNotfoundTree
		}

		if ret, err := json.Marshal(tree); err != nil {
			return nil, err
		} else {
			return ret, nil
		}
	}
}

func WriteJSON(w http.ResponseWriter, code int, obj any) {
	data, err := json.Marshal(obj)
	if err != nil {
		// Fallback in case of marshaling error. This should not happen, but
		// ensures the client gets a loggable reply from a broken server.
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(code)
	w.Write(data)
}

var wg sync.WaitGroup

func watchChannel(sn string, ch <-chan *response.RetScriptTask, wg *sync.WaitGroup) {
	defer wg.Done()
	logger.LogDebug("Watching channel: ", sn)

	for channelValue := range ch {
		fmt.Printf("Channel '%s' with value: '%s'\n", sn, channelValue.Code)
		response.EndTaskResponse[sn] = *channelValue
	}
}

func AddScriptTask(ctx context.Context, req map[string]any) ([]byte, error) {

	logger.LogDebug("AddScriptTask", req)

	sn := p.GetSnScript(req)

	logger.LogDebug("sn", sn)

	if sn == "" {
		return nil, ErrNotfoundSerial
	}

	err := tasks.AddToScripter(sn, req)
	if err != nil {
		return nil, err
	}

	tree := repository.GetCache().Get(sn)

	if tree == nil {
		logger.LogDebug("mp is nil")
		return nil, ErrNotfoundTree
	}

	url := p.GetXML(tree, "InternetGatewayDevice.ManagementServer.ConnectionRequestURL.Value")

	if crURL, ok := url.(string); ok {
		logger.LogDebug("crURL", crURL)
		dr := dac.NewRequest("", "", "GET", crURL, "")
		_, err := dr.Execute()

		if err != nil {
			return nil, NewAppError("500", err.Error())
		}
	} else {
		return nil, NewAppError("500", "crURL is nil")
	}

	wg.Add(1)
	response.EndTaskChansMap[sn] = make(chan *response.RetScriptTask, 1)
	go watchChannel(sn, response.EndTaskChansMap[sn], &wg)
	wg.Wait()
	logger.LogDebug("Task finished")

	if resp, ok := response.EndTaskResponse[sn]; !ok {
		return nil, ErrNotfoundTree
	} else {
		if ret, err := json.Marshal(resp.Code); err != nil {
			return nil, err
		} else {
			return ret, nil
		}
	}
}
