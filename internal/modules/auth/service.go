package auth

import (
	"bazaar/internal/platform/database"
	"bazaar/internal/platform/support/apperror"
	"bazaar/internal/platform/support/validator"
	"context"
)

type Service struct {
	db *database.Database
}

func New(db *database.Database) *Service {
	return &Service{db: db}
}

func (s *Service) RegisterOwner(ctx context.Context, rq RegisterRequest) (*RegisterResponse, error) {

	if !validator.PhoneValid(rq.Phone) {
		return nil, apperror.New("phone is not valid", nil)
	}

	hash, err := validator.HashPassword(rq.Password)
	if err != nil {
		return nil, err
	}

	repo := Repository{}

	tx, err := s.db.Begin(ctx)

	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	companyId, err := repo.CreateCompany(ctx, tx, rq.CompanyName)
	if err != nil {
		return nil, err
	}
	roleId, err := repo.CreateOwnerRole(ctx, tx, companyId)
	if err != nil {
		return nil, err
	}

	userId, err := repo.CreateUser(ctx, tx, companyId, roleId, rq.FirstName, rq.LastName, rq.Phone, hash)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &RegisterResponse{CompanyId: companyId, UserId: userId}, nil
}
