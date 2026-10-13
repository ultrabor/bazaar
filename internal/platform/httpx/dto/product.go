package dto

import "bazaar/internal/modules/product"

type CreateProductRequest struct {
	Name       string       `json:"name"`
	SKU        string       `json:"sku"`
	Unit       product.Unit `json:"unit"`
	CategoryId *string      `json:"category_id"`
}

type UpdateProductRequest struct {
	Name       string       `json:"name"`
	SKU        string       `json:"sku"`
	Unit       product.Unit `json:"unit"`
	CategoryId *string      `json:"category_id"`
}
