package user

import (
	"bazaar/internal/domain"
	"bazaar/internal/platform/database"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type repository struct {
	db *database.Database
}

var _ domain.UserRepository = (*repository)(nil)

func NewRepository(db *database.Database) domain.UserRepository {
	return &repository{db: db}
}

func (r *repository) GetUserById(ctx context.Context, userId, companyId string) (*domain.User, error) {
	var user domain.User

	err := r.db.Conn(ctx).QueryRow(ctx, `
		SELECT id, company_id, user_role_id, first_name, last_name, phone, password_hash
		FROM users
		WHERE id = $1 and company_id = $2 and deleted_at = 0
	`, userId, companyId).Scan(&user.Id, &user.CompanyId, &user.UserRoleId, &user.FirstName, &user.LastName, &user.Phone, &user.PasswordHash)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *repository) CreateUser(ctx context.Context, req *domain.CreateUserRequest) (*domain.CreateUserResponse, error) {
	var userId string
	err := r.db.Conn(ctx).QueryRow(ctx, `
		INSERT INTO users (id, company_id, user_role_id, first_name, last_name, phone, password_hash)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6)
		RETURNING id
	`, req.CompanyId, req.UserRoleId, req.FirstName, req.LastName, req.Phone, req.PasswordHash).Scan(&userId)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "users_active_phone_unique" {
			return nil, domain.ErrPhoneTaken
		}
		return nil, err
	}

	return &domain.CreateUserResponse{UserID: userId}, nil
}
