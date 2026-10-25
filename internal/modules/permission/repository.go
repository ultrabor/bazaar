package permission

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

var _ domain.PermissionRepository = (*repository)(nil)

func NewRepository(db *database.Database) domain.PermissionRepository {
	return &repository{db: db}
}

func (r *repository) GetListByUserId(ctx context.Context, userId string) ([]domain.Permission, error) {
	var permissions []domain.Permission

	query := `
		select id, label, code from permissions p
		join user_permissions u on u.permission_id = p.id and u.user_id = $1
	`

	rows, err := r.db.Conn(ctx).Query(ctx, query, userId)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var perm domain.Permission

		if err := rows.Scan(
			&perm.Id,
			&perm.Label,
			&perm.Code); err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}

	if err := rows.Err(); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPermissionNotFound
		}
		return nil, err
	}

	return permissions, nil
}

func (repo *repository) GetPermissionByCode(ctx context.Context, userId, code string) (*domain.Permission, error) {
	query := `
		select id, label, code from permissions p
		join user_permissions u on u.permission_id = p.id and u.user_id = $1
		where p.code = $2
	`
	per := domain.Permission{}

	err := repo.db.Conn(ctx).QueryRow(ctx, query, userId, code).Scan(&per.Id, &per.Label, &per.Code)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPermissionNotFound
		}
		return nil, err
	}

	return &per, nil
}

func (repo *repository) SetPermission(ctx context.Context, userId, permissionId string) error {
	query := `
		INSERT INTO user_permissions(user_id, permission_id)
		VALUES ($1, $2)
	`

	_, err := repo.db.Conn(ctx).Exec(ctx, query, userId, permissionId)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "locations_unique_active_name_idx" {
			return domain.ErrPermissionAlreadyExists
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrPermissionNotFound
		}
		return err
	}

	return nil
}

func (repo *repository) RemovePermission(ctx context.Context, userId, permissionId string) error {

	query := `
		DELETE FROM user_permissions 
		where user_id = $1 and permission_id = $2
	`

	cmdTag, err := repo.db.Conn(ctx).Exec(ctx, query, userId, permissionId)

	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrPermissionNotFound
	}

	return nil
}
