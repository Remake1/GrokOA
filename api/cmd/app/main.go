package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	appconfig "api/internal/config"
	httpcontroller "api/internal/controller/http"
	healthrepository "api/internal/repository/health"
	healthservice "api/internal/service/health"
	"api/pkg/httpserver"
)

func main() {
	cfg, err := appconfig.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	statusRepo := healthrepository.NewStatusRepository()
	statusService := healthservice.NewService(statusRepo)
	healthHandler := httpcontroller.NewHealthHandler(statusService)
	router := httpcontroller.NewRouter(healthHandler)

	server := httpserver.New(httpserver.Config{
		Address:         cfg.HTTP.Address(),
		ReadTimeout:     cfg.HTTP.ReadTimeout,
		WriteTimeout:    cfg.HTTP.WriteTimeout,
		IdleTimeout:     cfg.HTTP.IdleTimeout,
		ShutdownTimeout: cfg.HTTP.ShutdownTimeout,
	}, router)

	serverErrCh := make(chan error, 1)
	go func() {
		log.Printf("http server listening on %s", cfg.HTTP.Address())
		if serveErr := server.Start(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			serverErrCh <- serveErr
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		log.Println("shutdown signal received")
	case serveErr := <-serverErrCh:
		log.Printf("http server failed: %v", serveErr)
		os.Exit(1)
	}

	if shutdownErr := server.Shutdown(context.Background()); shutdownErr != nil {
		log.Printf("graceful shutdown failed: %v", shutdownErr)
		os.Exit(1)
	}

	log.Println("http server stopped")
}
