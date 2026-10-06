package dto

import "bazaar/internal/modules/product"

type CreateProductRequest struct {
	Name string       `json:"name"`
	SKU  string       `json:"sku"`
	Unit product.Unit `json:"unit"`
}

type GetAllProductsRequest struct {
	Limit int `json:"limit"`
	Page  int `json:"page"`
}

type UpdateProductRequest struct {
	Name string       `json:"name"`
	SKU  string       `json:"sku"`
	Unit product.Unit `json:"unit"`
}
