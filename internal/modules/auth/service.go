package auth

import (
	"bazaar/internal/domain"
	"bazaar/internal/platform/support/validator"
	"context"
	"time"
)

type service struct {
	repo      domain.AuthRepository
	tx        domain.TxManager
	roles     domain.RoleService
	jwtSecret []byte
}

var _ domain.AuthService = (*service)(nil)

func NewService(repo domain.AuthRepository, tx domain.TxManager, roles domain.RoleService, jwtSecret []byte) domain.AuthService {
	return &service{repo: repo, tx: tx, roles: roles, jwtSecret: jwtSecret}
}

func (s *service) RegisterOwner(ctx context.Context, rq domain.RegisterRequest) (*domain.RegisterResponse, error) {

	if rq.CompanyName == "" || rq.FirstName == "" || rq.Password == "" {
		return nil, domain.ErrInvalidCred
	}

	if !validator.PhoneValid(rq.Phone) {
		return nil, domain.ErrInvalidPhone
	}

	hash, err := validator.HashPassword(rq.Password)
	if err != nil {
		return nil, err
	}

	var res domain.RegisterResponse
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

		res = domain.RegisterResponse{CompanyId: companyId, UserId: userId}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (s *service) Login(ctx context.Context, rq domain.LoginRequest) (*domain.LoginResponse, error) {
	user, err := s.repo.GetUserByPhone(ctx, rq.Phone)
	if err != nil {
		return nil, err
	}

	if !validator.CheckPasswordHash(rq.Password, user.PasswordHash) {
		return nil, domain.ErrInvalidCred
	}

	token, err := validator.GenerateToken(user.Id, user.CompanyId, user.UserRoleId, time.Minute*30, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &domain.LoginResponse{Token: token}, nil
}

func (s *service) GetUserById(ctx context.Context, userId string) (*domain.User, error) {
	return s.repo.GetUserById(ctx, userId)
}
