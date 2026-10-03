package auth

import (
	"bazaar/internal/modules/user"
	"bazaar/internal/platform/database"
	"bazaar/internal/platform/support/validator"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type Service struct {
	db        *database.Database
	jwtSecret []byte
}

func New(db *database.Database, jwtSecret []byte) *Service {
	return &Service{db: db, jwtSecret: jwtSecret}
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

func (s *Service) Login(ctx context.Context, rq LoginRequest) (*LoginResponse, error) {
	var repo Repository

	user, err := repo.GetUserByPhone(ctx, s.db.GetDB(), rq.Phone)
	if err != nil {
		return nil, err
	}

	if !validator.CheckPasswordHash(rq.Password, user.PasswordHash) {
		return nil, ErrInvalidCred
	}

	token, err := validator.GenerateToken(user.Id, user.CompanyId, user.RoleId, time.Minute*30, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: token}, nil
}

func (s *Service) GetUserById(ctx context.Context, userId string) (*user.User, error) {
	var repo Repository

	user, err := repo.GetUserById(ctx, s.db.GetDB(), userId)
	if err != nil {
		return nil, err
	}

	return user, nil
}
