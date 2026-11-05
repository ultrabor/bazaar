package inventory

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

func New(db *database.Database, logger *slog.Logger) *Module {
	svc := NewService(NewRepository(db), db)
	return &Module{svc: svc, handler: NewHandler(logger, svc)}
}

func (m *Module) Service() *Service {
	return m.svc
}

func (m *Module) Routes(r chi.Router, mw *middleware.Middleware) {
	m.handler.Routes(r, mw)
}
