// Package permission is the module of personal user permissions. Other modules use only its
// exported API; tables of the module are written only from this package.
package permission

import (
	"bazaar/internal/platform/database"
	"bazaar/internal/platform/httpx/middleware"
	"context"
	"log/slog"

	"github.com/go-chi/chi/v5"
)

type Module struct {
	svc     *Service
	handler *Handler
}

func New(db *database.Database, logger *slog.Logger, users userService) *Module {
	svc := NewService(NewRepository(db), users)
	return &Module{svc: svc, handler: NewHandler(logger, svc)}
}

// Service returns the use cases other modules may call.
func (m *Module) Service() *Service {
	return m.svc
}

// HasPermission reports whether the user holds the permission code. It
// matches middleware.PermissionChecker.
func (m *Module) HasPermission(ctx context.Context, userId, code string) (bool, error) {
	return m.svc.HasPermission(ctx, userId, code)
}

// Routes mounts the module's HTTP endpoints on r.
func (m *Module) Routes(r chi.Router, mw *middleware.Middleware) {
	m.handler.Routes(r, mw)
}
