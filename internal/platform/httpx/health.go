package httpx

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	if err := h.sm.healthService.Ready(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("DB closed"))
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
