package category

import (
	"errors"
	"time"
)

type Category struct {
	Id        string    `json:"id"`
	CompanyId string    `json:"company_id"`
	Name      string    `json:"name"`
	Archived  bool      `json:"archived"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateCategoryRequest struct {
	CompanyId string `json:"company_id"`
	Name      string `json:"name"`
}

type CreateCategoryResponse struct {
	CategoryId string `json:"category_id"`
}

type UpdateCategoryRequest struct {
	CategoryId string `json:"category_id"`
	CompanyId  string `json:"company_id"`
	Name       string `json:"name"`
}

type UpdateCategoryResponse struct {
	CategoryId string `json:"category_id"`
}

type GetCompanyCategoriesRequest struct {
	CompanyId string `json:"company_id"`
	Page      int    `json:"page,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

var (
	ErrInvalidCategoryID     = errors.New("invalid category id")
	ErrInvalidCategoryName   = errors.New("invalid category name")
	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryAlreadyExists = errors.New("category already exists")
)
