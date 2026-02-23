package http

import (
	"context"
	"encoding/json"
	"net/http"
)

type healthService interface {
	Live(ctx context.Context) error
	Ready(ctx context.Context) error
}

type HealthHandler struct {
	service healthService
}

func NewHealthHandler(service healthService) *HealthHandler {
	return &HealthHandler{service: service}
}

func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Live(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})

		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Ready(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})

		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
