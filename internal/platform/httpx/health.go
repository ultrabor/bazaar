package httpx

import (
	"bazaar/internal/platform/supports/apperror"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	if err := h.sm.healthService.Ready(ctx); err != nil {
		var depErr *apperror.DependencyError
		status := http.StatusInternalServerError
		msg := "DB closed"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			msg = "deadline exceeded"
		case errors.As(err, &depErr):
			status = http.StatusServiceUnavailable
			msg = "service unavailable"
		default:
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(msg))
		h.logger.Error("ready handler", slog.Any("err", err))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("DB is OK"))
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
