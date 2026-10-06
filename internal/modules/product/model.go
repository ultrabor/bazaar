package product

import (
	"errors"
	"time"
)

type Unit string

const (
	UnitKg    Unit = "kg"
	UnitMeter Unit = "meter"
	UnitPiece Unit = "piece"
	UnitLiter Unit = "liter"
)

type Product struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	CompanyId string    `json:"company_id"`
	SKU       string    `json:"sku"`
	Unit      Unit      `json:"unit"`
	Archived  bool      `json:"archived"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateProductRequest struct {
	CompanyId string `json:"company_id"`
	Name      string `json:"name"`
	SKU       string `json:"sku"`
	Unit      Unit   `json:"unit"`
}

type CreateProductResponse struct {
	ProductId string `json:"product_id"`
}

type UpdateProductRequest struct {
	ProductId string `json:"product_id"`
	CompanyId string `json:"company_id"`
	Name      string `json:"name"`
	SKU       string `json:"sku"`
	Unit      Unit   `json:"unit"`
}

type UpdateProductResponse struct {
	ProductId string `json:"product_id"`
}

type GetCompanyProductsRequest struct {
	CompanyId string `json:"company_id"`
	Page      int    `json:"page,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

var ErrInvalidCred = errors.New("invalid credentials")
var ErrNotFound = errors.New("no rows in result set")
var ErrProductAlreadyExists = errors.New("product already exists")
var ErrInvalidUnitType = errors.New("invalid product unit type")
