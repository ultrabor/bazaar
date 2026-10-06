package dto

import "bazaar/internal/modules/location"

type CreateLocationRequest struct {
	Name    string        `json:"name"`
	Address string        `json:"address"`
	Type    location.Type `json:"type"`
}

type UpdateLocationRequest struct {
	Name    string        `json:"name"`
	Address string        `json:"address"`
	Type    location.Type `json:"type"`
}
