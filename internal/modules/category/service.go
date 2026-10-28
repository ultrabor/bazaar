package category

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

func (s *Service) CreateCategory(ctx context.Context, rq *CreateCategoryRequest) (*CreateCategoryResponse, error) {
	if rq == nil || rq.CompanyId == "" {
		return nil, apperror.ErrInvalidCred
	}
	rq.Name = strings.TrimSpace(rq.Name)
	if rq.Name == "" {
		return nil, ErrInvalidCategoryName
	}
	return s.repo.CreateCategory(ctx, rq)
}

func (s *Service) GetCompanyCategories(ctx context.Context, rq GetCompanyCategoriesRequest) ([]Category, error) {
	if rq.CompanyId == "" {
		return nil, apperror.ErrInvalidCred
	}
	return s.repo.GetCompanyCategories(ctx, rq)
}

func (s *Service) GetCategoryById(ctx context.Context, categoryId, companyId string) (*Category, error) {
	if _, err := uuid.Parse(categoryId); err != nil {
		return nil, ErrInvalidCategoryID
	}
	if companyId == "" {
		return nil, apperror.ErrInvalidCred
	}
	return s.repo.GetCategoryById(ctx, categoryId, companyId)
}

func (s *Service) UpdateCategory(ctx context.Context, rq *UpdateCategoryRequest) (*UpdateCategoryResponse, error) {
	if rq == nil || rq.CompanyId == "" {
		return nil, apperror.ErrInvalidCred
	}
	if _, err := uuid.Parse(rq.CategoryId); err != nil {
		return nil, ErrInvalidCategoryID
	}
	rq.Name = strings.TrimSpace(rq.Name)
	if rq.Name == "" {
		return nil, ErrInvalidCategoryName
	}
	return s.repo.UpdateCategory(ctx, rq)
}

func (s *Service) ArchiveCategory(ctx context.Context, categoryId, companyId string) error {
	if _, err := uuid.Parse(categoryId); err != nil {
		return ErrInvalidCategoryID
	}
	if companyId == "" {
		return apperror.ErrInvalidCred
	}
	return s.repo.ArchiveCategory(ctx, categoryId, companyId)
}

// repository is the storage Service needs; Repository implements it.
type repository interface {
	// CreateCategory returns ErrCategoryAlreadyExists for a duplicate active name.
	CreateCategory(ctx context.Context, rq *CreateCategoryRequest) (*CreateCategoryResponse, error)
	GetCompanyCategories(ctx context.Context, rq GetCompanyCategoriesRequest) ([]Category, error)
	// GetCategoryById returns an active category or ErrCategoryNotFound.
	GetCategoryById(ctx context.Context, categoryId, companyId string) (*Category, error)
	// UpdateCategory returns ErrCategoryNotFound or ErrCategoryAlreadyExists.
	UpdateCategory(ctx context.Context, rq *UpdateCategoryRequest) (*UpdateCategoryResponse, error)
	// ArchiveCategory returns ErrCategoryNotFound when nothing was archived.
	ArchiveCategory(ctx context.Context, categoryId, companyId string) error
}
