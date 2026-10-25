package role

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

func (r *repository) CreateSystemRole(ctx context.Context, companyId string) (string, error) {
	var id string

	err := r.db.Conn(ctx).QueryRow(ctx, "INSERT INTO user_roles(id, company_id, name, is_system) VALUES(gen_random_uuid(), $1, 'owner', true) RETURNING id",
		companyId).Scan(&id)

	if err != nil {
		return "", apperror.New("failed to create owner role", err)
	}

	return id, nil
}

func (r *repository) GiveAllPermissions(ctx context.Context, roleId string) error {
	query := `
		INSERT INTO role_permissions(user_role_id, permission_id)
		SELECT $1, p.id
		FROM permissions p
		WHERE p.deleted_at = 0
		ON CONFLICT DO NOTHING;
	`

	_, err := r.db.Conn(ctx).Exec(ctx, query, roleId)

	return err
}

func (r *repository) IsSubsetOfUser(ctx context.Context, roleId, userId string) (bool, error) {
	var ok bool

	err := r.db.Conn(ctx).QueryRow(ctx, `
		SELECT NOT EXISTS (
			SELECT 1
			FROM role_permissions rp
			JOIN permissions p ON p.id = rp.permission_id AND p.deleted_at = 0
			WHERE rp.user_role_id = $1
			  AND NOT EXISTS (
				SELECT 1 FROM user_permissions up
				WHERE up.user_id = $2 AND up.permission_id = rp.permission_id
			  )
		)
	`, roleId, userId).Scan(&ok)

	return ok, err
}

func (r *repository) ListRoles(ctx context.Context, companyId string) ([]domain.RoleSummary, error) {
	rows, err := r.db.Conn(ctx).Query(ctx, `
		SELECT r.id, r.name, r.is_system,
			(SELECT count(*)
			 FROM role_permissions rp
			 JOIN permissions p ON p.id = rp.permission_id AND p.deleted_at = 0
			 WHERE rp.user_role_id = r.id),
			(SELECT count(*)
			 FROM users u
			 WHERE u.user_role_id = r.id AND u.deleted_at = 0)
		FROM user_roles r
		WHERE r.company_id = $1 AND r.deleted_at = 0
		ORDER BY r.is_system DESC, r.created_at, r.id
	`, companyId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]domain.RoleSummary, 0)
	for rows.Next() {
		var role domain.RoleSummary
		if err := rows.Scan(&role.Id, &role.Name, &role.IsSystem, &role.PermissionsCount, &role.UserCount); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, rows.Err()
}

func (r *repository) GetPermissions(ctx context.Context, roleId string) ([]domain.Permission, error) {
	rows, err := r.db.Conn(ctx).Query(ctx, `
		SELECT p.id, p.label, p.code
		FROM role_permissions rp
		JOIN permissions p ON p.id = rp.permission_id AND p.deleted_at = 0
		WHERE rp.user_role_id = $1
		ORDER BY p.code
	`, roleId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := make([]domain.Permission, 0)
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.Id, &p.Label, &p.Code); err != nil {
			return nil, err
		}
		permissions = append(permissions, p)
	}

	return permissions, rows.Err()
}

func (r *repository) CreateRole(ctx context.Context, companyId, name string) (string, error) {
	var id string

	err := r.db.Conn(ctx).QueryRow(ctx, `
		INSERT INTO user_roles (company_id, name)
		VALUES ($1, $2)
		RETURNING id
	`, companyId, name).Scan(&id)
	if err != nil {
		if isDuplicateName(err) {
			return "", domain.ErrRoleAlreadyExists
		}
		return "", err
	}

	return id, nil
}

func (r *repository) RenameRole(ctx context.Context, roleId, companyId, name string) error {
	tag, err := r.db.Conn(ctx).Exec(ctx, `
		UPDATE user_roles
		SET name = $1, updated_at = now()
		WHERE id = $2 AND company_id = $3 AND deleted_at = 0 AND is_system = false
	`, name, roleId, companyId)
	if err != nil {
		if isDuplicateName(err) {
			return domain.ErrRoleAlreadyExists
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrRoleNotFound
	}

	return nil
}

func (r *repository) ArchiveRole(ctx context.Context, roleId, companyId string) error {
	// The conditions repeat the service checks, so the statement stays safe
	// even if a user was assigned the role in between.
	tag, err := r.db.Conn(ctx).Exec(ctx, `
		UPDATE user_roles
		SET deleted_at = extract(epoch from now())::bigint, updated_at = now()
		WHERE id = $1 AND company_id = $2 AND deleted_at = 0 AND is_system = false
		  AND NOT EXISTS (
			SELECT 1 FROM users u WHERE u.user_role_id = $1 AND u.deleted_at = 0
		  )
	`, roleId, companyId)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrRoleNotFound
	}

	return nil
}

func (r *repository) CountActiveUsers(ctx context.Context, roleId string) (int, error) {
	var n int

	err := r.db.Conn(ctx).QueryRow(ctx, `
		SELECT count(*) FROM users WHERE user_role_id = $1 AND deleted_at = 0
	`, roleId).Scan(&n)

	return n, err
}

func (r *repository) UserHasPermission(ctx context.Context, userId, permissionId string) (bool, error) {
	var ok bool

	err := r.db.Conn(ctx).QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM user_permissions up
			JOIN permissions p ON p.id = up.permission_id AND p.deleted_at = 0
			WHERE up.user_id = $1 AND up.permission_id = $2
		)
	`, userId, permissionId).Scan(&ok)
	if err != nil {
		// permissionId comes from the request; a value that is not a uuid
		// cannot be held by anyone.
		if isInvalidUUID(err) {
			return false, nil
		}
		return false, err
	}

	return ok, nil
}

func (r *repository) AddPermission(ctx context.Context, roleId, permissionId string) error {
	_, err := r.db.Conn(ctx).Exec(ctx, `
		INSERT INTO role_permissions (user_role_id, permission_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, roleId, permissionId)

	return err
}

func (r *repository) RemovePermission(ctx context.Context, roleId, permissionId string) error {
	tag, err := r.db.Conn(ctx).Exec(ctx, `
		DELETE FROM role_permissions
		WHERE user_role_id = $1 AND permission_id = $2
	`, roleId, permissionId)
	if err != nil {
		if isInvalidUUID(err) {
			return domain.ErrPermissionNotFound
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrPermissionNotFound
	}

	return nil
}

func (r *repository) GetUserRole(ctx context.Context, userId, companyId string) (*domain.Role, error) {
	var role domain.Role

	err := r.db.Conn(ctx).QueryRow(ctx, `
		SELECT r.id, r.company_id, r.name, r.is_system
		FROM users u
		JOIN user_roles r ON r.id = u.user_role_id
		WHERE u.id = $1 AND u.company_id = $2 AND u.deleted_at = 0
	`, userId, companyId).Scan(&role.Id, &role.CompanyId, &role.Name, &role.IsSystem)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || isInvalidUUID(err) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return &role, nil
}

func (r *repository) SetUserRole(ctx context.Context, userId, roleId string) error {
	return r.db.WithinTx(ctx, func(ctx context.Context) error {
		if _, err := r.db.Conn(ctx).Exec(ctx, `
			UPDATE users SET user_role_id = $2, updated_at = now()
			WHERE id = $1 AND deleted_at = 0
		`, userId, roleId); err != nil {
			return err
		}

		if _, err := r.db.Conn(ctx).Exec(ctx, `
			DELETE FROM user_permissions WHERE user_id = $1
		`, userId); err != nil {
			return err
		}

		return r.GivePermissions(ctx, userId, roleId)
	})
}

func isDuplicateName(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) &&
		pgErr.Code == "23505" &&
		pgErr.ConstraintName == "user_roles_company_name_unique"
}

func isInvalidUUID(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "22P02"
}
