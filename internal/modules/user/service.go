package user

import (
	"bazaar/internal/platform/support/validator"
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetUserById(ctx context.Context, userId string) (*User, error) {

	return s.repo.GetUserById(ctx, userId)
}

func (s *Service) CreateUser(ctx context.Context, r *CreateUserRequest) (string, error) {

	if r.CompanyId == "" || r.FirstName == "" || r.Password == "" {
		return "", ErrInvalidCred
	}

	if !validator.PhoneValid(r.Phone) {
		return "", ErrInvalidPhone
	}

	hash, err := validator.HashPassword(r.Password)
	if err != nil {
		return "", err
	}

	r.PasswordHash = hash

	tx, err := s.repo.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	userId, err := s.repo.CreateUser(ctx, r)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "users_active_phone_unique" {
			return "", ErrPhoneTaken
		}
		return "", err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return "", err
	}

	return userId, nil
}
