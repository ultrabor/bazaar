package role

import (
	"bazaar/internal/platform/database"
	"context"
)

type Repository struct {
	db *database.Database
}

func NewRepository(db *database.Database) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetList(ctx context.Context, companyId string) ([]GetListRoleResponse, error) {
	var res []GetListRoleResponse

	query := `
		SELECT r.id, r.name, r.is_system
		from user_role
		where company_id = $1 and deleted_at = 0
	`

	rows, err := r.db.GetDB().Query(ctx, query, companyId)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var row GetListRoleResponse

		err = rows.Scan(&row.Id, &row.Name, &row.IsSystem)
		if err != nil {
			return nil, err
		}

		res = append(res, row)
	}

	return res, nil
}
