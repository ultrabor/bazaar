package domain

import (
	"context"
	"errors"
)

type LocationType string

const (
	LocationTypeStore     LocationType = "store"
	LocationTypeWarehouse LocationType = "warehouse"
)

type Location struct {
	Id        string       `json:"id"`
	CompanyId string       `json:"company_id"`
	Name      string       `json:"name"`
	Address   string       `json:"address"`
	Type      LocationType `json:"type"`
	Archived  bool         `json:"archived"`
}

type CreateLocationRequest struct {
	CompanyId string       `json:"company_id"`
	Name      string       `json:"name"`
	Address   string       `json:"address"`
	Type      LocationType `json:"type"`
}

type CreateLocationResponse struct {
	LocationID string `json:"location_id"`
}

type UpdateLocationRequest struct {
	LocationId string       `json:"location_id"`
	CompanyId  string       `json:"company_id"`
	Name       string       `json:"name"`
	Address    string       `json:"address"`
	Type       LocationType `json:"type"`
}

type UpdateLocationResponse struct {
	LocationID string `json:"location_id"`
}

var (
	ErrLocationNotFound      = errors.New("location not found")
	ErrLocationAlreadyExists = errors.New("location already exists")
	ErrInvalidLocationType   = errors.New("invalid location type")
)

type LocationRepository interface {
	// GetLocationById returns an active location or ErrLocationNotFound.
	GetLocationById(ctx context.Context, locationId, companyId string) (*Location, error)
	GetCompanyLocations(ctx context.Context, companyId string) ([]Location, error)
	// CreateLocation returns ErrLocationAlreadyExists for a duplicate active name.
	CreateLocation(ctx context.Context, req *CreateLocationRequest) (*CreateLocationResponse, error)
	// UpdateLocation returns ErrLocationNotFound or ErrLocationAlreadyExists.
	UpdateLocation(ctx context.Context, req *UpdateLocationRequest) (*UpdateLocationResponse, error)
	// ArchiveLocation returns ErrLocationNotFound when nothing was archived.
	ArchiveLocation(ctx context.Context, locationId, companyId string) error
}

type LocationService interface {
	GetLocationById(ctx context.Context, locationId, companyId string) (*Location, error)
	GetCompanyLocations(ctx context.Context, companyId string) ([]Location, error)
	CreateLocation(ctx context.Context, req *CreateLocationRequest) (*CreateLocationResponse, error)
	UpdateLocation(ctx context.Context, req *UpdateLocationRequest) (*UpdateLocationResponse, error)
	ArchiveLocation(ctx context.Context, locationId, companyId string) error
}
