package category

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateCategory(ctx context.Context, rq *CreateCategoryRequest) (*CreateCategoryResponse, error) {
	if rq == nil || rq.CompanyId == "" {
		return nil, ErrInvalidCredentials
	}
	rq.Name = strings.TrimSpace(rq.Name)
	if rq.Name == "" {
		return nil, ErrInvalidName
	}
	return s.repo.CreateCategory(ctx, rq)
}

func (s *Service) GetCompanyCategories(ctx context.Context, rq GetCompanyCategoriesRequest) ([]Category, error) {
	if rq.CompanyId == "" {
		return nil, ErrInvalidCredentials
	}
	return s.repo.GetCompanyCategories(ctx, rq)
}

func (s *Service) GetCategoryById(ctx context.Context, categoryId, companyId string) (*Category, error) {
	if _, err := uuid.Parse(categoryId); err != nil {
		return nil, ErrInvalidCategoryID
	}
	if companyId == "" {
		return nil, ErrInvalidCredentials
	}
	return s.repo.GetCategoryById(ctx, categoryId, companyId)
}

func (s *Service) UpdateCategory(ctx context.Context, rq *UpdateCategoryRequest) (*UpdateCategoryResponse, error) {
	if rq == nil || rq.CompanyId == "" {
		return nil, ErrInvalidCredentials
	}
	if _, err := uuid.Parse(rq.CategoryId); err != nil {
		return nil, ErrInvalidCategoryID
	}
	rq.Name = strings.TrimSpace(rq.Name)
	if rq.Name == "" {
		return nil, ErrInvalidName
	}
	return s.repo.UpdateCategory(ctx, rq)
}

func (s *Service) ArchiveCategory(ctx context.Context, categoryId, companyId string) error {
	if _, err := uuid.Parse(categoryId); err != nil {
		return ErrInvalidCategoryID
	}
	if companyId == "" {
		return ErrInvalidCredentials
	}
	return s.repo.ArchiveCategory(ctx, categoryId, companyId)
}
