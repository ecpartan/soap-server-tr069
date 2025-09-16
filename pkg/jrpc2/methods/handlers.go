package methods

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/ecpartan/soap-server-tr069/internal/config"
	p "github.com/ecpartan/soap-server-tr069/internal/parsemap"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/pkg/jrpc2/methods/response"
	"github.com/ecpartan/soap-server-tr069/pkg/jrpc2/mwdto"
	repository "github.com/ecpartan/soap-server-tr069/repository/cache"
	"github.com/ecpartan/soap-server-tr069/soap"
	"github.com/ecpartan/soap-server-tr069/tasks/scripter"
	dac "github.com/xinsnake/go-http-digest-auth-client"
)

const (
	MethodAddScript = "AddScript"
	MethodGetTree   = "GetTree"
)

func Get(ctx context.Context, dto mwdto.Mwdto) ([]byte, error) {

	if sn, ok := dto.Reqw["Serial"].(string); !ok {
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
		logger.LogDebug("Channel '%s' with value: '%s'\n", sn, channelValue.Code)
		response.EndTaskResponse[sn] = *channelValue
	}
}

func AddScriptTask(ctx context.Context, dto mwdto.Mwdto) ([]byte, error) {

	logger.LogDebug("AddScriptTask", dto.Reqw)

	sn := p.GetSnScript(dto.Reqw)

	logger.LogDebug("sn", sn)

	if sn == "" {
		return nil, ErrNotfoundSerial
	}

	err := scripter.AddToScripter(sn, dto.Reqw, nil)
	if err != nil {
		return nil, err
	}

	c := repository.GetCache()
	if c == nil {
		cfg := config.GetConfig()

		c = repository.NewCache(context.Background(), cfg)
	}

	if c == nil {
		return nil, ErrNotfoundTree
	}

	tree := c.Get(sn)
	if tree == nil {
		logger.LogDebug("mp is nil")
		return nil, ErrNotfoundTree
	}

	url := p.GetXMLValue(tree, soap.CR_URL)

	if url != "" {
		logger.LogDebug("crURL", url)
		dr := dac.NewRequest("", "", "GET", url, "")
		_, err := dr.Execute()

		if err != nil {
			logger.LogDebug("AddScriptTask", err)

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
