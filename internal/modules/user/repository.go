package user

import (
	"bazaar/internal/platform/database"
	"context"

	"github.com/jackc/pgx/v5"
)

type Repository struct {
	db *database.Database
}

func NewRepository(db *database.Database) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetUserById(ctx context.Context, userId string) (*User, error) {
	var user User

	err := r.db.GetDB().QueryRow(ctx, `
		SELECT id, company_id, user_role_id, first_name, last_name, phone, password_hash
		FROM users
		WHERE id = $1
	`, userId).Scan(&user.Id, &user.CompanyId, &user.UserRoleId, &user.FirstName, &user.LastName, &user.Phone, &user.PasswordHash)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *Repository) CreateUser(ctx context.Context, req *CreateUserRequest) (string, error) {
	var userId string
	err := r.db.GetDB().QueryRow(ctx, `
		INSERT INTO users (id, company_id, user_role_id, first_name, last_name, phone, password_hash)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6)
		RETURNING id
	`, req.CompanyId, req.UserRoleId, req.FirstName, req.LastName, req.Phone, req.PasswordHash).Scan(&userId)

	if err != nil {
		return "", err
	}

	return userId, nil
}
