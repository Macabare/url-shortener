package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/Macabare/url-shortener/internal/config"
	"github.com/Macabare/url-shortener/internal/http-server/handlers/health"
	"github.com/Macabare/url-shortener/internal/http-server/handlers/url/get"
	"github.com/Macabare/url-shortener/internal/http-server/handlers/url/save"
	"github.com/Macabare/url-shortener/internal/http-server/handlers/url/stat"

	mwLogger "github.com/Macabare/url-shortener/internal/http-server/middleware/logger"
	sl "github.com/Macabare/url-shortener/internal/lib/logger/sl"
	"github.com/Macabare/url-shortener/internal/storage/sqlite"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found")
	}
	cfg := config.MustLoad()

	logger := setupLogger(cfg.Env)

	logger.Info("starting url-shortener...", slog.String("env", cfg.Env))

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		logger.Error("Failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(logger))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Get("/health", health.New(logger))
	router.Post("/shorten", save.New(logger, storage))
	router.Get("/{shortCode}", get.New(logger, storage))
	router.Get("/{shortCode}/stat", stat.New(logger, storage))

	logger.Info("starting http server...", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HttpServer.Timeout,
		WriteTimeout: cfg.HttpServer.Timeout,
		IdleTimeout:  cfg.HttpServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		logger.Error("failed to start server")
		os.Exit(1)
	}

	logger.Error("server stopped")
}

const (
	envLocal = "dev"
	envProd  = "prod"
)

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	switch env {
	case envLocal:
		logger = slog.New(
			slog.NewTextHandler(
				os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		logger = slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return logger
}
