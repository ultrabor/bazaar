package httpx

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"bazaar/internal/platform/httpx/middleware"
	"github.com/go-chi/chi/v5"
)

func TestRoutesRedirectTrailingSlash(t *testing.T) {
	router := chi.NewRouter()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := New(log, router, &ServiceManager{})
	handler.Routes(middleware.New(log, "secret", nil))

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
