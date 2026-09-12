package httpx

import "bazaar/internal/platform/httpx/middleware"

func (h *Handler) Routes(mw *middleware.Middleware) {
	h.router.Use(
		mw.Logger,
		mw.Recover,
	)
	h.router.Get("/healthz", h.Health)
	h.router.Get("/readyz", h.Ready)
}
