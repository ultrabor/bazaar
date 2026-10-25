package domain

import (
	"context"
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

// Product operations also return ErrInvalidCategoryID and
// ErrCategoryNotFound from category.go for a bad or inactive category.
var (
	ErrProductNotFound      = errors.New("product not found")
	ErrProductAlreadyExists = errors.New("product already exists")
	ErrInvalidUnitType      = errors.New("invalid product unit type")
	ErrInvalidProductID     = errors.New("invalid product id")
)

type ProductRepository interface {
	// GetProductById returns an active product or ErrProductNotFound.
	GetProductById(ctx context.Context, productId, companyId string) (*Product, error)
	GetCompanyProducts(ctx context.Context, rq GetCompanyProductsRequest) ([]Product, error)
	// CreateProduct returns ErrProductAlreadyExists for a duplicate active SKU
	// and ErrCategoryNotFound when the category is missing or archived.
	CreateProduct(ctx context.Context, rq *CreateProductRequest) (*CreateProductResponse, error)
	// UpdateProduct returns ErrProductNotFound, ErrProductAlreadyExists or
	// ErrCategoryNotFound.
	UpdateProduct(ctx context.Context, rq *UpdateProductRequest) (*UpdateProductResponse, error)
	// ArchiveProduct returns ErrProductNotFound when nothing was archived.
	ArchiveProduct(ctx context.Context, productId, companyId string) error
}

type ProductService interface {
	GetProductById(ctx context.Context, productId, companyId string) (*Product, error)
	GetCompanyProducts(ctx context.Context, rq GetCompanyProductsRequest) ([]Product, error)
	CreateProduct(ctx context.Context, rq *CreateProductRequest) (*CreateProductResponse, error)
	UpdateProduct(ctx context.Context, rq *UpdateProductRequest) (*UpdateProductResponse, error)
	ArchiveProduct(ctx context.Context, productId, companyId string) error
}
