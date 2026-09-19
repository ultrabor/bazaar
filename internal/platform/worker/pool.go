package worker

import (
	"context"
	"log/slog"
)

type Job func(ctx context.Context) error

type Pool struct {
	ctx     context.Context
	workers int
	jobs    chan Job
	log     *slog.Logger
}

func New(ctx context.Context, maxWorkers int, log *slog.Logger) *Pool {

	return &Pool{ctx: ctx, workers: 0, jobs: make(chan Job, maxWorkers), log: log}
}

func (p *Pool) Close() {
	p.log.Info("Closing pool for workers")
	close(p.jobs)
	p.ctx.Done()
}

func (p *Pool) Add(job Job) {
	select {
	case <-p.ctx.Done():
		p.log.Error("context done when added job")
		return
	case p.jobs <- job:
		p.log.Info("Added Job")
	}
}
