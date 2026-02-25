package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	authservice "api/internal/service/auth"
)

type authService interface {
	Authorize(ctx context.Context, key string) (string, error)
}

type AuthHandler struct {
	service authService
}

func NewAuthHandler(service authService) *AuthHandler {
	return &AuthHandler{service: service}
}

type authRequest struct {
	Key string `json:"key"`
}

func (h *AuthHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var request authRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})

		return
	}

	token, err := h.service.Authorize(r.Context(), request.Key)
	if err != nil {
		switch {
		case errors.Is(err, authservice.ErrEmptyKey):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": authservice.ErrEmptyKey.Error()})
		case errors.Is(err, authservice.ErrWrongKey):
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": authservice.ErrWrongKey.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		}

		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}
