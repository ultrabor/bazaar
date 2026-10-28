package location

import (
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
