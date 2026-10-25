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

func (s *Service) CreateUser(ctx context.Context, r *CreateUserRequest) (*CreateUserResponse, error) {

	if r.CompanyId == "" || r.FirstName == "" || r.Password == "" {
		return nil, ErrInvalidCred
	}

	if !validator.PhoneValid(r.Phone) {
		return nil, ErrInvalidPhone
	}

	roles, err := s.repo.GetCompanyRoles(ctx, r.CompanyId)
	if err != nil {
		return nil, err
	}

	ok := false
	for _, role := range roles {
		if role.Id == r.UserRoleId {
			ok = true
			break
		}
	}

	if !ok {
		return nil, ErrInvalidRole
	}

	hash, err := validator.HashPassword(r.Password)
	if err != nil {
		return nil, err
	}

	r.PasswordHash = hash

	tx, err := s.repo.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	res, err := s.repo.CreateUser(ctx, tx, r)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "users_active_phone_unique" {
			return nil, ErrPhoneTaken
		}
		return nil, err
	}

	err = s.repo.GiveRolePermissions(ctx, tx, res.UserID, r.UserRoleId)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)

	if err != nil {
		return nil, err
	}

	return res, nil
}
