package permission

import (
	"bazaar/internal/platform/database"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Repository struct {
	db *database.Database
}

func NewRepository(db *database.Database) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetListByUserId(ctx context.Context, userId string) ([]Permission, error) {
	var permissions []Permission

	query := `
		select id, label, code from permissions p
		join user_permissions u on u.permission_id = p.id and u.user_id = $1
	`

	rows, err := r.db.GetDB().Query(ctx, query, userId)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var perm Permission

		if err := rows.Scan(
			&perm.Id,
			&perm.Label,
			&perm.Code); err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}

	if err := rows.Err(); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return permissions, nil
}

func (repo *Repository) GetPermissionByCode(ctx context.Context, userId, code string) (*Permission, error) {
	query := `
		select id, label, code from permissions p
		join user_permissions u on u.permission_id = p.id and u.user_id = $1
		where p.code = $2
	`
	per := Permission{}

	err := repo.db.GetDB().QueryRow(ctx, query, userId, code).Scan(&per.Id, &per.Label, &per.Code)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &per, nil
}

func (repo *Repository) SetPermission(ctx context.Context, userId, permissionId string) error {
	query := `
		INSERT INTO user_permissions(user_id, permission_id)
		VALUES ($1, $2)
	`

	_, err := repo.db.GetDB().Exec(ctx, query, userId, permissionId)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "locations_unique_active_name_idx" {
			return ErrPermissionAlreadyExists
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	return nil
}

func (repo *Repository) RemovePermission(ctx context.Context, userId, permissionId string) error {

	query := `
		DELETE FROM user_permissions 
		where user_id = $1 and permissionId = $2
	`

	cmdTag, err := repo.db.GetDB().Exec(ctx, query, userId, permissionId)

	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
