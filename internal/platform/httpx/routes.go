package httpx

import (
	"bazaar/internal/platform/httpx/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) Routes(mw *middleware.Middleware) {

	h.router.Use(mw.Logger)
	h.router.Use(mw.Recover)

	h.router.Route("/auth", func(r chi.Router) {
		r.Post("/register-owner", h.RegisterOwner)
		r.Post("/login", h.Login)

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth)

			r.Get("/me", h.Me)
		})
	})

	h.router.Route("/users", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(mw.Auth)

			r.Post("/", h.CreateUser)
		})
	})

	h.router.Get("/healthz", h.Health)
	h.router.Get("/readyz", h.Ready)

	h.router.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	h.router.Get("/swagger/*", swaggerHandler)
}
