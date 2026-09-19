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
			p.worker(index)
		}(i)
	}
}

func (p *Pool) Close() {
	p.log.Info("Closing pool for workers")
	close(p.jobs)
	p.wg.Wait()
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

func (p *Pool) worker(workerId int) {

	for {
		select {
		case <-p.ctx.Done():
			return
		case job, ok := <-p.jobs:
			if !ok {
				return
			}
			p.log.Info("running new job", slog.Int("workerId", workerId))
			err := job(p.ctx)
			if err != nil {
				p.log.Error("fail worker", slog.Any("err", err), slog.Int("workerId", workerId))
				return
			}
		}
	}
}
