package main

import (
	"bazaar/internal/app/health"
	"bazaar/internal/platform/config"
	"bazaar/internal/platform/database"
	"bazaar/internal/platform/httpx"
	"bazaar/internal/platform/httpx/middleware"
	"bazaar/internal/platform/logger"
	"bazaar/internal/platform/support/goose"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.EnvLoad()

	// logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger := slog.New(logger.NewHandler(
		&slog.HandlerOptions{
			Level: logger.LogLevel(cfg.LogLevel),
		},
	).WithAttrs(
		[]slog.Attr{
			slog.String("service", "bazaar api"),
			slog.String("env", cfg.AppEnv),
		},
	))

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ctx, cancel := context.WithTimeout(
		rootCtx,
		5*time.Second,
	)
	defer cancel()

	db, err := database.New(ctx, cfg, logger)

	if err != nil {
		logger.Error("failed to connect db", "error", err)
		return
	}

	mig, err := goose.New(cfg)
	if err != nil {
		logger.Error("goose init", slog.Any("err", err))
		return
	}

	defer mig.Close()

	err = mig.Up(ctx)
	if err != nil {
		logger.Error("migration up", slog.Any("err", err))
		return
	}

	defer db.Close()

	router := chi.NewRouter()

	handler := httpx.New(logger, router, httpx.NewServiceManager(
		health.New(db),
	))

	handler.Routes(middleware.New(logger))

	server := http.Server{
		Addr:              cfg.HttpHost + ":" + cfg.HttpPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("Starting HTTP server",
			slog.String("host", cfg.HttpHost),
			slog.String("port", cfg.HttpPort),
		)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server forced to shutdown", slog.Any("err", err))
			panic(err)
		}
	}()

	<-rootCtx.Done()
	logger.Info("Shut downing ...")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown", slog.Any("err", err))
	}

}
