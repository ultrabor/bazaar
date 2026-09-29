package middleware

import (
	"bazaar/internal/platform/support/validator"
	"net/http"
	"strings"
)

func (m *Middleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == "" {
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		userId, err := validator.ValidateToken(token, []byte(m.secretKey))
		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		r.Header.Set("X-User-ID", userId)

		next.ServeHTTP(w, r)
	})
}
