package auth

import (
	"bazaar/internal/platform/database"
	"bazaar/internal/platform/support/apperror"
	"bazaar/internal/platform/support/validator"
	"context"
	"time"
)

type Service struct {
	repo      repository
	tx        database.TxManager
	roles     roleService
	jwtSecret []byte
}

func NewService(repo repository, tx database.TxManager, roles roleService, jwtSecret []byte) *Service {
	return &Service{repo: repo, tx: tx, roles: roles, jwtSecret: jwtSecret}
}

func (s *Service) RegisterOwner(ctx context.Context, rq RegisterRequest) (*RegisterResponse, error) {

	if rq.CompanyName == "" || rq.FirstName == "" || rq.Password == "" {
		return nil, apperror.ErrInvalidCred
	}

	if !validator.PhoneValid(rq.Phone) {
		return nil, ErrInvalidPhone
	}

	hash, err := validator.HashPassword(rq.Password)
	if err != nil {
		return nil, err
	}

	var res RegisterResponse
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		companyId, err := s.repo.CreateCompany(ctx, rq.CompanyName)
		if err != nil {
			return err
		}

		roleId, err := s.roles.CreateOwnerRole(ctx, companyId)
		if err != nil {
			return err
		}

		userId, err := s.repo.CreateUser(ctx, companyId, roleId, rq.FirstName, rq.LastName, rq.Phone, hash)
		if err != nil {
			return err
		}

		if err := s.roles.GivePermissions(ctx, userId, roleId); err != nil {
			return err
		}

		res = RegisterResponse{CompanyId: companyId, UserId: userId}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (s *Service) Login(ctx context.Context, rq LoginRequest) (*LoginResponse, error) {
	user, err := s.repo.GetUserByPhone(ctx, rq.Phone)
	if err != nil {
		return nil, err
	}

	if !validator.CheckPasswordHash(rq.Password, user.PasswordHash) {
		return nil, apperror.ErrInvalidCred
	}

	token, err := validator.GenerateToken(user.Id, user.CompanyId, user.UserRoleId, time.Minute*30, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: token}, nil
}

func (s *Service) GetUserById(ctx context.Context, userId string) (*User, error) {
	return s.repo.GetUserById(ctx, userId)
}

// roleService is what registration needs from the role module.
type roleService interface {
	CreateOwnerRole(ctx context.Context, companyId string) (string, error)
	GivePermissions(ctx context.Context, userId, roleId string) error
}

// repository is the storage Service needs; Repository implements it.
type repository interface {
	CreateCompany(ctx context.Context, name string) (string, error)
	// CreateUser returns ErrPhoneTaken when an active user has the phone.
	CreateUser(ctx context.Context, companyId, roleId, firstName, lastName, phone, passwordHash string) (string, error)
	// GetUserByPhone and GetUserById return ErrInvalidCred for a missing user.
	GetUserByPhone(ctx context.Context, phone string) (*User, error)
	GetUserById(ctx context.Context, userId string) (*User, error)
}
