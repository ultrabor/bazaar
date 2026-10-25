package user

import (
	"bazaar/internal/domain"
	"bazaar/internal/platform/support/validator"
	"context"
	"errors"
)

type service struct {
	repo  domain.UserRepository
	tx    domain.TxManager
	roles domain.RoleService
}

var _ domain.UserService = (*service)(nil)

func NewService(repo domain.UserRepository, tx domain.TxManager, roles domain.RoleService) domain.UserService {
	return &service{repo: repo, tx: tx, roles: roles}
}

func (s *service) GetUserById(ctx context.Context, userId, companyId string) (*domain.User, error) {
	return s.repo.GetUserById(ctx, userId, companyId)
}

func (s *service) CreateUser(ctx context.Context, r *domain.CreateUserRequest) (*domain.CreateUserResponse, error) {

	if r.CompanyId == "" || r.FirstName == "" || r.Password == "" {
		return nil, domain.ErrInvalidCred
	}

	if !validator.PhoneValid(r.Phone) {
		return nil, domain.ErrInvalidPhone
	}

	if err := s.roles.CheckAssignable(ctx, r.UserRoleId, r.CompanyId); err != nil {
		if errors.Is(err, domain.ErrRoleNotAssignable) {
			return nil, domain.ErrInvalidRole
		}
		return nil, err
	}

	hash, err := validator.HashPassword(r.Password)
	if err != nil {
		return nil, err
	}

	r.PasswordHash = hash

	var res *domain.CreateUserResponse
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		res, err = s.repo.CreateUser(ctx, r)
		if err != nil {
			return err
		}

		return s.roles.GivePermissions(ctx, res.UserID, r.UserRoleId)
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}
