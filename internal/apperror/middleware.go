package apperror

import (
	"errors"
	"net/http"

	logger "github.com/ecpartan/soap-server-tr069/log"
)

type appHandler func(http.ResponseWriter, *http.Request) error

func Middleware(h appHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var appErr *AppError
		err := h(w, r)
		logger.LogDebug("err", err)
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
				}
				err := err.(*AppError)
				w.WriteHeader(http.StatusBadRequest)
				w.Write(err.Marshal())
				return
			}
			w.WriteHeader(http.StatusNoContent)
			w.Write(systemError(err.Error()).Marshal())
		}
	}
}
