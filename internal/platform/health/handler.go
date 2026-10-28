package health

import (
	"bazaar/internal/platform/httpx"
	"bazaar/internal/platform/support/apperror"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	logger *slog.Logger
	svc    *Service
}

func NewHandler(logger *slog.Logger, svc *Service) *Handler {
	return &Handler{logger: logger, svc: svc}
}

// Routes mounts the liveness and readiness probes on r.
func (h *Handler) Routes(r chi.Router) {
	r.Get("/healthz", h.Health)
	r.Get("/readyz", h.Ready)
}

// @Summary Readiness probe
// @Description Checks database connectivity
// @Tags Health
// @Produce plain
// @Success 200 {string} string "DB is OK"
// @Failure 500 {object} httpx.Error "Internal error"
// @Failure 503 {object} httpx.Error "Service unavailable"
// @Router /readyz [get]
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	if err := h.svc.Ready(ctx); err != nil {
		var depErr *apperror.DependencyError
		status := http.StatusInternalServerError
		code := "internal_error"
		msg := "internal error"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			code = "service_unavailable"
			msg = "service unavailable"
		case errors.As(err, &depErr):
			status = http.StatusServiceUnavailable
			code = "service_unavailable"
			msg = "service unavailable"
		default:
		}
		httpx.WriteError(w, status, code, msg)
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
