package product

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

var _ domain.ProductRepository = (*repository)(nil)

func NewRepository(db *database.Database) domain.ProductRepository {
	return &repository{db: db}
}

func (r *repository) GetProductById(ctx context.Context, productId, companyId string) (*domain.Product, error) {
	var product domain.Product

	err := r.db.Conn(ctx).QueryRow(ctx, `
		SELECT id, company_id, category_id, name, sku, unit, archived, created_at, updated_at
		FROM products
		where id = $1 and company_id = $2 and archived = false
		`, productId, companyId,
	).Scan(
		&product.Id,
		&product.CompanyId,
		&product.CategoryId,
		&product.Name,
		&product.SKU,
		&product.Unit,
		&product.Archived,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, err
	}

	return &product, nil
}

//GetCompanyProducts, CreateProduct, UpdateProduct

func (r *repository) GetCompanyProducts(ctx context.Context, rq domain.GetCompanyProductsRequest) ([]domain.Product, error) {

	query := `
	SELECT id, company_id, category_id, name, sku, unit, archived, created_at, updated_at
		FROM products
		where company_id = $1 and archived = false
			and ($2 = '' or name ilike '%' || $2 || '%' escape E'\\')
			and ($3 = '' or sku = $3)
			and ($4 = '' or unit = $4)
			and ($5 = '' or category_id = nullif($5, '')::uuid)
		ORDER BY created_at desc, id desc
		LIMIT $6 OFFSET $7
		`

	rows, err := r.db.Conn(ctx).Query(
		ctx,
		query,
		rq.CompanyId,
		rq.Name,
		rq.SKU,
		rq.Unit,
		rq.CategoryId,
		rq.Limit,
		(rq.Page-1)*rq.Limit,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]domain.Product, 0)

	for rows.Next() {
		var product domain.Product

		if err := rows.Scan(
			&product.Id,
			&product.CompanyId,
			&product.CategoryId,
			&product.Name,
			&product.SKU,
			&product.Unit,
			&product.Archived,
			&product.CreatedAt,
			&product.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, err
	}

	return products, nil
}

// CreateProduct and UpdateProduct lock the category row, so the category
// cannot be archived between the check and the write.
func (r *repository) CreateProduct(ctx context.Context, rq *domain.CreateProductRequest) (*domain.CreateProductResponse, error) {
	var productId string

	err := r.db.WithinTx(ctx, func(ctx context.Context) error {
		if rq.CategoryId != nil {
			if err := r.lockActiveCategory(ctx, *rq.CategoryId, rq.CompanyId); err != nil {
				return err
			}
		}

		err := r.db.Conn(ctx).QueryRow(ctx, `
	INSERT INTO products (company_id, name, sku, unit, category_id)
	VALUES ($1, $2, $3, $4, $5::uuid)
	RETURNING id
	`, rq.CompanyId, rq.Name, rq.SKU, rq.Unit, rq.CategoryId,
		).Scan(&productId)

		if err != nil {
			if isDuplicateSKU(err) {
				return domain.ErrProductAlreadyExists
			}
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrCategoryNotFound
			}
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &domain.CreateProductResponse{ProductId: productId}, nil
}

func (r *repository) UpdateProduct(ctx context.Context, rq *domain.UpdateProductRequest) (*domain.UpdateProductResponse, error) {
	var productId string

	err := r.db.WithinTx(ctx, func(ctx context.Context) error {
		var currentCategoryId *string
		err := r.db.Conn(ctx).QueryRow(ctx, `
		SELECT category_id
		FROM products
		WHERE id = $1 AND company_id = $2 AND archived = false
		FOR UPDATE
	`, rq.ProductId, rq.CompanyId).Scan(&currentCategoryId)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrProductNotFound
			}
			return err
		}

		keepsCurrentCategory := rq.CategoryId != nil &&
			currentCategoryId != nil &&
			*rq.CategoryId == *currentCategoryId
		if rq.CategoryId != nil && !keepsCurrentCategory {
			if err := r.lockActiveCategory(ctx, *rq.CategoryId, rq.CompanyId); err != nil {
				return err
			}
		}

		err = r.db.Conn(ctx).QueryRow(ctx, `
	UPDATE products
	SET name = $1, sku = $2, unit = $3, category_id = $4::uuid, updated_at = now()
	where id = $5 and company_id = $6 and archived = false
	RETURNING id
	`, rq.Name, rq.SKU, rq.Unit, rq.CategoryId, rq.ProductId, rq.CompanyId,
		).Scan(&productId)

		if err != nil {
			if isDuplicateSKU(err) {
				return domain.ErrProductAlreadyExists
			}
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &domain.UpdateProductResponse{ProductId: productId}, nil
}

// lockActiveCategory must run inside WithinTx: FOR SHARE holds the lock
// only until the transaction ends.
func (r *repository) lockActiveCategory(ctx context.Context, categoryId, companyId string) error {
	var archived bool
	err := r.db.Conn(ctx).QueryRow(ctx, `
		SELECT archived
		FROM categories
		WHERE id = $1 AND company_id = $2
		FOR SHARE
	`, categoryId, companyId).Scan(&archived)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrCategoryNotFound
		}
		return err
	}
	if archived {
		return domain.ErrCategoryNotFound
	}
	return nil
}

func (r *repository) ArchiveProduct(ctx context.Context, productId, companyId string) error {
	res, err := r.db.Conn(ctx).Exec(ctx, `
	UPDATE products
	SET archived = true, updated_at = now()
	where id = $1 and company_id = $2 and archived = false
	`, productId, companyId,
	)

	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return domain.ErrProductNotFound
	}

	return nil
}

func isDuplicateSKU(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) &&
		pgErr.Code == "23505" &&
		pgErr.ConstraintName == "products_company_sku_active_unique"
}
