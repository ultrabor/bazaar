package product

import (
	"bazaar/internal/platform/support/apperror"
	"context"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	repo repository
}

func NewService(repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetProductById(ctx context.Context, productId, companyId string) (*Product, error) {
	if _, err := uuid.Parse(productId); err != nil {
		return nil, ErrInvalidProductID
	}

	product, err := s.repo.GetProductById(ctx, productId, companyId)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (s *Service) GetCompanyProducts(ctx context.Context, rq GetCompanyProductsRequest) ([]Product, error) {
	rq.Name = escapeLikePattern(strings.TrimSpace(rq.Name))
	rq.SKU = strings.TrimSpace(rq.SKU)
	rq.Unit = ProductUnit(strings.ToLower(strings.TrimSpace(string(rq.Unit))))

	if rq.Unit != "" && rq.Unit != ProductUnitKg && rq.Unit != ProductUnitLiter && rq.Unit != ProductUnitMeter && rq.Unit != ProductUnitPiece {
		return nil, ErrInvalidUnitType
	}
	if rq.CategoryId != "" {
		if _, err := uuid.Parse(rq.CategoryId); err != nil {
			return nil, ErrInvalidCategoryID
		}
	}

	products, err := s.repo.GetCompanyProducts(ctx, rq)
	if err != nil {
		return nil, err
	}

	return products, nil
}

func escapeLikePattern(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `%`, `\%`)
	return strings.ReplaceAll(value, `_`, `\_`)
}

func (s *Service) CreateProduct(ctx context.Context, r *CreateProductRequest) (*CreateProductResponse, error) {
	if r == nil || r.CompanyId == "" {
		return nil, apperror.ErrInvalidCred
	}

	if r.Unit != ProductUnitKg && r.Unit != ProductUnitLiter && r.Unit != ProductUnitMeter && r.Unit != ProductUnitPiece {
		return nil, ErrInvalidUnitType
	}

	r.Name = strings.TrimSpace(r.Name)
	r.SKU = strings.TrimSpace(r.SKU)

	if r.Name == "" || r.SKU == "" {
		return nil, apperror.ErrInvalidCred
	}
	if err := validateCategoryId(r.CategoryId); err != nil {
		return nil, err
	}

	res, err := s.repo.CreateProduct(ctx, r)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Service) UpdateProduct(ctx context.Context, r *UpdateProductRequest) (*UpdateProductResponse, error) {
	if r == nil || r.CompanyId == "" || r.ProductId == "" {
		return nil, apperror.ErrInvalidCred
	}
	if _, err := uuid.Parse(r.ProductId); err != nil {
		return nil, ErrInvalidProductID
	}

	if r.Unit != ProductUnitKg && r.Unit != ProductUnitLiter && r.Unit != ProductUnitMeter && r.Unit != ProductUnitPiece {
		return nil, ErrInvalidUnitType
	}

	r.Name = strings.TrimSpace(r.Name)
	r.SKU = strings.TrimSpace(r.SKU)

	if r.Name == "" || r.SKU == "" {
		return nil, apperror.ErrInvalidCred
	}
	if err := validateCategoryId(r.CategoryId); err != nil {
		return nil, err
	}

	res, err := s.repo.UpdateProduct(ctx, r)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Service) ArchiveProduct(ctx context.Context, productId, companyId string) error {
	if productId == "" || companyId == "" {
		return apperror.ErrInvalidCred
	}
	if _, err := uuid.Parse(productId); err != nil {
		return ErrInvalidProductID
	}

	err := s.repo.ArchiveProduct(ctx, productId, companyId)

	return err
}

func validateCategoryId(categoryId *string) error {
	if categoryId == nil {
		return nil
	}
	if _, err := uuid.Parse(*categoryId); err != nil {
		return ErrInvalidCategoryID
	}
	return nil
}

// repository is the storage Service needs; Repository implements it.
type repository interface {
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
