package location

import (
	"bazaar/internal/platform/httpx/httpxtest"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestNonUUIDLocationIDReturns400(t *testing.T) {
	h := NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), NewService(nil))
	r := chi.NewRouter()
	r.Get("/location/{id}", h.GetLocationById)
	r.Put("/location/{id}", h.UpdateLocation)
	r.Patch("/location/{id}/archive", h.ArchiveLocation)

	tests := []struct {
		method, path, body string
	}{
		{http.MethodGet, "/location/abc", ""},
		{http.MethodPut, "/location/abc", `{"name":"Main","address":"Address","type":"store"}`},
		{http.MethodPatch, "/location/abc/archive", ""},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			res := httpxtest.Serve(t, r, tt.method, tt.path, tt.body)

			if res.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400: %s", res.Code, res.Body)
			}
			if code := httpxtest.ErrorCode(t, res); code != "invalid_location_id" {
				t.Errorf("code = %q, want invalid_location_id", code)
			}
		})
	}
}
