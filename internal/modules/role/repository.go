package role

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

var _ domain.RoleRepository = (*repository)(nil)

func NewRepository(db *database.Database) domain.RoleRepository {
	return &repository{db: db}
}

func (r *repository) GetById(ctx context.Context, roleId, companyId string) (*domain.Role, error) {
	var role domain.Role

	err := r.db.Conn(ctx).QueryRow(ctx, `
		SELECT id, company_id, name, is_system
		FROM user_roles
		WHERE id = $1 and company_id = $2 and deleted_at = 0
	`, roleId, companyId).Scan(&role.Id, &role.CompanyId, &role.Name, &role.IsSystem)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRoleNotFound
		}
		// roleId comes from the request body; a value that is not a uuid
		// cannot match any role.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "22P02" {
			return nil, domain.ErrRoleNotFound
		}
		return nil, err
	}

	return &role, nil
}

func (r *repository) GetCompanyRoles(ctx context.Context, companyId string) ([]domain.Role, error) {
	rows, err := r.db.Conn(ctx).Query(ctx, `
		SELECT id, company_id, name, is_system
		FROM user_roles
		WHERE company_id = $1 and deleted_at = 0 and is_system = false
	`, companyId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.Id, &role.CompanyId, &role.Name, &role.IsSystem); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *repository) GivePermissions(ctx context.Context, userId, roleId string) error {
	query := `
		INSERT INTO user_permissions(user_id, permission_id)
		SELECT $1, rp.permission_id
		FROM role_permissions rp
		JOIN permissions p ON p.id = rp.permission_id AND p.deleted_at = 0
		WHERE rp.user_role_id = $2
		ON CONFLICT DO NOTHING;
	`

	_, err := r.db.Conn(ctx).Exec(ctx, query, userId, roleId)

	return err
}
