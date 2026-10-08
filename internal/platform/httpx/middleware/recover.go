package middleware

import (
	"bazaar/internal/platform/httpx/dto"
	"log/slog"
	"net/http"
	"runtime/debug"
)

func (m *Middleware) Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				if err == http.ErrAbortHandler {
					panic(err)
				}

				m.logger.Error("Recover", slog.Any("error", err), slog.String("stack", string(debug.Stack())))

				w.Header().Set("Connection", "close")
				dto.WriteError(w, http.StatusInternalServerError, "internal_error", "internal error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
