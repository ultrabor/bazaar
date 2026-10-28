package app

import (
	"bazaar/internal/platform/health"
	"bazaar/internal/platform/httpx"
	"bazaar/internal/platform/httpx/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// Module is a feature module that serves HTTP endpoints.
type Module interface {
	Routes(r chi.Router, mw *middleware.Middleware)
}

// NewRouter sets the global middleware first, as chi requires, then mounts
// every module, the probes and Swagger.
func NewRouter(mw *middleware.Middleware, probes *health.Handler, modules ...Module) *chi.Mux {
	router := chi.NewRouter()
	router.Use(chimiddleware.RedirectSlashes)
	router.Use(mw.Logger)
	router.Use(mw.Recover)

	for _, m := range modules {
		m.Routes(router, mw)
	}

	probes.Routes(router)

	router.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	router.Get("/swagger/*", httpx.SwaggerHandler)

	return router
}
