package middleware

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/creachadair/jrpc2"
	"github.com/ecpartan/soap-server-tr069/internal/apperror"
	jrcp2server "github.com/ecpartan/soap-server-tr069/jrpc2"
	logger "github.com/ecpartan/soap-server-tr069/log"
	repository "github.com/ecpartan/soap-server-tr069/repository/cache"
	"github.com/ecpartan/soap-server-tr069/server/handlers"
	"github.com/fanliao/go-promise"
	"github.com/julienschmidt/httprouter"
)

type handlerJrpc2 struct {
	Cache *repository.Cache
}

func NewHandler(Cache *repository.Cache) handlers.Handler {
	return &handlerJrpc2{
		Cache: Cache,
	}
}

func (h *handlerJrpc2) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, "/front", apperror.Middleware(h.ExecFrontReq))
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
						err := jrcp2server.Instance.Server.Client.CallResult(ctx, req.Method, script, &ret)
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
