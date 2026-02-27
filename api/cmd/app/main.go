package main

import (
	appconfig "api/config"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	httpcontroller "api/internal/controller/http"
	healthrepository "api/internal/repository/health"
	authservice "api/internal/service/auth"
	healthservice "api/internal/service/health"
	"api/pkg/httpserver"
	applogger "api/pkg/logger"
)

func main() {
	cfg, err := appconfig.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger, err := applogger.New(applogger.Config{
		Level:           cfg.Logging.Level,
		Format:          cfg.Logging.Format,
		TimeFieldFormat: cfg.Logging.TimeFieldFormat,
		IncludeCaller:   cfg.Logging.IncludeCaller,
	})
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}

	statusRepo := healthrepository.NewStatusRepository()
	statusService := healthservice.NewService(statusRepo)
	healthHandler := httpcontroller.NewHealthHandler(statusService)
	authSvc := authservice.NewService(cfg.Auth.AccessKey, cfg.Auth.JWTSecret, cfg.Auth.TokenTTL)
	authHandler := httpcontroller.NewAuthHandler(authSvc)
	router := httpcontroller.NewRouter(healthHandler, authHandler, logger)

	server := httpserver.New(httpserver.Config{
		Address:         cfg.HTTP.Address(),
		ReadTimeout:     cfg.HTTP.ReadTimeout,
		WriteTimeout:    cfg.HTTP.WriteTimeout,
		IdleTimeout:     cfg.HTTP.IdleTimeout,
		ShutdownTimeout: cfg.HTTP.ShutdownTimeout,
	}, router)

	serverErrCh := make(chan error, 1)
	go func() {
		logger.Info().Str("address", cfg.HTTP.Address()).Msg("http server listening")
		if serveErr := server.Start(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			serverErrCh <- serveErr
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		logger.Info().Msg("shutdown signal received")
	case serveErr := <-serverErrCh:
		logger.Error().Err(serveErr).Msg("http server failed")
		os.Exit(1)
	}

	if shutdownErr := server.Shutdown(context.Background()); shutdownErr != nil {
		logger.Error().Err(shutdownErr).Msg("graceful shutdown failed")
		os.Exit(1)
	}

	logger.Info().Msg("http server stopped")
}
