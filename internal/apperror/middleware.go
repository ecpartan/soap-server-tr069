package apperror

import (
	"errors"
	"net/http"

	logger "github.com/ecpartan/soap-server-tr069/log"
)

type AppHandler func(http.ResponseWriter, *http.Request) error

func Middleware(h AppHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.LogDebug("Enter middleware", r.URL.Path)

		var appErr *AppError
		err := h(w, r)

		if err != nil {
			if errors.As(err, &appErr) {
				if errors.Is(err, ErrNotFound) {
					w.WriteHeader(http.StatusNotFound)
					w.Write(ErrNotFound.Marshal())
					return
				} else if errors.Is(err, ErrAlreadyExist) {
					w.WriteHeader(http.StatusConflict)
					w.Write(ErrNotFound.Marshal())
					return
				} else if errors.Is(err, ErrUnAuthorized) {
					w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
					w.WriteHeader(http.StatusUnauthorized)
					w.Write(ErrUnAuthorized.Marshal())
					return
				}
				err := err.(*AppError)
				w.WriteHeader(http.StatusBadRequest)
				w.Write(err.Marshal())
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(systemError(err.Error()).Marshal())
		}
	}
}
