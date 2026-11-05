package inventory_test

import (
	"bazaar/internal/modules/inventory"
	"bazaar/internal/platform/config"
	"bazaar/internal/platform/database"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"os"
	"sync"
	"testing"
)

// errRollback ends a test transaction, so nothing stays in the database.
var errRollback = errors.New("rollback test data")

// The tests run the ledger SQL against a real database. They need migrated
// Postgres and run only with BAZAAR_INTEGRATION=1, e.g.
// `BAZAAR_INTEGRATION=1 make test` (make exports POSTGRES_* from .env).
func connect(t *testing.T) *database.Database {
	t.Helper()
	if os.Getenv("BAZAAR_INTEGRATION") == "" {
		t.Skip("set BAZAAR_INTEGRATION=1 to run against Postgres")
	}
	db, err := database.New(context.Background(), config.EnvLoad(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(db.Close)
	return db
}

// inRolledBackTx runs fn in a transaction that is always rolled back.
func inRolledBackTx(t *testing.T, db *database.Database, fn func(ctx context.Context, f *fixture)) {
	t.Helper()
	err := db.WithinTx(context.Background(), func(ctx context.Context) error {
		fn(ctx, newFixture(ctx, t, db))
		return errRollback
	})
	if !errors.Is(err, errRollback) {
		t.Fatalf("WithinTx: %v", err)
	}
}

func TestLedgerOnDatabase(t *testing.T) {
	db := connect(t)
	svc := inventory.NewService(inventory.NewRepository(db), db)

	t.Run("increase creates then adds to the balance", func(t *testing.T) {
		inRolledBackTx(t, db, func(ctx context.Context, f *fixture) {
			f.increase(svc, 1500)
			m := f.increase(svc, 250)

			f.wantBalance(svc, 1750)
			f.wantMovements(svc, 250, 1500)
			if m.RefType == nil || *m.RefType != inventory.RefTypePurchase {
				t.Errorf("ref_type = %v, want purchase", m.RefType)
			}
		})
	})

	t.Run("decrease subtracts and records a negative movement", func(t *testing.T) {
		inRolledBackTx(t, db, func(ctx context.Context, f *fixture) {
			f.increase(svc, 2000)

			if _, err := svc.DecreaseBalance(ctx, f.change(inventory.ReasonSale, 500)); err != nil {
				t.Fatalf("DecreaseBalance: %v", err)
			}

			f.wantBalance(svc, 1500)
			f.wantMovements(svc, -500, 2000)
		})
	})

	t.Run("decrease below zero is rejected and changes nothing", func(t *testing.T) {
		inRolledBackTx(t, db, func(ctx context.Context, f *fixture) {
			f.increase(svc, 1000)

			_, err := svc.DecreaseBalance(ctx, f.change(inventory.ReasonSale, 1001))

			if !errors.Is(err, inventory.ErrInsufficientStock) {
				t.Fatalf("err = %v, want ErrInsufficientStock", err)
			}
			f.wantBalance(svc, 1000)
			f.wantMovements(svc, 1000)
		})
	})

	t.Run("decrease without a balance row is rejected", func(t *testing.T) {
		inRolledBackTx(t, db, func(ctx context.Context, f *fixture) {
			_, err := svc.DecreaseBalance(ctx, f.change(inventory.ReasonWriteOff, 1))

			if !errors.Is(err, inventory.ErrInsufficientStock) {
				t.Fatalf("err = %v, want ErrInsufficientStock", err)
			}
		})
	})

	t.Run("check stock", func(t *testing.T) {
		inRolledBackTx(t, db, func(ctx context.Context, f *fixture) {
			f.increase(svc, 1500)

			if err := svc.CheckStock(ctx, f.company, f.location, f.product, 1500); err != nil {
				t.Errorf("CheckStock(1500) = %v, want nil", err)
			}
			if err := svc.CheckStock(ctx, f.company, f.location, f.product, 1501); !errors.Is(err, inventory.ErrInsufficientStock) {
				t.Errorf("CheckStock(1501) = %v, want ErrInsufficientStock", err)
			}
		})
	})

	t.Run("archived location rejects changes but stays readable", func(t *testing.T) {
		inRolledBackTx(t, db, func(ctx context.Context, f *fixture) {
			f.increase(svc, 3000)
			f.exec(`UPDATE locations SET archived = true WHERE id = $1`, f.location)

			if _, err := svc.IncreaseBalance(ctx, f.change(inventory.ReasonPurchase, 1)); !errors.Is(err, inventory.ErrLocationArchived) {
				t.Errorf("increase err = %v, want ErrLocationArchived", err)
			}
			if _, err := svc.DecreaseBalance(ctx, f.change(inventory.ReasonWriteOff, 1)); !errors.Is(err, inventory.ErrLocationArchived) {
				t.Errorf("decrease err = %v, want ErrLocationArchived", err)
			}

			f.wantBalance(svc, 3000)
			f.wantMovements(svc, 3000)
			if err := svc.CheckStock(ctx, f.company, f.location, f.product, 3000); err != nil {
				t.Errorf("CheckStock on archived location = %v, want nil", err)
			}
			balances, err := svc.ListBalances(ctx, inventory.BalanceFilter{CompanyID: f.company, LocationID: f.location})
			if err != nil || len(balances) != 1 || balances[0].Quantity != 3000 {
				t.Errorf("ListBalances on archived location = %+v, %v; want one row of 3000", balances, err)
			}
		})
	})

	t.Run("archived product rejects purchase but can leave stock", func(t *testing.T) {
		inRolledBackTx(t, db, func(ctx context.Context, f *fixture) {
			f.increase(svc, 2000)
			f.exec(`UPDATE products SET archived = true WHERE id = $1`, f.product)

			if _, err := svc.IncreaseBalance(ctx, f.change(inventory.ReasonPurchase, 1)); !errors.Is(err, inventory.ErrProductArchived) {
				t.Errorf("purchase err = %v, want ErrProductArchived", err)
			}
			if _, err := svc.DecreaseBalance(ctx, f.change(inventory.ReasonSale, 500)); err != nil {
				t.Errorf("sale of archived product: %v", err)
			}
			if _, err := svc.IncreaseBalance(ctx, f.change(inventory.ReasonReturn, 100)); err != nil {
				t.Errorf("return of archived product: %v", err)
			}

			f.wantBalance(svc, 1600)
			f.wantMovements(svc, 100, -500, 2000)
		})
	})

	t.Run("another company cannot see or change the stock", func(t *testing.T) {
		inRolledBackTx(t, db, func(ctx context.Context, f *fixture) {
			f.increase(svc, 1000)
			other := newFixture(ctx, t, db)

			c := f.change(inventory.ReasonPurchase, 1)
			c.CompanyID, c.UserID = other.company, other.user
			if _, err := svc.IncreaseBalance(ctx, c); !errors.Is(err, inventory.ErrLocationNotFound) {
				t.Errorf("foreign location err = %v, want ErrLocationNotFound", err)
			}

			c = other.change(inventory.ReasonPurchase, 1)
			c.ProductID = f.product
			if _, err := svc.IncreaseBalance(ctx, c); !errors.Is(err, inventory.ErrProductNotFound) {
				t.Errorf("foreign product err = %v, want ErrProductNotFound", err)
			}

			got, err := svc.GetBalance(ctx, other.company, f.location, f.product)
			if err != nil || got != 0 {
				t.Errorf("foreign balance = %d, %v; want 0, nil", got, err)
			}
			if _, err := svc.ListBalances(ctx, inventory.BalanceFilter{CompanyID: other.company, LocationID: f.location}); !errors.Is(err, inventory.ErrLocationNotFound) {
				t.Errorf("foreign ListBalances err = %v, want ErrLocationNotFound", err)
			}
		})
	})

	t.Run("movements are append-only", func(t *testing.T) {
		for _, sql := range []string{
			`UPDATE inventory_movements SET qty = 1 WHERE location_id = $1`,
			`DELETE FROM inventory_movements WHERE location_id = $1`,
		} {
			inRolledBackTx(t, db, func(ctx context.Context, f *fixture) {
				f.increase(svc, 1000)

				_, err := db.Conn(ctx).Exec(ctx, sql, f.location)

				if err == nil {
					t.Errorf("%s succeeded, want append-only error", sql)
				}
			})
		}
	})
}

// TestConcurrentDecreaseNeverOverdraws commits its data, because each
// goroutine needs its own transaction. Cleanup removes it.
func TestConcurrentDecreaseNeverOverdraws(t *testing.T) {
	db := connect(t)
	svc := inventory.NewService(inventory.NewRepository(db), db)
	ctx := context.Background()
	f := newFixture(ctx, t, db)
	t.Cleanup(func() { f.purge() })

	const stock, workers = 5, 10
	f.increase(svc, stock*inventory.QtyScale)

	var wg sync.WaitGroup
	errs := make(chan error, workers)
	for range workers {
		wg.Go(func() {
			_, err := svc.DecreaseBalance(ctx, f.change(inventory.ReasonSale, inventory.QtyScale))
			errs <- err
		})
	}
	wg.Wait()
	close(errs)

	succeeded := 0
	for err := range errs {
		switch {
		case err == nil:
			succeeded++
		case !errors.Is(err, inventory.ErrInsufficientStock):
			t.Errorf("unexpected error: %v", err)
		}
	}
	if succeeded != stock {
		t.Errorf("successful decreases = %d, want %d", succeeded, stock)
	}
	f.wantBalance(svc, 0)
}

type fixture struct {
	ctx      context.Context
	t        *testing.T
	db       *database.Database
	company  string
	user     string
	location string
	product  string
}

func newFixture(ctx context.Context, t *testing.T, db *database.Database) *fixture {
	t.Helper()
	f := &fixture{ctx: ctx, t: t, db: db}
	var role string
	f.scan(&f.company, `INSERT INTO companies(id, label) VALUES (gen_random_uuid(), 'inventory test') RETURNING id`)
	f.scan(&role, `INSERT INTO user_roles(company_id, name) VALUES ($1, 'Stock keeper') RETURNING id`, f.company)
	f.scan(&f.user, `
		INSERT INTO users(id, company_id, user_role_id, first_name, phone, password_hash)
		VALUES (gen_random_uuid(), $1, $2, 'Test', $3, 'x') RETURNING id
	`, f.company, role, fmt.Sprintf("+99801%07d", rand.IntN(10_000_000)))
	f.scan(&f.location, `
		INSERT INTO locations(id, company_id, name, address, type)
		VALUES (gen_random_uuid(), $1, 'Main warehouse', 'Test street', 'warehouse') RETURNING id
	`, f.company)
	f.scan(&f.product, `
		INSERT INTO products(company_id, name, sku, unit) VALUES ($1, 'Rice', 'RICE-1', 'kg') RETURNING id
	`, f.company)
	return f
}

func (f *fixture) change(reason inventory.Reason, qty int64) inventory.Change {
	return inventory.Change{
		CompanyID:  f.company,
		LocationID: f.location,
		ProductID:  f.product,
		UserID:     f.user,
		Qty:        qty,
		Reason:     reason,
	}
}

func (f *fixture) increase(svc *inventory.Service, qty int64) *inventory.Movement {
	f.t.Helper()
	c := f.change(inventory.ReasonPurchase, qty)
	c.RefType, c.RefID = inventory.RefTypePurchase, f.company
	m, err := svc.IncreaseBalance(f.ctx, c)
	if err != nil {
		f.t.Fatalf("IncreaseBalance(%d): %v", qty, err)
	}
	return m
}

func (f *fixture) wantBalance(svc *inventory.Service, want int64) {
	f.t.Helper()
	got, err := svc.GetBalance(f.ctx, f.company, f.location, f.product)
	if err != nil {
		f.t.Fatalf("GetBalance: %v", err)
	}
	if got != want {
		f.t.Errorf("balance = %d, want %d", got, want)
	}
}

// wantMovements checks the ledger quantities, newest first.
func (f *fixture) wantMovements(svc *inventory.Service, want ...int64) {
	f.t.Helper()
	movements, err := svc.ListMovements(f.ctx, inventory.MovementFilter{
		CompanyID: f.company, LocationID: f.location, ProductID: f.product,
	})
	if err != nil {
		f.t.Fatalf("ListMovements: %v", err)
	}
	got := make([]int64, 0, len(movements))
	for _, m := range movements {
		got = append(got, m.Qty)
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		f.t.Errorf("movements = %v, want %v", got, want)
	}
}

func (f *fixture) scan(dst *string, sql string, args ...any) {
	f.t.Helper()
	if err := f.db.Conn(f.ctx).QueryRow(f.ctx, sql, args...).Scan(dst); err != nil {
		f.t.Fatalf("%s: %v", sql, err)
	}
}

func (f *fixture) exec(sql string, args ...any) {
	f.t.Helper()
	if _, err := f.db.Conn(f.ctx).Exec(f.ctx, sql, args...); err != nil {
		f.t.Fatalf("%s: %v", sql, err)
	}
}

// purge deletes committed fixture data. The ledger is append-only, so it
// disables the guard trigger inside its own transaction and enables it
// again before commit. ALTER TABLE is transactional and needs only table
// ownership, which the migrating user has; on a rollback the trigger
// stays enabled.
func (f *fixture) purge() {
	err := f.db.WithinTx(f.ctx, func(ctx context.Context) error {
		conn := f.db.Conn(ctx)
		if _, err := conn.Exec(ctx, `ALTER TABLE inventory_movements DISABLE TRIGGER inventory_movements_no_update_delete`); err != nil {
			return fmt.Errorf("disable append-only trigger: %w", err)
		}
		for _, sql := range []string{
			`DELETE FROM inventory_movements WHERE company_id = $1`,
			`DELETE FROM inventory_balances WHERE company_id = $1`,
			`DELETE FROM products WHERE company_id = $1`,
			`DELETE FROM locations WHERE company_id = $1`,
			`DELETE FROM users WHERE company_id = $1`,
			`DELETE FROM user_roles WHERE company_id = $1`,
			`DELETE FROM companies WHERE id = $1`,
		} {
			if _, err := conn.Exec(ctx, sql, f.company); err != nil {
				return fmt.Errorf("%s: %w", sql, err)
			}
		}
		if _, err := conn.Exec(ctx, `ALTER TABLE inventory_movements ENABLE TRIGGER inventory_movements_no_update_delete`); err != nil {
			return fmt.Errorf("enable append-only trigger: %w", err)
		}
		return nil
	})
	if err != nil {
		f.t.Errorf("cleanup left test company %s: %v", f.company, err)
	}
}
