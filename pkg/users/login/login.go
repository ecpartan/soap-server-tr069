package login

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ecpartan/soap-server-tr069/internal/apperror"
	logger "github.com/ecpartan/soap-server-tr069/log"
	usecase_user "github.com/ecpartan/soap-server-tr069/repository/db/domain/usecase/user"
	"github.com/ecpartan/soap-server-tr069/server/handlers"
	"github.com/julienschmidt/httprouter"
)

type handlerLogin struct {
	service *usecase_user.Service
}

func NewHandler(service *usecase_user.Service) handlers.Handler {
	return &handlerLogin{
		service: service,
	}
}

type login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *handlerLogin) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, "/Login", apperror.Middleware(h.Login))
}

// Login
// @Summary Login and authenticate user
// @Tags login
// @Success 200 {object} loginResponse
// @Router  /login [post]
func (h *handlerLogin) Login(w http.ResponseWriter, r *http.Request) error {
	logger.LogDebug("Enter Login")
	auth, err := io.ReadAll(r.Body)

	if err != nil {
		return fmt.Errorf("could not read POST: %v", err)
	}

	login := login{}
	err = json.Unmarshal(auth, &login)
	if err != nil {
		return fmt.Errorf("could not unmarshal POST: %v", err)
	}

	user, err := h.service.GetUserbyLogin(login.Username)
	if err != nil {
		return fmt.Errorf("not found user")
	}

	logger.LogDebug("users", user)

	if user.Password != login.Password {
		return fmt.Errorf("password is not corrected")
	}

	jwtsecret := getJWTsecret()
	/*
		id, err := strconv(user.ID)
		if err != nil {
			return fmt.Errorf("invalid user ID: %v", err)
		}*/

	t, err := generateJWT(user.ID, jwtsecret)
	if err != nil {
		return fmt.Errorf("could not generate JWT: %v", err)
	}

	dat, err := json.Marshal(loginResponse{Token: t})

	if err != nil {
		return fmt.Errorf("internal error")
	}

	w.Write(dat)

	return nil
}
