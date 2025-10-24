package monitoring

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

var (
	HBURL = "/heartbeat"
)

type Handler struct {
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, HBURL, h.Heartbeat)
}

// Heartbeat
// @Summary Heartbeat metrics
// @Tags metrics
// @Success 204
// @Router /heartbeat [get]
func (h *Handler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
