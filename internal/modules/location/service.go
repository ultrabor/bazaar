package location

import (
	"context"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
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
		return nil, ErrInvalidCred
	}
	if r.Type != TypeStore && r.Type != TypeWarehouse {
		return nil, ErrInvalidType
	}

	res, err := s.repo.CreateLocation(ctx, r)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Service) UpdateLocation(ctx context.Context, r *UpdateLocationRequest) (*UpdateLocationResponse, error) {

	if r == nil || r.CompanyId == "" || r.LocationId == "" || r.Name == "" || r.Address == "" {
		return nil, ErrInvalidCred
	}
	if r.Type != TypeStore && r.Type != TypeWarehouse {
		return nil, ErrInvalidType
	}

	res, err := s.repo.UpdateLocation(ctx, r)
	if err != nil {
		return nil, err
	}

	return res, nil
}
