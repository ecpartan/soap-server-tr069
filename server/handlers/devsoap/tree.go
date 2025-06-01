package devsoap

import (
	"fmt"
	"io"
	"net/http"
	"os"

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
	return &handlerCR{
		Cache: Cache,
	}
}

func (h *handlerGetTree) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, "/GetTree", apperror.Middleware(h.GetTree))
}
func (h *handlerGetTree) GetTree(w http.ResponseWriter, r *http.Request) error {
	soapRequestBytes, _ := io.ReadAll(r.Body)
	logger.LogDebug("soapRequestBytes", string(soapRequestBytes))

	dat, _ := os.ReadFile("notify.xml")

	w.Header().Set("Content-Length", fmt.Sprint(len(dat)))
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")

	w.Write(dat)
	return nil
}
