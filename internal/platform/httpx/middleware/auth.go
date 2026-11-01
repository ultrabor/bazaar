package middleware

import (
	"bazaar/internal/platform/httpx"
	"bazaar/internal/platform/identity"
	"bazaar/internal/platform/support/apperror"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

func (m *Middleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			httpx.WriteError(w, http.StatusUnauthorized, "authorization_header_missing", "authorization header is missing")
			return
		}

		token := parts[1]

		u, err := m.loadUser(r.Context(), token)
		if err != nil {
			if errors.Is(err, apperror.ErrInvalidCred) {
				httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
				return
			}

			m.logger.Error("load authorization error", slog.Any("err", err))

			httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal error")
			return
		}

		next.ServeHTTP(w, r.WithContext(identity.WithUser(r.Context(), u)))
	})
}
