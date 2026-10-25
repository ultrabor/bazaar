package location

import (
	"bazaar/internal/domain"
	"context"
)

type service struct {
	repo domain.LocationRepository
}

var _ domain.LocationService = (*service)(nil)

func NewService(repo domain.LocationRepository) domain.LocationService {
	return &service{repo: repo}
}

func (s *service) GetCompanyLocations(ctx context.Context, companyId string) ([]domain.Location, error) {
	locations, err := s.repo.GetCompanyLocations(ctx, companyId)
	if err != nil {
		return nil, err
	}

	return locations, nil
}

func (s *service) GetLocationById(ctx context.Context, locationId, companyId string) (*domain.Location, error) {
	return s.repo.GetLocationById(ctx, locationId, companyId)
}

func (s *service) CreateLocation(ctx context.Context, r *domain.CreateLocationRequest) (*domain.CreateLocationResponse, error) {

	if r == nil || r.CompanyId == "" || r.Name == "" || r.Address == "" {
		return nil, domain.ErrInvalidCred
	}
	if !validType(r.Type) {
		return nil, domain.ErrInvalidLocationType
	}

	res, err := s.repo.CreateLocation(ctx, r)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *service) UpdateLocation(ctx context.Context, r *domain.UpdateLocationRequest) (*domain.UpdateLocationResponse, error) {

	if r == nil || r.CompanyId == "" || r.LocationId == "" || r.Name == "" || r.Address == "" {
		return nil, domain.ErrInvalidCred
	}
	if !validType(r.Type) {
		return nil, domain.ErrInvalidLocationType
	}

	res, err := s.repo.UpdateLocation(ctx, r)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *service) ArchiveLocation(ctx context.Context, locationId, companyId string) error {
	if locationId == "" || companyId == "" {
		return domain.ErrInvalidCred
	}

	err := s.repo.ArchiveLocation(ctx, locationId, companyId)
	if err != nil {
		return err
	}

	return nil
}

func validType(t domain.LocationType) bool {
	return t == domain.LocationTypeStore || t == domain.LocationTypeWarehouse
}
