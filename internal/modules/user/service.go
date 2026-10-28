package user

import (
	"bazaar/internal/modules/role"
	"bazaar/internal/platform/database"
	"bazaar/internal/platform/support/apperror"
	"bazaar/internal/platform/support/validator"
	"context"
	"errors"
)

type Service struct {
	repo  repository
	tx    database.TxManager
	roles roleService
}

func NewService(repo repository, tx database.TxManager, roles roleService) *Service {
	return &Service{repo: repo, tx: tx, roles: roles}
}

func (s *Service) GetUserById(ctx context.Context, userId, companyId string) (*User, error) {
	return s.repo.GetUserById(ctx, userId, companyId)
}

func (s *Service) CreateUser(ctx context.Context, r *CreateUserRequest) (*CreateUserResponse, error) {

	if r.CompanyId == "" || r.FirstName == "" || r.Password == "" {
		return nil, apperror.ErrInvalidCred
	}

	if !validator.PhoneValid(r.Phone) {
		return nil, ErrInvalidPhone
	}

	if err := s.roles.CheckAssignable(ctx, r.UserRoleId, r.CompanyId); err != nil {
		if errors.Is(err, role.ErrRoleNotAssignable) {
			return nil, ErrInvalidRole
		}
		return nil, err
	}

	if err := s.roles.CheckGrantable(ctx, r.UserRoleId, r.CreatorId); err != nil {
		return nil, err
	}

	hash, err := validator.HashPassword(r.Password)
	if err != nil {
		return nil, err
	}

	r.PasswordHash = hash

	var res *CreateUserResponse
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

// roleService is what CreateUser needs from the role module.
type roleService interface {
	CheckAssignable(ctx context.Context, roleId, companyId string) error
	CheckGrantable(ctx context.Context, roleId, grantorId string) error
	GivePermissions(ctx context.Context, userId, roleId string) error
}

// repository is the storage Service needs; Repository implements it.
type repository interface {
	GetUserById(ctx context.Context, userId, companyId string) (*User, error)
	// CreateUser expects req.PasswordHash to be set. Returns ErrPhoneTaken
	// when an active user with the same phone exists.
	CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error)
}
