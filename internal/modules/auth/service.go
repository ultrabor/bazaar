package auth

import (
	"bazaar/internal/platform/database"
	"bazaar/internal/platform/support/apperror"
	"bazaar/internal/platform/support/validator"
	"context"
	"time"
)

type Service struct {
	repo       repository
	tx         database.TxManager
	roles      roleService
	sessionTTL time.Duration
}

func NewService(repo repository, tx database.TxManager, roles roleService, sessionTTL time.Duration) *Service {
	return &Service{repo: repo, tx: tx, roles: roles, sessionTTL: sessionTTL}
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

	token, err := newToken()
	if err != nil {
		return nil, err
	}

	// Only the hash is stored; the raw token goes to the client once.
	_, err = s.repo.CreateSession(ctx, user.Id, hashToken(token), time.Now().Add(s.sessionTTL), rq.UserAgent, rq.Ip)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: token}, nil
}

// Logout revokes the session of the request token.
func (s *Service) Logout(ctx context.Context, sessionId string) error {
	if sessionId == "" {
		return ErrUnauthorized
	}

	return s.repo.RevokeSession(ctx, sessionId)
}

// LoadSession returns the user and session of a raw token. A missing,
// expired or revoked session is apperror.ErrInvalidCred.
func (s *Service) LoadSession(ctx context.Context, token string) (*User, *Session, error) {
	session, err := s.repo.GetActiveSessionByTokenHash(ctx, hashToken(token))
	if err != nil {
		return nil, nil, err
	}

	user, err := s.repo.GetUserById(ctx, session.UserId)
	if err != nil {
		return nil, nil, err
	}

	return user, session, nil
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
	CreateSession(ctx context.Context, userId, tokenHash string, expiresAt time.Time, userAgent, ip string) (string, error)
	// GetActiveSessionByTokenHash returns ErrInvalidCred for a missing,
	// expired or revoked session, or a deleted user.
	GetActiveSessionByTokenHash(ctx context.Context, tokenHash string) (*Session, error)
	// RevokeSession returns ErrInvalidCred when there is nothing to revoke.
	RevokeSession(ctx context.Context, sessionId string) error
}
