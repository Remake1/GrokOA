package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"api/internal/dto"
	authservice "api/internal/service/auth"

	"github.com/go-chi/httprate"
)

type authService interface {
	Authorize(ctx context.Context, key string) (string, error)
}

type AuthHandler struct {
	service            authService
	failedLoginLimiter *httprate.RateLimiter
}

func NewAuthHandler(service authService, failedAttemptLimit int, failedAttemptWindow time.Duration) *AuthHandler {
	limiter := httprate.NewRateLimiter(
		failedAttemptLimit,
		failedAttemptWindow,
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "too many failed login attempts"})
		}),
	)

	return &AuthHandler{
		service:            service,
		failedLoginLimiter: limiter,
	}
}

func (h *AuthHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var request dto.AuthRequest
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
			if h.respondOnFailedLoginLimit(w, r) {
				return
			}

			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": authservice.ErrWrongKey.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		}

		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *AuthHandler) respondOnFailedLoginLimit(w http.ResponseWriter, r *http.Request) bool {
	if h.failedLoginLimiter == nil {
		return false
	}

	key, err := httprate.KeyByIP(r)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return true
	}

	return h.failedLoginLimiter.RespondOnLimit(w, r, key)
}
