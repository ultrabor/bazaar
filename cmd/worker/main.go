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

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	poolCtx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	pool := worker.New(poolCtx, 3, log)
	defer pool.Close()

	<-rootCtx.Done()

	log.Info("Shut downing workers")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	<-shutdownCtx.Done()
}
