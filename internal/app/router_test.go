package app

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"bazaar/internal/platform/health"
	"bazaar/internal/platform/httpx/middleware"
)

func TestRoutesRedirectTrailingSlash(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := NewRouter(middleware.New(log, "secret", nil, nil), health.NewHandler(log, nil))

	for _, path := range []string{"/user/", "/location/"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != http.StatusMovedPermanently {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusMovedPermanently)
			}
			if location := response.Header().Get("Location"); location != path[:len(path)-1] {
				t.Fatalf("Location = %q, want %q", location, path[:len(path)-1])
			}
		})
	}
}
