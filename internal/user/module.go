// Package user is the module of company users. Other modules use only its
// exported API; tables of the module are written only from this package.
package user

import (
	"bazaar/internal/platform/database"
	"bazaar/internal/platform/httpx/middleware"
	"log/slog"

	"github.com/go-chi/chi/v5"
)

type Module struct {
	svc     *Service
	handler *Handler
}

func New(db *database.Database, logger *slog.Logger, roles roleService) *Module {
	svc := NewService(NewRepository(db), db, roles)
	return &Module{svc: svc, handler: NewHandler(logger, svc)}
}

// Service returns the use cases other modules may call.
func (m *Module) Service() *Service {
	return m.svc
}

// Routes mounts the module's HTTP endpoints on r.
func (m *Module) Routes(r chi.Router, mw *middleware.Middleware) {
	m.handler.Routes(r, mw)
}
