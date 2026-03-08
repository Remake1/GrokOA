package main

import (
	appconfig "api/config"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	httpcontroller "api/internal/controller/http"
	geminirepository "api/internal/repository/ai/gemini"
	openairepository "api/internal/repository/ai/openai"
	healthrepository "api/internal/repository/health"
	aiservice "api/internal/service/ai"
	authservice "api/internal/service/auth"
	healthservice "api/internal/service/health"
	roomservice "api/internal/service/room"
	screenshotservice "api/internal/service/screenshot"
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

	roomManager := roomservice.NewManager(cfg.Room.GracePeriod)
	defer roomManager.Stop()

	screenshotSvc, err := screenshotservice.NewService(cfg.Screenshot.Dir)
	if err != nil {
		log.Fatalf("init screenshot service: %v", err)
	}

	if strings.TrimSpace(cfg.AI.OpenAI.APIKey) == "" {
		logger.Warn().Msg("OPENAI_API_KEY is empty, OpenAI provider is disabled")
	} else {
		openaiRepo, repoErr := openairepository.NewRepository(cfg.AI.OpenAI.APIKey)
		if repoErr != nil {
			log.Fatalf("init openai repository: %v", repoErr)
		}

		if registerErr := aiservice.GlobalRegistry.Register(aiservice.ProviderOpenAI, openaiRepo, cfg.AI.OpenAI.Models); registerErr != nil {
			log.Fatalf("register openai provider: %v", registerErr)
		}

		logger.Info().
			Str("provider", string(aiservice.ProviderOpenAI)).
			Int("models_count", len(cfg.AI.OpenAI.Models)).
			Msg("ai provider registered")
	}

	if strings.TrimSpace(cfg.AI.Gemini.APIKey) == "" {
		logger.Warn().Msg("GEMINI_API_KEY is empty, Gemini provider is disabled")
	} else {
		geminiRepo, repoErr := geminirepository.NewRepository(context.Background(), cfg.AI.Gemini.APIKey)
		if repoErr != nil {
			log.Fatalf("init gemini repository: %v", repoErr)
		}

		if registerErr := aiservice.GlobalRegistry.Register(aiservice.ProviderGemini, geminiRepo, cfg.AI.Gemini.Models); registerErr != nil {
			log.Fatalf("register gemini provider: %v", registerErr)
		}

		logger.Info().
			Str("provider", string(aiservice.ProviderGemini)).
			Int("models_count", len(cfg.AI.Gemini.Models)).
			Msg("ai provider registered")
	}

	roomHandler := httpcontroller.NewRoomHandler(roomManager, screenshotSvc, aiservice.GlobalRegistry, authSvc, logger)
	router := httpcontroller.NewRouter(healthHandler, authHandler, roomHandler, logger)

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
