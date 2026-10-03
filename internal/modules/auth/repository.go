package auth

import (
	"bazaar/internal/modules/user"
	"bazaar/internal/platform/support/apperror"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{}

func (r *Repository) CreateCompany(ctx context.Context, tx pgx.Tx, name string) (string, error) {
	var id string

	err := tx.QueryRow(ctx, "INSERT INTO companies(id, label) VALUES(gen_random_uuid(), $1) RETURNING id",
		name).Scan(&id)

	if err != nil {
		return "", apperror.New("failed to create company", err)
	}

	return id, nil
}

func (r *Repository) CreateOwnerRole(ctx context.Context, tx pgx.Tx, companyId string) (string, error) {
	var id string

	err := tx.QueryRow(ctx, "INSERT INTO user_roles(id, company_id, name) VALUES(gen_random_uuid(), $1, 'owner') RETURNING id",
		companyId).Scan(&id)

	if err != nil {
		return "", apperror.New("failed to create owner role", err)
	}

	return id, nil
}

func (r *Repository) CreateUser(
	ctx context.Context,
	tx pgx.Tx,
	companyID, roleID, firstName, lastName, phone, passwordHash string,
) (string, error) {
	var id string

	err := tx.QueryRow(ctx, `
        INSERT INTO users (
            id, company_id, user_role_id,
            first_name, last_name, phone, password_hash
        )
        VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6)
        RETURNING id
    `, companyID, roleID, firstName, lastName, phone, passwordHash).Scan(&id)

	if err != nil {
		return "", apperror.New("failed to create user", err)
	}

	return id, nil
}

func (r *Repository) GetUserByPhone(ctx context.Context, db *pgxpool.Pool, phone string) (*user.User, error) {
	var u user.User

	err := db.QueryRow(ctx, "SELECT id, company_id, user_role_id, first_name, last_name, phone, password_hash FROM users WHERE phone = $1 and deleted_at = 0", phone).Scan(&u.Id, &u.CompanyId, &u.UserRoleId, &u.FirstName, &u.LastName, &u.Phone, &u.PasswordHash)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrInvalidCred
		}
		return nil, apperror.New("failed to get user by phone", err)
	}

	return &u, nil
}

func (r *Repository) GetUserById(ctx context.Context, db *pgxpool.Pool, userId string) (*user.User, error) {
	var u user.User

	err := db.QueryRow(ctx, "SELECT id, company_id, user_role_id, first_name, last_name, phone FROM users WHERE id = $1 and deleted_at = 0", userId).Scan(&u.Id, &u.CompanyId, &u.UserRoleId, &u.FirstName, &u.LastName, &u.Phone)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrInvalidCred
		}
		return nil, apperror.New("failed to get user by id", err)
	}

	return &u, nil
}
