package httpx

import (
	"bazaar/internal/platform/httpx/dto"
	"bazaar/internal/platform/support/apperror"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

// @Summary Readiness probe
// @Description Checks database connectivity
// @Tags Health
// @Produce plain
// @Success 200 {string} string "DB is OK"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /readyz [get]
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	if err := h.sm.healthService.Ready(ctx); err != nil {
		var depErr *apperror.DependencyError
		status := http.StatusInternalServerError
		code := "internal_error"
		msg := "internal error"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			code = "deadline_exceeded"
			msg = "deadline exceeded"
		case errors.As(err, &depErr):
			status = http.StatusServiceUnavailable
			code = "service_unavailable"
			msg = "service unavailable"
		default:
		}
		dto.WriteError(w, status, code, msg)
		h.logger.Error("ready handler", slog.Any("err", err))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("DB is OK"))
}

// @Summary Liveness probe
// @Tags Health
// @Produce plain
// @Success 200 {string} string "OK"
// @Router /healthz [get]
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
