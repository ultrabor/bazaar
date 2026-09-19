package worker

import (
	"context"
	"log/slog"
)

type Job func(ctx context.Context) error

type Pool struct {
	workers int
	jobs    chan Job
	log     *slog.Logger
}

func New(ctx context.Context, maxWorkers int, log *slog.Logger) *Pool {

	return &Pool{workers: 0, jobs: make(chan Job, maxWorkers), log: log}
}

func (p *Pool) Close() {
	p.log.Info("Closing pool for workers")
	close(p.jobs)
}

func (p *Pool) Add(job Job) {
	p.jobs <- job
	p.workers++
	p.log.Info("Added Job")
}
