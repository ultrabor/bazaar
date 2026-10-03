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

func (s *Service) GetUserById(ctx context.Context, userId, companyId string) (*User, error) {
	return s.repo.GetUserById(ctx, userId, companyId)
}

func (s *Service) CreateUser(ctx context.Context, r *CreateUserRequest) (string, error) {

	if r.CompanyId == "" || r.FirstName == "" || r.Password == "" {
		return "", ErrInvalidCred
	}

	if !validator.PhoneValid(r.Phone) {
		return "", ErrInvalidPhone
	}

	roleIds, err := s.repo.GetCompanyRoles(ctx, r.CompanyId)
	if err != nil {
		return "", err
	}

	ok := false
	for _, roleId := range roleIds {
		if roleId == r.UserRoleId {
			ok = true
			break
		}
	}

	if !ok {
		return "", errors.New("invalid role for company")
	}

	hash, err := validator.HashPassword(r.Password)
	if err != nil {
		return "", err
	}

	r.PasswordHash = hash

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

	return userId, nil
}
