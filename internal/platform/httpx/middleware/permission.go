package middleware

import (
	"bazaar/internal/domain"
	"bazaar/internal/platform/httpx/dto"
	"errors"
	"log/slog"
	"net/http"
)

func (m *Middleware) Permission(next http.HandlerFunc, code string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if ctx == nil {
			dto.WriteError(w, http.StatusBadRequest, "forbidden", "bad request")
			return
		}

		user, ok := CurrentUser(ctx)
		if !ok {
			dto.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
			return
		}

		_, err := m.loadPermision(ctx, user.Id, code)
		if err != nil {
			if errors.Is(err, domain.ErrPermissionNotFound) {
				m.logger.Info("permission denied", slog.String("user_id", user.Id), slog.String("permission", code))
				dto.WriteError(w, http.StatusForbidden, "do_not_have_enough_permission", "dont have enough permission")
				return
			}
			m.logger.Error("permission check failed", slog.String("permission", code), slog.Any("err", err))
			dto.WriteError(w, http.StatusInternalServerError, "internal_error", "internal error")
			return
		}

		next.ServeHTTP(w, r)

	})
}
