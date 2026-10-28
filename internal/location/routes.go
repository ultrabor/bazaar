package location

import (
	"bazaar/internal/platform/httpx/middleware"
	"context"
	"log/slog"

	"github.com/go-chi/chi/v5"
)

// locationService is what the handlers call; Service implements it.
type locationService interface {
	GetLocationById(ctx context.Context, locationId, companyId string) (*Location, error)
	GetCompanyLocations(ctx context.Context, companyId string) ([]Location, error)
	CreateLocation(ctx context.Context, req *CreateLocationRequest) (*CreateLocationResponse, error)
	UpdateLocation(ctx context.Context, req *UpdateLocationRequest) (*UpdateLocationResponse, error)
	ArchiveLocation(ctx context.Context, locationId, companyId string) error
}

type Handler struct {
	logger *slog.Logger
	svc    locationService
}

func NewHandler(logger *slog.Logger, svc locationService) *Handler {
	return &Handler{logger: logger, svc: svc}
}

func (h *Handler) Routes(r chi.Router, mw *middleware.Middleware) {
	r.Group(func(r chi.Router) {
		r.Use(mw.Auth)

		r.Post("/location", mw.Permission(h.CreateLocation, "locations.create"))
		r.Get("/location", mw.Permission(h.GetAllLocations, "locations.view"))
		r.Get("/location/{id}", mw.Permission(h.GetLocationById, "locations.view"))
		r.Put("/location/{id}", mw.Permission(h.UpdateLocation, "locations.edit"))
		r.Patch("/location/{id}/archive", mw.Permission(h.ArchiveLocation, "locations.archive"))
	})
}
