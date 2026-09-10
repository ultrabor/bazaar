package main

import (
	"bazaar/internal/platform/config"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

func main() {
	cfg := config.EnvLoad()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// db := database.New(ctx, cfg, logger)

	router := chi.NewRouter()

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	err := http.ListenAndServe(cfg.HttpHost+cfg.HttpPort, router)
	if err != nil {
		logger.Error("Error on listen", "err", err)
	}
}
