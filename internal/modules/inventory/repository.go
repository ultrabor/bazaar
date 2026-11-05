package inventory

import (
	"bazaar/internal/platform/database"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Repository struct {
	db *database.Database
}

func NewRepository(db *database.Database) *Repository {
	return &Repository{db: db}
}

func (r *Repository) LockLocation(ctx context.Context, companyID, locationID string) (bool, error) {
	var archived bool
	err := r.db.Conn(ctx).QueryRow(ctx, `
		SELECT archived FROM locations
		WHERE id = $1 AND company_id = $2
		FOR SHARE
	`, locationID, companyID).Scan(&archived)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, ErrLocationNotFound
	}
	if err != nil {
		return false, fmt.Errorf("lock location: %w", err)
	}
	return archived, nil
}

func (r *Repository) LocationExists(ctx context.Context, companyID, locationID string) error {
	var exists bool
	err := r.db.Conn(ctx).QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM locations WHERE id = $1 AND company_id = $2)
	`, locationID, companyID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check location: %w", err)
	}
	if !exists {
		return ErrLocationNotFound
	}
	return nil
}

func (r *Repository) ListBalances(ctx context.Context, f BalanceFilter) ([]Balance, error) {
	rows, err := r.db.Conn(ctx).Query(ctx, `
		SELECT company_id, location_id, product_id, (quantity * 1000)::bigint, created_at, updated_at
		FROM inventory_balances
		WHERE company_id = $1 AND location_id = $2
		ORDER BY product_id
		LIMIT $3 OFFSET $4
	`, f.CompanyID, f.LocationID, f.Limit, f.Offset)
	if err != nil {
		return nil, fmt.Errorf("list balances: %w", err)
	}
	defer rows.Close()

	balances := make([]Balance, 0)
	for rows.Next() {
		var b Balance
		if err := rows.Scan(&b.CompanyID, &b.LocationID, &b.ProductID, &b.Quantity, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan balance: %w", err)
		}
		balances = append(balances, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list balances: %w", err)
	}
	return balances, nil
}

func (r *Repository) ProductExists(ctx context.Context, companyID, productID string) error {
	var exists bool
	err := r.db.Conn(ctx).QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM products WHERE id = $1 AND company_id = $2)
	`, productID, companyID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check product: %w", err)
	}
	if !exists {
		return ErrProductNotFound
	}
	return nil
}

func (r *Repository) Increase(ctx context.Context, companyID, locationID, productID string, qty int64) error {
	_, err := r.db.Conn(ctx).Exec(ctx, `
		INSERT INTO inventory_balances (company_id, location_id, product_id, quantity)
		VALUES ($1, $2, $3, $4::bigint::numeric / 1000)
		ON CONFLICT (location_id, product_id)
		DO UPDATE SET quantity   = inventory_balances.quantity + EXCLUDED.quantity,
		              updated_at = now()
	`, companyID, locationID, productID, qty)
	if err != nil {
		return fmt.Errorf("increase balance: %w", err)
	}
	return nil
}

func (r *Repository) Decrease(ctx context.Context, companyID, locationID, productID string, qty int64) error {
	tag, err := r.db.Conn(ctx).Exec(ctx, `
		UPDATE inventory_balances
		SET quantity = quantity - $4::bigint::numeric / 1000, updated_at = now()
		WHERE company_id = $1 AND location_id = $2 AND product_id = $3
		  AND quantity >= $4::bigint::numeric / 1000
	`, companyID, locationID, productID, qty)
	if err != nil {
		return fmt.Errorf("decrease balance: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrInsufficientStock
	}
	return nil
}

func (r *Repository) GetBalance(ctx context.Context, companyID, locationID, productID string) (int64, error) {
	var qty int64
	err := r.db.Conn(ctx).QueryRow(ctx, `
		SELECT (quantity * 1000)::bigint FROM inventory_balances
		WHERE company_id = $1 AND location_id = $2 AND product_id = $3
	`, companyID, locationID, productID).Scan(&qty)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("get balance: %w", err)
	}
	return qty, nil
}

func (r *Repository) InsertMovement(ctx context.Context, m Movement) (*Movement, error) {
	err := r.db.Conn(ctx).QueryRow(ctx, `
		INSERT INTO inventory_movements
			(company_id, location_id, product_id, user_id, qty, reason, ref_type, ref_id)
		VALUES ($1, $2, $3, $4, $5::bigint::numeric / 1000, $6, $7, $8)
		RETURNING id, created_at
	`, m.CompanyID, m.LocationID, m.ProductID, m.UserID, m.Qty, m.Reason, m.RefType, m.RefID,
	).Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert movement: %w", err)
	}
	return &m, nil
}

func (r *Repository) ListMovements(ctx context.Context, f MovementFilter) ([]Movement, error) {
	rows, err := r.db.Conn(ctx).Query(ctx, `
		SELECT id, company_id, location_id, product_id, user_id,
		       (qty * 1000)::bigint, reason, ref_type, ref_id, created_at
		FROM inventory_movements
		WHERE company_id = $1 AND location_id = $2 AND product_id = $3
		ORDER BY seq DESC
		LIMIT $4 OFFSET $5
	`, f.CompanyID, f.LocationID, f.ProductID, f.Limit, f.Offset)
	if err != nil {
		return nil, fmt.Errorf("list movements: %w", err)
	}
	defer rows.Close()

	movements := make([]Movement, 0)
	for rows.Next() {
		var m Movement
		if err := rows.Scan(
			&m.ID, &m.CompanyID, &m.LocationID, &m.ProductID, &m.UserID,
			&m.Qty, &m.Reason, &m.RefType, &m.RefID, &m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan movement: %w", err)
		}
		movements = append(movements, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list movements: %w", err)
	}
	return movements, nil
}
