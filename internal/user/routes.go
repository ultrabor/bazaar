package user

import (
	"bazaar/internal/platform/httpx/middleware"
	"context"
	"log/slog"

	"github.com/go-chi/chi/v5"
)

// userService is what the handlers call; Service implements it.
type userService interface {
	GetUserById(ctx context.Context, userId, companyId string) (*User, error)
	CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error)
}

type Handler struct {
	logger *slog.Logger
	svc    userService
}

func NewHandler(logger *slog.Logger, svc userService) *Handler {
	return &Handler{logger: logger, svc: svc}
}

func (h *Handler) Routes(r chi.Router, mw *middleware.Middleware) {
	r.Group(func(r chi.Router) {
		r.Use(mw.Auth)

		r.Post("/user", mw.Permission(h.CreateUser, "users.create"))
		r.Get("/user/{id}", mw.Permission(h.GetUserById, "users.view"))
	})
}
