package httpx

import (
	_ "embed"
	"net/http"
)

//go:embed swagger/openapi.json
var openAPISpec []byte

//go:embed swagger/index.html
var swaggerPage []byte

func (h *Handler) SwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(swaggerPage)
}

func (h *Handler) OpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write(openAPISpec)
}
