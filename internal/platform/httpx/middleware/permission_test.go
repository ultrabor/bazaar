package middleware

import (
	"bazaar/internal/platform/identity"
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPermissionLogsOnlyFailuresAsError(t *testing.T) {
	tests := []struct {
		name       string
		allowed    bool
		loadErr    error
		wantStatus int
		wantError  bool
	}{
		{name: "allowed", allowed: true, wantStatus: http.StatusOK},
		{name: "denied", wantStatus: http.StatusForbidden},
		{name: "database error", loadErr: errors.New("db down"), wantStatus: http.StatusInternalServerError, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			log := slog.New(slog.NewTextHandler(&logs, nil))
			load := func(ctx context.Context, userId, code string) (bool, error) {
				return tt.allowed, tt.loadErr
			}
			next := func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }
			handler := New(log, "secret", nil, load).Permission(next, "users.view")

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = req.WithContext(identity.WithUser(req.Context(), &identity.User{Id: "user-1"}))
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)

			if res.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", res.Code, tt.wantStatus)
			}
			if got := strings.Contains(logs.String(), "level=ERROR"); got != tt.wantError {
				t.Fatalf("ERROR logged = %v, want %v; logs: %s", got, tt.wantError, logs.String())
			}
		})
	}
}
