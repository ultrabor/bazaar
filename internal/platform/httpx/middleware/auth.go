package middleware

import (
	"bazaar/internal/modules/auth"
	"bazaar/internal/platform/httpx/dto"
	"bazaar/internal/platform/support/validator"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

func (m *Middleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			dto.WriteError(w, http.StatusUnauthorized, "authorization_header_missing", "authorization header is missing")
			return
		}

		token := parts[1]

		userId, err := validator.ValidateToken(token, []byte(m.secretKey))
		if err != nil {
			dto.WriteError(w, http.StatusUnauthorized, "invalid_token", "invalid token")
			return
		}

		u, err := m.loadUser(r.Context(), userId)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidCred) {
				dto.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
				return
			}

			m.logger.Error("load authorization error", slog.Any("err", err))

			dto.WriteError(w, http.StatusInternalServerError, "internal_error", "internal error")
			return
		}

		ctx := context.WithValue(r.Context(), currentUserKey, u)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
