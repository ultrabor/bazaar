package httpx

import (
	"bazaar/internal/platform/httpx/middleware"
	"net/http"
)

func (h *Handler) Routes(mw *middleware.Middleware) {
	h.router.Use(
		mw.Logger,
		mw.Recover,
	)
	h.router.Get("/healthz", h.Health)
	h.router.Get("/readyz", h.Ready)
	h.router.Post("/auth/register-owner", h.RegisterOwner)
	h.router.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	h.router.Get("/swagger/*", swaggerHandler)
}
