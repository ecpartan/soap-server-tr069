package login

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/ecpartan/soap-server-tr069/internal/apperror"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/repository/db"
	"github.com/ecpartan/soap-server-tr069/server/handlers"
	"github.com/julienschmidt/httprouter"
)

type handlerLogin struct {
	db *db.Service
}

func NewHandler(db *db.Service) handlers.Handler {
	return &handlerLogin{
		db: db,
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

	users, err := h.db.GetUser(login.Username)
	if err != nil {
		return fmt.Errorf("not found user")
	}

	logger.LogDebug("users", users)

	if users.Password != login.Password {
		return fmt.Errorf("password is not corrected")
	}

	jwtsecret := getJWTsecret()

	id, err := strconv.Atoi(users.Id)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	t, err := generateJWT(id, jwtsecret)
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
