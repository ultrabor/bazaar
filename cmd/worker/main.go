package main

import (
	"bazaar/internal/platform/config"
	"bazaar/internal/platform/logger"
	"bazaar/internal/platform/worker"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.EnvLoad()
	log := logger.New(cfg, "bazaar worker")
	log.Info("worker runtime started")

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	poolCtx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	pool := worker.New(poolCtx, log, 3)
	defer pool.Close()
	pool.Start()

	for range 40 {
		pool.Add(func(ctx context.Context) error {
			time.Sleep(3 * time.Second)
			log.Info("added some work")
			return nil
		})
	}

	<-rootCtx.Done()

	log.Info("Shut downing workers")

	// shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancelShutdown()
	// <-shutdownCtx.Done()
}
