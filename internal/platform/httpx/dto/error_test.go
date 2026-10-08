package dto

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()

	WriteError(w, http.StatusBadRequest, "invalid_pagination", "limit must be between 1 and 100")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if got, want := w.Header().Get("Content-Type"), "application/json"; got != want {
		t.Fatalf("Content-Type = %q, want %q", got, want)
	}
	if got, want := w.Body.String(), "{\"error\":{\"code\":\"invalid_pagination\",\"message\":\"limit must be between 1 and 100\"}}\n"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}
