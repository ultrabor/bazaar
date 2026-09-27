package httpx

import (
	"bazaar/internal/modules/auth"
	"bazaar/internal/platform/support/apperror"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

func (h *Handler) RegisterOwner(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var rq auth.RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&rq)

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte("invalid input"))
		h.logger.Error("invalid input", slog.Any("err", err))
		return
	}

	res, err := h.sm.authService.RegisterOwner(ctx, rq)

	if err != nil {
		var depErr *apperror.DependencyError
		status := http.StatusInternalServerError
		msg := "DB closed"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			msg = "deadline exceeded"
		case errors.As(err, &depErr):
			status = http.StatusBadRequest
			msg = "not valid phone number"
		}

		w.WriteHeader(status)
		_, _ = w.Write([]byte(msg))
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	re, err := json.Marshal(res)

	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("service unavailable"))
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	_, _ = w.Write(re)
}
