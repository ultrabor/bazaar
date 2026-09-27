package auth

import (
	"bazaar/internal/platform/database"
	"bazaar/internal/platform/support/validator"
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

type Service struct {
	db *database.Database
}

func New(db *database.Database) *Service {
	return &Service{db: db}
}

func (s *Service) RegisterOwner(ctx context.Context, rq RegisterRequest) (*RegisterResponse, error) {

	if rq.CompanyName == "" || rq.FirstName == "" || rq.Password == "" {
		return nil, ErrInvalidCred
	}

	if !validator.PhoneValid(rq.Phone) {
		return nil, ErrInvalidPhone
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "users_active_phone_unique" {
			return nil, ErrPhoneTaken
		}
		return nil, err

	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &RegisterResponse{CompanyId: companyId, UserId: userId}, nil
}
