package httpx

import (
	"bazaar/internal/platform/httpx/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func (h *Handler) Routes(mw *middleware.Middleware) {

	h.router.Use(chimiddleware.RedirectSlashes)
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

	h.router.Group(func(r chi.Router) {
		r.Use(mw.Auth)

		r.Post("/user", h.CreateUser)
		r.Get("/user/{id}", h.GetUserById)

		r.Post("/location", h.CreateLocation)
		r.Get("/location", h.GetAllLocations)
		r.Get("/location/{id}", h.GetLocationById)
		r.Put("/location/{id}", h.UpdateLocation)
		r.Post("/location/{id}/archive", h.ArchiveLocation)
	})

	h.router.Get("/healthz", h.Health)
	h.router.Get("/readyz", h.Ready)

	h.router.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	h.router.Get("/swagger/*", swaggerHandler)
}
