package location

import "errors"

type Type string

const (
	TypeStore     Type = "store"
	TypeWarehouse Type = "warehouse"
)

type Location struct {
	Id        string `json:"id"`
	CompanyId string `json:"company_id"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	Type      Type   `json:"type"`
	Archived  bool   `json:"archived"`
}

type CreateLocationRequest struct {
	CompanyId string `json:"company_id"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	Type      Type   `json:"type"`
}

type CreateLocationResponse struct {
	LocationID string `json:"location_id"`
}

type UpdateLocationRequest struct {
	LocationId string `json:"location_id"`
	CompanyId  string `json:"company_id"`
	Name       string `json:"name"`
	Address    string `json:"address"`
	Type       Type   `json:"type"`
}

type UpdateLocationResponse struct {
	LocationID string `json:"location_id"`
}

var ErrInvalidCred = errors.New("invalid credentials")
var ErrNotFound = errors.New("no rows in result set")
var ErrLocationAlreadyExists = errors.New("location already exists")
var ErrInvalidType = errors.New("invalid location type")
