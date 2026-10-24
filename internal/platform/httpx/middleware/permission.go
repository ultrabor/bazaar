package middleware

import (
	"bazaar/internal/modules/user"
	"bazaar/internal/platform/httpx/dto"
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
			dto.WriteError(w, http.StatusMethodNotAllowed, "not_enough_permession", "not enough permession")
			return
		}

		next.ServeHTTP(w, r)

	})
}
