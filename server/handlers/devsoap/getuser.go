package devsoap

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ecpartan/soap-server-tr069/internal/apperror"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/pkg/users/login"
	usecase_user "github.com/ecpartan/soap-server-tr069/repository/db/domain/usecase/user"
	"github.com/ecpartan/soap-server-tr069/server/handlers"
	"github.com/julienschmidt/httprouter"
)

type handlerGetUsers struct {
	service *usecase_user.Service
}

type Middleware struct {
	next http.Handler
}

func (m Middleware) Wrap(handler http.Handler) http.Handler {
	m.next = handler
	return m.next
}

func NewHandlerGetUsers(service *usecase_user.Service) handlers.Handler {
	return &handlerGetUsers{
		service: service,
	}
}
func (h *handlerGetUsers) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, "/GetUsers", apperror.Middleware(login.AuthMiddleware(h.GetUsers)))
}

// Get Users info
// @Summary Get Users
// @Tags Frontend
// @Success 200
// @Router /GetUsers [get]
func (h *handlerGetUsers) GetUsers(w http.ResponseWriter, r *http.Request) error {
	logger.LogDebug("Enter GetUsers")

	users := h.service.GetAll()

	logger.LogDebug("users", users)
	dat, err := json.Marshal(users)
	if err != nil {
		return fmt.Errorf("not found tree")
	}
	logger.LogDebug("users", string(dat))

	w.Header().Set("Content-Length", fmt.Sprint(len(dat)))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", r.RemoteAddr)

	w.Write(dat)

	return nil
}
