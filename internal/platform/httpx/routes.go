package httpx

import (
	"net/http"

	"bazaar/internal/platform/httpx/middleware"

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

		r.Post("/user", mw.Permission(h.CreateUser, "user.create"))
		r.Get("/user/{id}", mw.Permission(h.GetUserById, "user.view"))

		r.Post("/location", mw.Permission(h.CreateLocation, "location.create"))
		r.Get("/location", mw.Permission(h.GetAllLocations, "location.view"))
		r.Get("/location/{id}", mw.Permission(h.GetLocationById, "location.view"))
		r.Put("/location/{id}", mw.Permission(h.UpdateLocation, "location.edit"))
		r.Patch("/location/{id}/archive", mw.Permission(h.ArchiveLocation, "location.archive"))

		r.Post("/product", mw.Permission(h.CreateProduct, "product.create"))
		r.Get("/product", mw.Permission(h.GetAllProducts, "product.view"))
		r.Get("/product/{id}", mw.Permission(h.GetProductById, "product.view"))
		r.Put("/product/{id}", mw.Permission(h.UpdateProduct, "product.edit"))
		r.Patch("/product/{id}/archive", mw.Permission(h.ArchiveProduct, "product.archive"))

		r.Post("/category", mw.Permission(h.CreateCategory, "category.create"))
		r.Get("/category", mw.Permission(h.GetAllCategories, "category.view"))
		r.Get("/category/{id}", mw.Permission(h.GetCategoryById, "category.view"))
		r.Put("/category/{id}", mw.Permission(h.UpdateCategory, "category.edit"))
		r.Patch("/category/{id}/archive", mw.Permission(h.ArchiveCategory, "category.archive"))

		r.Get("/permissions", h.GetListPermissions)
		r.Post("/permission/set", h.SetPermission)
		r.Post("/permission/remove", h.RemovePermission)
	})

	h.router.Get("/healthz", h.Health)
	h.router.Get("/readyz", h.Ready)

	h.router.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	h.router.Get("/swagger/*", swaggerHandler)
}
