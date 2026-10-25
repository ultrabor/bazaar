package dto

import "bazaar/internal/domain"

type CreateProductRequest struct {
	Name       string             `json:"name"`
	SKU        string             `json:"sku"`
	Unit       domain.ProductUnit `json:"unit"`
	CategoryId *string            `json:"category_id"`
}

type UpdateProductRequest struct {
	Name       string             `json:"name"`
	SKU        string             `json:"sku"`
	Unit       domain.ProductUnit `json:"unit"`
	CategoryId *string            `json:"category_id"`
}
