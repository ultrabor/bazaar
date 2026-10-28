package product

import (
	"errors"
	"time"
)

type ProductUnit string

const (
	ProductUnitKg    ProductUnit = "kg"
	ProductUnitMeter ProductUnit = "meter"
	ProductUnitPiece ProductUnit = "piece"
	ProductUnitLiter ProductUnit = "liter"
)

type Product struct {
	Id         string      `json:"id"`
	Name       string      `json:"name"`
	CompanyId  string      `json:"company_id"`
	CategoryId *string     `json:"category_id"`
	SKU        string      `json:"sku"`
	Unit       ProductUnit `json:"unit"`
	Archived   bool        `json:"archived"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type CreateProductRequest struct {
	CompanyId  string      `json:"company_id"`
	CategoryId *string     `json:"category_id"`
	Name       string      `json:"name"`
	SKU        string      `json:"sku"`
	Unit       ProductUnit `json:"unit"`
}

type CreateProductResponse struct {
	ProductId string `json:"product_id"`
}

type UpdateProductRequest struct {
	ProductId  string      `json:"product_id"`
	CompanyId  string      `json:"company_id"`
	CategoryId *string     `json:"category_id"`
	Name       string      `json:"name"`
	SKU        string      `json:"sku"`
	Unit       ProductUnit `json:"unit"`
}

type UpdateProductResponse struct {
	ProductId string `json:"product_id"`
}

type GetCompanyProductsRequest struct {
	CompanyId  string      `json:"company_id"`
	Name       string      `json:"name,omitempty"`
	SKU        string      `json:"sku,omitempty"`
	Unit       ProductUnit `json:"unit,omitempty"`
	CategoryId string      `json:"category_id,omitempty"`
	Page       int         `json:"page,omitempty"`
	Limit      int         `json:"limit,omitempty"`
}

var (
	ErrProductNotFound      = errors.New("product not found")
	ErrProductAlreadyExists = errors.New("product already exists")
	ErrInvalidUnitType      = errors.New("invalid product unit type")
	ErrInvalidProductID     = errors.New("invalid product id")
	// ErrInvalidCategoryID and ErrCategoryNotFound report a bad or inactive
	// category of a product. Product checks the categories table itself, so
	// it keeps its own errors instead of importing the category module.
	ErrInvalidCategoryID = errors.New("invalid category id")
	ErrCategoryNotFound  = errors.New("category not found")
)
