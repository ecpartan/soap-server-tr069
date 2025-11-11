package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/creachadair/jrpc2"
	"github.com/ecpartan/soap-server-tr069/internal/apperror"
	"github.com/ecpartan/soap-server-tr069/internal/config"
	logger "github.com/ecpartan/soap-server-tr069/log"
	jrcp2server "github.com/ecpartan/soap-server-tr069/pkg/jrpc2"
	"github.com/ecpartan/soap-server-tr069/pkg/jrpc2/mwdto"
	repository "github.com/ecpartan/soap-server-tr069/repository/cache"
	"github.com/ecpartan/soap-server-tr069/server/handlers"
	"github.com/ecpartan/soap-server-tr069/tasks/tasker"
	"github.com/fanliao/go-promise"
	"github.com/julienschmidt/httprouter"
)

type handlerJrpc2 struct {
	Cache     *repository.Cache
	execTasks *tasker.Tasker
}

func NewHandler(Cache *repository.Cache, execTasks *tasker.Tasker) handlers.Handler {
	return &handlerJrpc2{
		Cache:     Cache,
		execTasks: execTasks,
	}
}

func (h *handlerJrpc2) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, "/front", apperror.Middleware(h.ExecFrontReq))
	router.HandlerFunc(http.MethodGet, "/frontcli", apperror.Middleware(h.ExecFrontWithoutJRPC2))
}

func RequestToFrontCli(cfg *config.Config, jsonData []byte) (string, error) {

	client := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/frontcli", cfg.Server.Host, cfg.Server.Port)

	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// Execute Frontend request
// @Summary Get Users
// @Tags Frontend
// @Success 200
// @Router / [get]
func (h *handlerJrpc2) ExecFrontReq(w http.ResponseWriter, r *http.Request) error {
	logger.LogDebug("Enter ExecFrontReq")

	msg, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	parsereq, err := jrpc2.ParseRequests(msg)
	if err != nil {
		return err
	}

	srvMethods := jrcp2server.Instance.Server.Server.ServerInfo().Methods

	logger.LogDebug("Methods", srvMethods)

	for _, req := range parsereq {
		for _, m := range srvMethods {
			if req.Method == m {
				mp := make(map[string]any)
				if err := json.Unmarshal(req.Params, &mp); err != nil {
					logger.LogDebug("Get", err)
					return err
				}
				logger.LogDebug("Script", mp)

				if script, ok := mp["Script"].(map[string]any); ok {
					logger.LogDebug("Script", script)

					task := func() (any, error) {
						var ret []byte
						ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
						defer cancel()
						err := jrcp2server.Instance.Server.Client.CallResult(ctx, req.Method, mwdto.Mwdto{script, h.execTasks.ExecTasks}, &ret)
						return ret, err
					}

					f := promise.Start(task).OnSuccess(func(result any) {
						logger.LogDebug("Success", result)
					}).OnFailure(func(v any) {
						logger.LogDebug("Failure", v)
					})
					result, err := f.Get()

					if err != nil {
						logger.LogDebug("Get", err)
						return err
					} else {
						w.Write(result.([]byte))
					}
				} else {
					return apperror.ErrInvalidType
				}
			}
		}
	}

	return nil
}

func (h *handlerJrpc2) ExecFrontWithoutJRPC2(w http.ResponseWriter, r *http.Request) error {
	logger.LogDebug("Enter ExecFrExecFrontWithoutJRPC2ontReq")
	logger.LogDebug("soapRequestBytes", h.execTasks)

	msg, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	var mp map[string]any
	err = json.Unmarshal(msg, &mp)

	if err != nil {
		return err
	}

	logger.LogDebug("Script", mp)

	if script, ok := mp["Script"].(map[string]any); ok {
		logger.LogDebug("Script", script)

		task := func() (any, error) {
			var ret []byte
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			err := jrcp2server.Instance.Server.Client.CallResult(ctx, "AddScript", mwdto.Mwdto{Reqw: script, ExecTasks: h.execTasks.ExecTasks}, &ret)
			return ret, err
		}

		f := promise.Start(task).OnSuccess(func(result any) {
			logger.LogDebug("Success", result)
		}).OnFailure(func(v any) {
			logger.LogDebug("Failure", v)
		})
		result, err := f.Get()

		if err != nil {
			logger.LogDebug("Get", err)
			return err
		} else {
			w.Write(result.([]byte))
		}
	} else {
		return apperror.ErrInvalidType
	}

	return nil
}
