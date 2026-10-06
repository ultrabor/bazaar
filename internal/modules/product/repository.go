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
		SELECT id, company_id, name, sku, unit, archived
		FROM products
		where id = $1 and company_id = $2 and archived = false
		`, productId, companyId,
	).Scan(
		&product.Id,
		&product.CompanyId,
		&product.Name,
		&product.SKU,
		&product.Unit,
		&product.Archived,
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

func (r *Repository) GetCompanyProducts(ctx context.Context, companyId string) ([]Product, error) {
	rows, err := r.db.GetDB().Query(ctx, `
	SELECT id, company_id, name, sku, unit, archived
		FROM products
		where company_id = $1 and archived = false
		`, companyId,
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
			&product.Name,
			&product.SKU,
			&product.Unit,
			&product.Archived,
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
	INSERT INTO products (id, company_id, name, sku, unit)
	VALUES(gen_random(), $1, $2, $3, $4)
	RETURNING id
	`, rq.CompanyId, rq.Name, rq.SKU, rq.Unit,
	).Scan(&productId)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "products_company_sku_active_unique" {
			return nil, ErrProductAlreadyExists
		}
		return nil, err
	}

	return &CreateProductResponse{ProductId: productId}, nil
}

func (r *Repository) UpdateProduct(ctx context.Context, rq *UpdateProductRequest) (*UpdateProductResponse, error) {
	var productId string

	err := r.db.GetDB().QueryRow(ctx, `
	UPDATE products
	SET name = $1, sku = $2, unit = $3, updated_at = now()
	where id = $4, and company_id = $5 and archived = false
	RETURNING id
	`, rq.Name, rq.SKU, rq.Unit, rq.ProductId, rq.CompanyId,
	).Scan(&productId)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "products_company_sku_active_unique" {
			return nil, ErrProductAlreadyExists
		}
		if errors.Is(err, pgx.ErrNoRows) {
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
	where id = $1 and companyId = $2 and archived = false
	`,
	)

	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
