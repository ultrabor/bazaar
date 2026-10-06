package product

import "errors"

type Product struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	CompanyId string `json:"company_id"`
	SKU       string `json:"sku"`
	Unit      string `json:"unit"`
	Archived  bool   `json:"archived"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type CreateProductRequest struct {
	CompanyId string `json:"company_id"`
	Name      string `json:"name"`
	SKU       string `json:"sku"`
	Unit      string `json:"unit"`
}

type CreateProductResponse struct {
	ProductId string `json:"product_id"`
}

type UpdateProductRequest struct {
	ProductId string `json:"product_id"`
	CompanyId string `json:"company_id"`
	Name      string `json:"name"`
	SKU       string `json:"sku"`
	Unit      string `json:"unit"`
}

type UpdateProductResponse struct {
	ProductId string `json:"product_id"`
}

var ErrInvalidCred = errors.New("invalid credentials")
var ErrNotFound = errors.New("no rows in result set")
var ErrProductAlreadyExists = errors.New("product already exists")
var ErrInvalidUnitType = errors.New("invalid product unit type")
