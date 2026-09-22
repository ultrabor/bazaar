package worker_test

import (
	"io"
	"log/slog"
	"testing"
)

func testLogger() *slog.Logger {
	return slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)
}

func TestMaxWorker(t *testing.T) {

}

func TestCancellation(t *testing.T) {

}

func TestCloseWaitWorker(t *testing.T) {

}

func TestErrOfOneJobKillWorker(t *testing.T) {

}
