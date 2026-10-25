package category

import (
	"bazaar/internal/domain"
	"context"
	"strings"

	"github.com/google/uuid"
)

type service struct {
	repo domain.CategoryRepository
}

var _ domain.CategoryService = (*service)(nil)

func NewService(repo domain.CategoryRepository) domain.CategoryService {
	return &service{repo: repo}
}

func (s *service) CreateCategory(ctx context.Context, rq *domain.CreateCategoryRequest) (*domain.CreateCategoryResponse, error) {
	if rq == nil || rq.CompanyId == "" {
		return nil, domain.ErrInvalidCred
	}
	rq.Name = strings.TrimSpace(rq.Name)
	if rq.Name == "" {
		return nil, domain.ErrInvalidCategoryName
	}
	return s.repo.CreateCategory(ctx, rq)
}

func (s *service) GetCompanyCategories(ctx context.Context, rq domain.GetCompanyCategoriesRequest) ([]domain.Category, error) {
	if rq.CompanyId == "" {
		return nil, domain.ErrInvalidCred
	}
	return s.repo.GetCompanyCategories(ctx, rq)
}

func (s *service) GetCategoryById(ctx context.Context, categoryId, companyId string) (*domain.Category, error) {
	if _, err := uuid.Parse(categoryId); err != nil {
		return nil, domain.ErrInvalidCategoryID
	}
	if companyId == "" {
		return nil, domain.ErrInvalidCred
	}
	return s.repo.GetCategoryById(ctx, categoryId, companyId)
}

func (s *service) UpdateCategory(ctx context.Context, rq *domain.UpdateCategoryRequest) (*domain.UpdateCategoryResponse, error) {
	if rq == nil || rq.CompanyId == "" {
		return nil, domain.ErrInvalidCred
	}
	if _, err := uuid.Parse(rq.CategoryId); err != nil {
		return nil, domain.ErrInvalidCategoryID
	}
	rq.Name = strings.TrimSpace(rq.Name)
	if rq.Name == "" {
		return nil, domain.ErrInvalidCategoryName
	}
	return s.repo.UpdateCategory(ctx, rq)
}

func (s *service) ArchiveCategory(ctx context.Context, categoryId, companyId string) error {
	if _, err := uuid.Parse(categoryId); err != nil {
		return domain.ErrInvalidCategoryID
	}
	if companyId == "" {
		return domain.ErrInvalidCred
	}
	return s.repo.ArchiveCategory(ctx, categoryId, companyId)
}
