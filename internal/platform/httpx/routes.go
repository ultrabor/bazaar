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

		r.Post("/user", mw.Permission(h.CreateUser, "users.create"))
		r.Get("/user/{id}", mw.Permission(h.GetUserById, "users.view"))

		r.Post("/location", mw.Permission(h.CreateLocation, "locations.create"))
		r.Get("/location", mw.Permission(h.GetAllLocations, "locations.view"))
		r.Get("/location/{id}", mw.Permission(h.GetLocationById, "locations.view"))
		r.Put("/location/{id}", mw.Permission(h.UpdateLocation, "locations.edit"))
		r.Patch("/location/{id}/archive", mw.Permission(h.ArchiveLocation, "locations.archive"))

		r.Post("/product", mw.Permission(h.CreateProduct, "products.create"))
		r.Get("/product", mw.Permission(h.GetAllProducts, "products.view"))
		r.Get("/product/{id}", mw.Permission(h.GetProductById, "products.view"))
		r.Put("/product/{id}", mw.Permission(h.UpdateProduct, "products.edit"))
		r.Patch("/product/{id}/archive", mw.Permission(h.ArchiveProduct, "products.archive"))

		r.Post("/category", mw.Permission(h.CreateCategory, "categories.create"))
		r.Get("/category", mw.Permission(h.GetAllCategories, "categories.view"))
		r.Get("/category/{id}", mw.Permission(h.GetCategoryById, "categories.view"))
		r.Put("/category/{id}", mw.Permission(h.UpdateCategory, "categories.edit"))
		r.Patch("/category/{id}/archive", mw.Permission(h.ArchiveCategory, "categories.archive"))

		r.Get("/permissions", mw.Permission(h.GetListPermissions, "permissions.view"))
		r.Post("/permission/set", mw.Permission(h.SetPermission, "permissions.grant"))
		r.Post("/permission/remove", mw.Permission(h.RemovePermission, "permissions.revoke"))
	})

	h.router.Get("/healthz", h.Health)
	h.router.Get("/readyz", h.Ready)

	h.router.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	h.router.Get("/swagger/*", swaggerHandler)
}
