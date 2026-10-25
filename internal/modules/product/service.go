package product

import (
	"bazaar/internal/domain"
	"context"
	"strings"

	"github.com/google/uuid"
)

type service struct {
	repo domain.ProductRepository
}

var _ domain.ProductService = (*service)(nil)

func NewService(repo domain.ProductRepository) domain.ProductService {
	return &service{repo: repo}
}

func (s *service) GetProductById(ctx context.Context, productId, companyId string) (*domain.Product, error) {
	if _, err := uuid.Parse(productId); err != nil {
		return nil, domain.ErrInvalidProductID
	}

	product, err := s.repo.GetProductById(ctx, productId, companyId)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (s *service) GetCompanyProducts(ctx context.Context, rq domain.GetCompanyProductsRequest) ([]domain.Product, error) {
	rq.Name = escapeLikePattern(strings.TrimSpace(rq.Name))
	rq.SKU = strings.TrimSpace(rq.SKU)
	rq.Unit = domain.ProductUnit(strings.ToLower(strings.TrimSpace(string(rq.Unit))))

	if rq.Unit != "" && rq.Unit != domain.ProductUnitKg && rq.Unit != domain.ProductUnitLiter && rq.Unit != domain.ProductUnitMeter && rq.Unit != domain.ProductUnitPiece {
		return nil, domain.ErrInvalidUnitType
	}
	if rq.CategoryId != "" {
		if _, err := uuid.Parse(rq.CategoryId); err != nil {
			return nil, domain.ErrInvalidCategoryID
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

func (s *service) CreateProduct(ctx context.Context, r *domain.CreateProductRequest) (*domain.CreateProductResponse, error) {
	if r == nil || r.CompanyId == "" {
		return nil, domain.ErrInvalidCred
	}

	if r.Unit != domain.ProductUnitKg && r.Unit != domain.ProductUnitLiter && r.Unit != domain.ProductUnitMeter && r.Unit != domain.ProductUnitPiece {
		return nil, domain.ErrInvalidUnitType
	}

	r.Name = strings.TrimSpace(r.Name)
	r.SKU = strings.TrimSpace(r.SKU)

	if r.Name == "" || r.SKU == "" {
		return nil, domain.ErrInvalidCred
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

func (s *service) UpdateProduct(ctx context.Context, r *domain.UpdateProductRequest) (*domain.UpdateProductResponse, error) {
	if r == nil || r.CompanyId == "" || r.ProductId == "" {
		return nil, domain.ErrInvalidCred
	}
	if _, err := uuid.Parse(r.ProductId); err != nil {
		return nil, domain.ErrInvalidProductID
	}

	if r.Unit != domain.ProductUnitKg && r.Unit != domain.ProductUnitLiter && r.Unit != domain.ProductUnitMeter && r.Unit != domain.ProductUnitPiece {
		return nil, domain.ErrInvalidUnitType
	}

	r.Name = strings.TrimSpace(r.Name)
	r.SKU = strings.TrimSpace(r.SKU)

	if r.Name == "" || r.SKU == "" {
		return nil, domain.ErrInvalidCred
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

func (s *service) ArchiveProduct(ctx context.Context, productId, companyId string) error {
	if productId == "" || companyId == "" {
		return domain.ErrInvalidCred
	}
	if _, err := uuid.Parse(productId); err != nil {
		return domain.ErrInvalidProductID
	}

	err := s.repo.ArchiveProduct(ctx, productId, companyId)

	return err
}

func validateCategoryId(categoryId *string) error {
	if categoryId == nil {
		return nil
	}
	if _, err := uuid.Parse(*categoryId); err != nil {
		return domain.ErrInvalidCategoryID
	}
	return nil
}
