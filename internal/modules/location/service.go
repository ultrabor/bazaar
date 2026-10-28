package location

import (
	"bazaar/internal/platform/support/apperror"
	"context"
)

type Service struct {
	repo repository
}

func NewService(repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetCompanyLocations(ctx context.Context, companyId string) ([]Location, error) {
	locations, err := s.repo.GetCompanyLocations(ctx, companyId)
	if err != nil {
		return nil, err
	}

	return locations, nil
}

func (s *Service) GetLocationById(ctx context.Context, locationId, companyId string) (*Location, error) {
	return s.repo.GetLocationById(ctx, locationId, companyId)
}

func (s *Service) CreateLocation(ctx context.Context, r *CreateLocationRequest) (*CreateLocationResponse, error) {

	if r == nil || r.CompanyId == "" || r.Name == "" || r.Address == "" {
		return nil, apperror.ErrInvalidCred
	}
	if !validType(r.Type) {
		return nil, ErrInvalidLocationType
	}

	res, err := s.repo.CreateLocation(ctx, r)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Service) UpdateLocation(ctx context.Context, r *UpdateLocationRequest) (*UpdateLocationResponse, error) {

	if r == nil || r.CompanyId == "" || r.LocationId == "" || r.Name == "" || r.Address == "" {
		return nil, apperror.ErrInvalidCred
	}
	if !validType(r.Type) {
		return nil, ErrInvalidLocationType
	}

	res, err := s.repo.UpdateLocation(ctx, r)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Service) ArchiveLocation(ctx context.Context, locationId, companyId string) error {
	if locationId == "" || companyId == "" {
		return apperror.ErrInvalidCred
	}

	err := s.repo.ArchiveLocation(ctx, locationId, companyId)
	if err != nil {
		return err
	}

	return nil
}

func validType(t LocationType) bool {
	return t == LocationTypeStore || t == LocationTypeWarehouse
}

// repository is the storage Service needs; Repository implements it.
type repository interface {
	// GetLocationById returns an active location or ErrLocationNotFound.
	GetLocationById(ctx context.Context, locationId, companyId string) (*Location, error)
	GetCompanyLocations(ctx context.Context, companyId string) ([]Location, error)
	// CreateLocation returns ErrLocationAlreadyExists for a duplicate active name.
	CreateLocation(ctx context.Context, req *CreateLocationRequest) (*CreateLocationResponse, error)
	// UpdateLocation returns ErrLocationNotFound or ErrLocationAlreadyExists.
	UpdateLocation(ctx context.Context, req *UpdateLocationRequest) (*UpdateLocationResponse, error)
	// ArchiveLocation returns ErrLocationNotFound when nothing was archived.
	ArchiveLocation(ctx context.Context, locationId, companyId string) error
}
