package httpx

import (
	"bazaar/internal/platform/identity"
	"context"
	"encoding/json"
	"net/http"
)

// Error is the body of every error response.
type Error struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Error{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

// WriteJSON writes value as a JSON body with the given status.
func WriteJSON(w http.ResponseWriter, status int, value any) {
	payload, err := json.Marshal(value)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(payload)
}

// CurrentUser returns the authenticated user from ctx. When there is none
// it writes 401 and returns false, so the handler only has to return.
func CurrentUser(w http.ResponseWriter, ctx context.Context) (*identity.User, bool) {
	u, ok := identity.FromContext(ctx)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
	}
	return u, ok
}
