package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

func NewRouter(healthHandler *HealthHandler, authHandler *AuthHandler, roomHandler *RoomHandler, logger zerolog.Logger) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(RequestLogger(logger))
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(30 * time.Second))

	router.Route("/api", func(r chi.Router) {
		r.Get("/live", healthHandler.Live)
		r.Get("/ready", healthHandler.Ready)
		r.Post("/auth", authHandler.Authorize)

		r.Get("/ws/client", roomHandler.HandleWebClient)
		r.Get("/ws/desktop", roomHandler.HandleDesktop)
	})

	return router
}
