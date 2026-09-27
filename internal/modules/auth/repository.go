package auth

import (
	"bazaar/internal/platform/support/apperror"
	"context"

	"github.com/jackc/pgx/v5"
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
		return "", apperror.New("failed to create company", err)
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
		return "", apperror.New("failed to create company", err)
	}

	return id, nil
}
