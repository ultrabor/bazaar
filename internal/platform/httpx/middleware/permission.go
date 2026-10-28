package middleware

import (
	"bazaar/internal/platform/httpx"
	"bazaar/internal/platform/identity"
	"log/slog"
	"net/http"
)

func (m *Middleware) Permission(next http.HandlerFunc, code string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if ctx == nil {
			httpx.WriteError(w, http.StatusBadRequest, "forbidden", "bad request")
			return
		}

		user, ok := identity.FromContext(ctx)
		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
			return
		}

		allowed, err := m.hasPermission(ctx, user.Id, code)
		if err != nil {
			m.logger.Error("permission check failed", slog.String("permission", code), slog.Any("err", err))
			httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal error")
			return
		}
		if !allowed {
			m.logger.Info("permission denied", slog.String("user_id", user.Id), slog.String("permission", code))
			httpx.WriteError(w, http.StatusForbidden, "do_not_have_enough_permission", "dont have enough permission")
			return
		}

		next.ServeHTTP(w, r)

	})
}
