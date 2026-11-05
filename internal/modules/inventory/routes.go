package inventory

import (
	"bazaar/internal/platform/httpx/middleware"
	"context"
	"log/slog"

	"github.com/go-chi/chi/v5"
)

// inventoryService is what the handlers call; Service implements it.
type inventoryService interface {
	ListBalances(ctx context.Context, f BalanceFilter) ([]Balance, error)
	ListMovements(ctx context.Context, f MovementFilter) ([]Movement, error)
}

type Handler struct {
	logger *slog.Logger
	svc    inventoryService
}

func NewHandler(logger *slog.Logger, svc inventoryService) *Handler {
	return &Handler{logger: logger, svc: svc}
}

// Routes are read-only: stock changes only through Service, called by
// purchasing, transfers, stocktakes, sales and returns.
func (h *Handler) Routes(r chi.Router, mw *middleware.Middleware) {
	r.Group(func(r chi.Router) {
		r.Use(mw.Auth)

		r.Get("/location/{id}/stock", mw.Permission(h.ListStock, "inventory.view"))
		r.Get("/location/{id}/stock/{productId}/movements", mw.Permission(h.ListMovements, "inventory.view"))
	})
}
