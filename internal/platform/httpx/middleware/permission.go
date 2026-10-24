package middleware

import (
	"bazaar/internal/modules/permission"
	"bazaar/internal/modules/user"
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

		user, ok := ctx.Value(currentUserKey).(*user.User)
		if !ok {
			dto.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
			return
		}

		_, err := m.loadPermision(ctx, user.Id, code)
		if err != nil {
			status := http.StatusInternalServerError
			code := "internal_error"
			msg := "internal error"
			if errors.Is(err, permission.ErrNotFound) {
				status = http.StatusForbidden
				code = "do_not_have_enough_permission"
				msg = "dont have enough permission"
			}
			m.logger.Error("service unavailable", slog.Any("err", err))
			dto.WriteError(w, status, code, msg)
			return
		}

		next.ServeHTTP(w, r)

	})
}
