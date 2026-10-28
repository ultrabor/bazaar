// Package httpxtest helps module handler tests call handlers as the
// authenticated user without Auth and Permission middleware.
package httpxtest

import (
	"bazaar/internal/platform/httpx"
	"bazaar/internal/platform/identity"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Caller is the user Serve puts into the request context.
var Caller = identity.User{Id: "admin-1", CompanyId: "company-from-token"}

// Serve sends the request to router as Caller and returns the response.
func Serve(t *testing.T, router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	caller := Caller
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	ctx := identity.WithUser(req.Context(), &caller)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req.WithContext(ctx))
	return res
}

// ErrorCode returns the error code of an httpx.Error response body.
func ErrorCode(t *testing.T, res *httptest.ResponseRecorder) string {
	t.Helper()

	var body httpx.Error
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	return body.Error.Code
}
