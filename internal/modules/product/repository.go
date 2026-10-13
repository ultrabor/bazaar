package product

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

func (r *Repository) GetProductById(ctx context.Context, productId, companyId string) (*Product, error) {
	var product Product

	err := r.db.GetDB().QueryRow(ctx, `
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
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &product, nil
}

//GetCompanyProducts, CreateProduct, UpdateProduct

func (r *Repository) GetCompanyProducts(ctx context.Context, rq GetCompanyProductsRequest) ([]Product, error) {

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

	rows, err := r.db.GetDB().Query(
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

	products := make([]Product, 0)

	for rows.Next() {
		var product Product

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
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return products, nil
}

func (r *Repository) CreateProduct(ctx context.Context, rq *CreateProductRequest) (*CreateProductResponse, error) {
	var productId string

	err := r.db.GetDB().QueryRow(ctx, `
	INSERT INTO products (company_id, name, sku, unit, category_id)
	SELECT $1, $2, $3, $4, $5::uuid
	WHERE $5::uuid IS NULL OR EXISTS (
		SELECT 1 FROM categories
		WHERE id = $5::uuid AND company_id = $1 AND archived = false
	)
	RETURNING id
	`, rq.CompanyId, rq.Name, rq.SKU, rq.Unit, rq.CategoryId,
	).Scan(&productId)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "products_company_sku_active_unique" {
			return nil, ErrProductAlreadyExists
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}

	return &CreateProductResponse{ProductId: productId}, nil
}

func (r *Repository) UpdateProduct(ctx context.Context, rq *UpdateProductRequest) (*UpdateProductResponse, error) {
	var productId string

	err := r.db.GetDB().QueryRow(ctx, `
	UPDATE products AS product
	SET name = $1, sku = $2, unit = $3, category_id = $4::uuid, updated_at = now()
	where id = $5 and company_id = $6 and archived = false
		and (
			$4::uuid is null
			or product.category_id = $4::uuid
			or exists (
				select 1 from categories
				where id = $4::uuid and company_id = $6 and archived = false
			)
		)
	RETURNING id
	`, rq.Name, rq.SKU, rq.Unit, rq.CategoryId, rq.ProductId, rq.CompanyId,
	).Scan(&productId)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "products_company_sku_active_unique" {
			return nil, ErrProductAlreadyExists
		}
		if errors.Is(err, pgx.ErrNoRows) {
			var productExists bool
			lookupErr := r.db.GetDB().QueryRow(ctx, `
				SELECT EXISTS (
					SELECT 1 FROM products
					WHERE id = $1 AND company_id = $2 AND archived = false
				)
			`, rq.ProductId, rq.CompanyId).Scan(&productExists)
			if lookupErr != nil {
				return nil, lookupErr
			}
			if productExists {
				return nil, ErrCategoryNotFound
			}
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &UpdateProductResponse{ProductId: productId}, nil
}

func (r *Repository) ArchiveProduct(ctx context.Context, productId, companyId string) error {
	res, err := r.db.GetDB().Exec(ctx, `
	UPDATE products 
	SET archived = true, updated_at = now()	
	where id = $1 and company_id = $2 and archived = false
	`, productId, companyId,
	)

	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
