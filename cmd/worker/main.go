package main

import (
	"bazaar/internal/platform/config"
	"bazaar/internal/platform/logger"
	"bazaar/internal/platform/worker"
	"context"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.EnvLoad()
	log := logger.New(cfg, "bazaar worker")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool := worker.New(3, log)
	defer pool.Close()

	go func(p worker.Pool) {
		for {
			select {
			case <-ctx.Done:

			}
		}
	}(pool)
}
