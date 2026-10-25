package category

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

var _ domain.CategoryRepository = (*repository)(nil)

func NewRepository(db *database.Database) domain.CategoryRepository {
	return &repository{db: db}
}

func (r *repository) CreateCategory(ctx context.Context, rq *domain.CreateCategoryRequest) (*domain.CreateCategoryResponse, error) {
	var categoryId string
	err := r.db.Conn(ctx).QueryRow(ctx, `
		INSERT INTO categories (company_id, name)
		VALUES ($1, $2)
		RETURNING id
	`, rq.CompanyId, rq.Name).Scan(&categoryId)
	if err != nil {
		if isDuplicateName(err) {
			return nil, domain.ErrCategoryAlreadyExists
		}
		return nil, err
	}

	return &domain.CreateCategoryResponse{CategoryId: categoryId}, nil
}

func (r *repository) GetCompanyCategories(ctx context.Context, rq domain.GetCompanyCategoriesRequest) ([]domain.Category, error) {
	rows, err := r.db.Conn(ctx).Query(ctx, `
		SELECT id, company_id, name, archived, created_at, updated_at
		FROM categories
		WHERE company_id = $1 AND archived = false
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`, rq.CompanyId, rq.Limit, (rq.Page-1)*rq.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]domain.Category, 0)
	for rows.Next() {
		var item domain.Category
		if err := rows.Scan(
			&item.Id,
			&item.CompanyId,
			&item.Name,
			&item.Archived,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *repository) GetCategoryById(ctx context.Context, categoryId, companyId string) (*domain.Category, error) {
	var item domain.Category
	err := r.db.Conn(ctx).QueryRow(ctx, `
		SELECT id, company_id, name, archived, created_at, updated_at
		FROM categories
		WHERE id = $1 AND company_id = $2 AND archived = false
	`, categoryId, companyId).Scan(
		&item.Id,
		&item.CompanyId,
		&item.Name,
		&item.Archived,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}

	return &item, nil
}

func (r *repository) UpdateCategory(ctx context.Context, rq *domain.UpdateCategoryRequest) (*domain.UpdateCategoryResponse, error) {
	var categoryId string
	err := r.db.Conn(ctx).QueryRow(ctx, `
		UPDATE categories
		SET name = $1, updated_at = now()
		WHERE id = $2 AND company_id = $3 AND archived = false
		RETURNING id
	`, rq.Name, rq.CategoryId, rq.CompanyId).Scan(&categoryId)
	if err != nil {
		if isDuplicateName(err) {
			return nil, domain.ErrCategoryAlreadyExists
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}

	return &domain.UpdateCategoryResponse{CategoryId: categoryId}, nil
}

func (r *repository) ArchiveCategory(ctx context.Context, categoryId, companyId string) error {
	result, err := r.db.Conn(ctx).Exec(ctx, `
		UPDATE categories
		SET archived = true, updated_at = now()
		WHERE id = $1 AND company_id = $2 AND archived = false
	`, categoryId, companyId)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrCategoryNotFound
	}

	return nil
}

func isDuplicateName(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) &&
		pgErr.Code == "23505" &&
		pgErr.ConstraintName == "categories_company_name_active_unique"
}
