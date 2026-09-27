package httpx

import "bazaar/internal/platform/httpx/middleware"

func (h *Handler) Routes(mw *middleware.Middleware) {
	h.router.Use(
		mw.Logger,
		mw.Recover,
	)
	h.router.Get("/healthz", h.Health)
	h.router.Get("/readyz", h.Ready)
	h.router.Post("/auth/register-owner", h.RegisterOwner)
	h.router.Get("/swagger", h.SwaggerUI)
	h.router.Get("/swagger/index.html", h.SwaggerUI)
	h.router.Get("/swagger/openapi.json", h.OpenAPI)
}
