package product

import (
	"context"
	"strings"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetProductById(ctx context.Context, productId, companyId string) (*Product, error) {
	product, err := s.repo.GetProductById(ctx, productId, companyId)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (s *Service) GetCompanyProducts(ctx context.Context, rq GetCompanyProductsRequest) ([]Product, error) {
	products, err := s.repo.GetCompanyProducts(ctx, rq)
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (s *Service) CreateProduct(ctx context.Context, r *CreateProductRequest) (*CreateProductResponse, error) {
	if r == nil || r.CompanyId == "" {
		return nil, ErrInvalidCred
	}

	if r.Unit != UnitKg && r.Unit != UnitLiter && r.Unit != UnitMeter && r.Unit != UnitPiece {
		return nil, ErrInvalidUnitType
	}

	r.Name = strings.TrimSpace(r.Name)
	r.SKU = strings.TrimSpace(r.SKU)

	if r.Name == "" || r.SKU == "" {
		return nil, ErrInvalidCred
	}

	res, err := s.repo.CreateProduct(ctx, r)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Service) UpdateProduct(ctx context.Context, r *UpdateProductRequest) (*UpdateProductResponse, error) {
	if r == nil || r.CompanyId == "" || r.ProductId == "" {
		return nil, ErrInvalidCred
	}

	if r.Unit != UnitKg && r.Unit != UnitLiter && r.Unit != UnitMeter && r.Unit != UnitPiece {
		return nil, ErrInvalidUnitType
	}

	r.Name = strings.TrimSpace(r.Name)
	r.SKU = strings.TrimSpace(r.SKU)

	if r.Name == "" || r.SKU == "" {
		return nil, ErrInvalidCred
	}

	res, err := s.repo.UpdateProduct(ctx, r)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Service) ArchiveProduct(ctx context.Context, productId, companyId string) error {
	if productId == "" || companyId == "" {
		return ErrInvalidCred
	}

	err := s.repo.ArchiveProduct(ctx, productId, companyId)

	return err
}
