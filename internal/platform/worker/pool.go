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
	mux sync.Mutex
	wg  sync.WaitGroup

	workers int
	jobs    chan Job
	maxW    int
}

func New(ctx context.Context, maxWorkers int, log *slog.Logger, maxW int) *Pool {
	log.Info("pool created successfully")
	return &Pool{
		ctx:     ctx,
		log:     log,
		workers: maxWorkers,
		jobs:    make(chan Job),
		maxW:    maxW,
		wg:      sync.WaitGroup{},
		mux:     sync.Mutex{},
	}

}

func (p *Pool) Close() {
	p.log.Info("Closing pool for workers")
	close(p.jobs)
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

func (p *Pool) worker() {

	for job := range p.jobs {
		select {
		case <-p.ctx.Done():
			p.log.Error("context done")
			return
		default:
			p.wg.Add(1)
			p.mux.Lock()
			p.workers--
			p.mux.Unlock()
			go func(j Job) {
				defer p.wg.Done()

				err := j
				if err != nil {
					p.log.Error("job fail", slog.Any("err", err))
					return
				}
				p.mux.Lock()
				p.workers++
				p.mux.Unlock()
			}(job)
		}
	}

	p.wg.Wait()

	// go func() {
	// 	select {
	// 	case <-p.ctx.Done():
	// 		p.log.Error("context done")
	// 		return
	// 	case job := <-p.jobs:
	// 		for {
	// 			go func(ctx context.Context) {
	// 				err := job(p.ctx)
	// 				if err != nil {
	// 					p.log.Info("")
	// 				}
	// 			}(p.ctx)
	// 		}
	// 	}
	// }()
}
