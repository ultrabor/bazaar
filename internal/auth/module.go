// Package auth is the module of owner registration, login and token users. Other modules use only its
// exported API; tables of the module are written only from this package.
package auth

import (
	"bazaar/internal/platform/database"
	"bazaar/internal/platform/httpx/middleware"
	"bazaar/internal/platform/identity"
	"context"
	"log/slog"

	"github.com/go-chi/chi/v5"
)

type Module struct {
	svc     *Service
	handler *Handler
}

func New(db *database.Database, logger *slog.Logger, roles roleService, jwtSecret []byte) *Module {
	svc := NewService(NewRepository(db), db, roles, jwtSecret)
	return &Module{svc: svc, handler: NewHandler(logger, svc)}
}

// Service returns the use cases other modules may call.
func (m *Module) Service() *Service {
	return m.svc
}

// LoadUser returns the active user of a token. It matches
// middleware.UserLoader: a missing user is apperror.ErrInvalidCred.
func (m *Module) LoadUser(ctx context.Context, userId string) (*identity.User, error) {
	u, err := m.svc.GetUserById(ctx, userId)
	if err != nil {
		return nil, err
	}

	return &identity.User{
		Id:         u.Id,
		CompanyId:  u.CompanyId,
		UserRoleId: u.UserRoleId,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		Phone:      u.Phone,
	}, nil
}

// Routes mounts the module's HTTP endpoints on r.
func (m *Module) Routes(r chi.Router, mw *middleware.Middleware) {
	m.handler.Routes(r, mw)
}
