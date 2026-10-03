package middleware

import (
	"bazaar/internal/modules/auth"
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
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		userId, err := validator.ValidateToken(token, []byte(m.secretKey))
		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		u, err := m.loadUser(r.Context(), userId)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidCred) {
				http.Error(w, "Unathorized", http.StatusUnauthorized)
				return
			}

			m.logger.Error("load authorization error", slog.Any("err", err))

			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), currentUserKey, u)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
