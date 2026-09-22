package worker

import (
	"context"
	"log/slog"
	"sync"
)

type Job func(ctx context.Context) error

type Pool struct {
	ctx context.Context
	log *slog.Logger
	wg  sync.WaitGroup

	jobs chan Job
	maxW int
}

func New(ctx context.Context, log *slog.Logger, maxW int) *Pool {
	log.Info("pool created successfully")
	pool := &Pool{
		ctx:  ctx,
		log:  log,
		jobs: make(chan Job),
		maxW: maxW,
		wg:   sync.WaitGroup{},
	}

	return pool
}

func (p *Pool) Start() {
	for i := range p.maxW {
		p.wg.Add(1)
		go func(index int) {
			defer p.wg.Done()
			p.worker(index + 1)
		}(i)
	}
}

func (p *Pool) Close() {
	p.log.Info("Closing pool for workers")
	close(p.jobs)
	p.wg.Wait()
}

func (p *Pool) Add(job Job) error {
	select {
	case <-p.ctx.Done():
		p.log.Error("context done when added job")
		return p.ctx.Err()
	case p.jobs <- job:
		p.log.Info("Added Job")
		return nil
	}
}

func (p *Pool) worker(workerId int) {
	for {
		select {
		case <-p.ctx.Done():
			return
		case job, ok := <-p.jobs:
			if !ok {
				return
			}
			p.log.Info("worker start", slog.Int("worker_id", workerId))
			err := job(p.ctx)
			if err != nil {
				p.log.Error("worker failure", slog.Int("worker_id", workerId), slog.Any("err", err))
			} else {
				p.log.Info("worker finish", slog.Int("worker_id", workerId))
			}
		}
	}
}
