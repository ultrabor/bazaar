package dto

import "bazaar/internal/domain"

type CreateLocationRequest struct {
	Name    string              `json:"name"`
	Address string              `json:"address"`
	Type    domain.LocationType `json:"type"`
}

type UpdateLocationRequest struct {
	Name    string              `json:"name"`
	Address string              `json:"address"`
	Type    domain.LocationType `json:"type"`
}
