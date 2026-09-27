package httpx

import (
	"bazaar/internal/modules/auth"
	"context"
	"encoding/json"
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
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("service unavailable"))
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
	_, _ = w.Write(re)
}
