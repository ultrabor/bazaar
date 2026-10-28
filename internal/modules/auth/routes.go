package auth

import (
	"bazaar/internal/platform/httpx/middleware"
	"context"
	"log/slog"

	"github.com/go-chi/chi/v5"
)

// authService is what the handlers call; Service implements it.
type authService interface {
	// RegisterOwner creates the company, its owner role with all permissions
	// and the owner user in one transaction.
	RegisterOwner(ctx context.Context, rq RegisterRequest) (*RegisterResponse, error)
	Login(ctx context.Context, rq LoginRequest) (*LoginResponse, error)
}

type Handler struct {
	logger *slog.Logger
	svc    authService
}

func NewHandler(logger *slog.Logger, svc authService) *Handler {
	return &Handler{logger: logger, svc: svc}
}

func (h *Handler) Routes(r chi.Router, mw *middleware.Middleware) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register-owner", h.RegisterOwner)
		r.Post("/login", h.Login)

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth)

			r.Get("/me", h.Me)
		})
	})
}
