package auth

import (
	"bazaar/internal/domain"
	"bazaar/internal/platform/database"
	"bazaar/internal/platform/support/apperror"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type repository struct {
	db *database.Database
}

var _ domain.AuthRepository = (*repository)(nil)

func NewRepository(db *database.Database) domain.AuthRepository {
	return &repository{db: db}
}

func (r *repository) CreateCompany(ctx context.Context, name string) (string, error) {
	var id string

	err := r.db.Conn(ctx).QueryRow(ctx, "INSERT INTO companies(id, label) VALUES(gen_random_uuid(), $1) RETURNING id",
		name).Scan(&id)

	if err != nil {
		return "", apperror.New("failed to create company", err)
	}

	return id, nil
}

func (r *repository) CreateUser(
	ctx context.Context,
	companyID, roleID, firstName, lastName, phone, passwordHash string,
) (string, error) {
	var id string

	err := r.db.Conn(ctx).QueryRow(ctx, `
        INSERT INTO users (
            id, company_id, user_role_id,
            first_name, last_name, phone, password_hash
        )
        VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6)
        RETURNING id
    `, companyID, roleID, firstName, lastName, phone, passwordHash).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "users_active_phone_unique" {
			return "", domain.ErrPhoneTaken
		}
		return "", apperror.New("failed to create user", err)
	}

	return id, nil
}

func (r *repository) GetUserByPhone(ctx context.Context, phone string) (*domain.User, error) {
	var u domain.User

	err := r.db.Conn(ctx).QueryRow(ctx, "SELECT id, company_id, user_role_id, first_name, last_name, phone, password_hash FROM users WHERE phone = $1 and deleted_at = 0", phone).Scan(&u.Id, &u.CompanyId, &u.UserRoleId, &u.FirstName, &u.LastName, &u.Phone, &u.PasswordHash)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidCred
		}
		return nil, apperror.New("failed to get user by phone", err)
	}

	return &u, nil
}

func (r *repository) GetUserById(ctx context.Context, userId string) (*domain.User, error) {
	var u domain.User

	err := r.db.Conn(ctx).QueryRow(ctx, "SELECT id, company_id, user_role_id, first_name, last_name, phone FROM users WHERE id = $1 and deleted_at = 0", userId).Scan(&u.Id, &u.CompanyId, &u.UserRoleId, &u.FirstName, &u.LastName, &u.Phone)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidCred
		}
		return nil, apperror.New("failed to get user by id", err)
	}

	return &u, nil
}
