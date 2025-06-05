package devsoap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ecpartan/soap-server-tr069/internal/apperror"
	logger "github.com/ecpartan/soap-server-tr069/log"
	repository "github.com/ecpartan/soap-server-tr069/repository/cache"
	"github.com/ecpartan/soap-server-tr069/server/handlers"
	"github.com/julienschmidt/httprouter"
)

type handlerGetTree struct {
	Cache *repository.Cache
}

func NewHandlerGetTree(Cache *repository.Cache) handlers.Handler {
	return &handlerGetTree{
		Cache: Cache,
	}
}
func (h *handlerGetTree) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, "/GetTree/:sn", apperror.Middleware(h.GetTree))
}
func (h *handlerGetTree) GetTree(w http.ResponseWriter, r *http.Request) error {
	logger.LogDebug("Enter GetTree")
	soapRequestBytes, _ := io.ReadAll(r.Body)
	logger.LogDebug("soapRequestBytes", string(soapRequestBytes))
	sn := httprouter.ParamsFromContext(r.Context()).ByName("sn")

	if sn == "" {
		return fmt.Errorf("not found sn")
	}

	tree := h.Cache.Get(sn)
	if tree == nil {
		return fmt.Errorf("not found tree")
	}
	dat, err := json.Marshal(tree)
	if err != nil {
		return fmt.Errorf("not found tree")
	}
	w.Header().Set("Content-Length", fmt.Sprint(len(dat)))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", r.RemoteAddr)

	w.Write(dat)

	return nil
}
