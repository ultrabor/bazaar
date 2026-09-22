package worker_test

import (
	"bazaar/internal/platform/worker"
	"context"
	"errors"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"
)

func testLogger() *slog.Logger {
	return slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)
}

func TestMaxWorker(t *testing.T) {

	const (
		maxWorker = 3
		jobCount  = 6
	)

	pool := worker.New(t.Context(), testLogger(), maxWorker)
	pool.Start()

	var running atomic.Int64
	var maxRunning atomic.Int64

	release := make(chan struct{})
	started := make(chan struct{}, jobCount)

	addDone := make(chan struct{})

	go func() {
		for range jobCount {
			err := pool.Add(func(ctx context.Context) error {
				current := running.Add(1)
				defer running.Add(-1)

				for {
					max := maxRunning.Load()

					if max >= current {
						break
					}

					if maxRunning.CompareAndSwap(max, current) {
						break
					}

				}
				started <- struct{}{}

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-release:
				}

				return nil
			})
			if err != nil {
				t.Errorf("failed to add work err: %v", err)
			}
		}
		close(addDone)
	}()

	for range maxWorker {
		<-started
	}

	close(release)

	<-addDone

	pool.Close()

	max := maxRunning.Load()
	if max != maxWorker {
		t.Errorf("max running jobs: got: %d, want: %d", maxRunning.Load(), maxWorker)
	}
}

func TestCancellation(t *testing.T) {
	const (
		maxWorker = 1
	)

	ctx, cancel := context.WithCancel(t.Context())

	pool := worker.New(ctx, testLogger(), maxWorker)
	pool.Start()

	started := make(chan struct{})
	errChan := make(chan error)

	err := pool.Add(func(ctxJob context.Context) error {
		started <- struct{}{}

		<-ctxJob.Done()
		errChan <- ctxJob.Err()
		return ctxJob.Err()
	})

	if err != nil {
		t.Fatalf("failed to add err: %v", err)
	}

	<-started
	cancel()

	err = <-errChan
	if !errors.Is(err, context.Canceled) {
		t.Error("doesn't get context cancellation")
	}

	pool.Close()
}

func TestCloseWaitWorker(t *testing.T) {
	const (
		maxWorker = 1
	)

	pool := worker.New(context.Background(), testLogger(), maxWorker)
	pool.Start()

	started := make(chan struct{})
	release := make(chan struct{})

	err := pool.Add(func(ctx context.Context) error {
		close(started)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-release:
			return nil
		}
	})

	if err != nil {
		t.Fatalf("pool.Add failed to add err: %v", err)
	}

	closeDone := make(chan struct{})

	<-started

	go func() {
		pool.Close()
		close(closeDone)
	}()

	select {
	case <-closeDone:
		close(release)
		t.Fatal("pool.Close returned before active job finished")
	case <-time.After(50 * time.Millisecond):
		close(release)
		select {
		case <-time.After(time.Second):
			t.Fatal("pool.Close did not return after job finished")
		case <-closeDone:
		}
	}
}

func TestJobErrorDoesNotStopWorker(t *testing.T) {
	const (
		maxWorker = 1
	)

	ctx, cancel := context.WithCancel(t.Context())

	pool := worker.New(ctx, testLogger(), maxWorker)
	pool.Start()

	err := pool.Add(func(ctx context.Context) error {
		return errors.New("some errors")
	})

	if err != nil {
		t.Fatalf("pool.Add failed to add err: %v", err)
	}

	liveWorker := make(chan struct{})

	addRes := make(chan error, 1)
	go func() {
		addRes <- pool.Add(func(ctx context.Context) error {
			close(liveWorker)
			return nil
		})

	}()

	select {
	case <-time.After(time.Second):
		t.Error("pool.worker is killed by error")
		cancel()
		<-addRes
	case <-liveWorker:
		if addErr := <-addRes; addErr != nil {
			t.Fatalf("pool.Add failed to add err: %v", addErr)
		}
	}

	cancel()
	pool.Close()
}
