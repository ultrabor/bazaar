package main

import (
	"bazaar/internal/platform/config"
	"bazaar/internal/platform/database"
	"bazaar/pkg/goose"
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.EnvLoad()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx, cancel := context.WithTimeout(
		context.Background(),
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
		logger.Error("goose init", "err", err)
		return
	}

	defer mig.Close()

	err = mig.Up()
	if err != nil {
		logger.Error("migration up", "err", err)
	}

	defer db.Close()

	router := chi.NewRouter()

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})

	err = http.ListenAndServe(cfg.HttpHost+":"+cfg.HttpPort, router)
	if err != nil {
		logger.Error("Error on listen", "err", err)
	}
}
