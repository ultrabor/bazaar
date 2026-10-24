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

func (r *Repository) GetUserById(ctx context.Context, userId, companyId string) (*User, error) {
	var user User

	err := r.db.GetDB().QueryRow(ctx, `
		SELECT id, company_id, user_role_id, first_name, last_name, phone, password_hash
		FROM users
		WHERE id = $1 and company_id = $2 and deleted_at = 0
	`, userId, companyId).Scan(&user.Id, &user.CompanyId, &user.UserRoleId, &user.FirstName, &user.LastName, &user.Phone, &user.PasswordHash)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetCompanyRoles(ctx context.Context, companyId string) ([]UserRole, error) {
	rows, err := r.db.GetDB().Query(ctx, `
		SELECT id, company_id, name
		FROM user_roles
		WHERE company_id = $1 and deleted_at = 0
	`, companyId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []UserRole
	for rows.Next() {
		var role UserRole
		if err := rows.Scan(&role.Id, &role.CompanyId, &role.Name); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return roles, nil
}

func (r *Repository) CreateUser(ctx context.Context, tx pgx.Tx, req *CreateUserRequest) (*CreateUserResponse, error) {
	var userId string
	err := tx.QueryRow(ctx, `
		INSERT INTO users (id, company_id, user_role_id, first_name, last_name, phone, password_hash)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6)
		RETURNING id
	`, req.CompanyId, req.UserRoleId, req.FirstName, req.LastName, req.Phone, req.PasswordHash).Scan(&userId)

	if err != nil {
		return nil, err
	}

	return &CreateUserResponse{userId}, nil
}

func (r *Repository) GiveRolePermissions(ctx context.Context, tx pgx.Tx, userId, userRoleId string) error {
	query := `
		INSERT INTO user_permissions(user_id, permission_id)
		SELECT $1, rp.permission_id
		FROM role_permissions rp
		JOIN permissions p ON p.id = rp.permission_id AND p.deleted_at = 0
		WHERE rp.user_role_id = $2
		ON CONFLICT DO NOTHING;
	`

	cmdTag, err := tx.Exec(ctx, query, userId, userRoleId)

	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
