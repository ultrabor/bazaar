package dto

type CreateProductRequest struct {
	Name string `json:"name"`
	SKU  string `json:"sku"`
	Unit string `json:"unit"`
}

type UpdateProductRequest struct {
	Name string `json:"name"`
	SKU  string `json:"sku"`
	Unit string `json:"unit"`
}
