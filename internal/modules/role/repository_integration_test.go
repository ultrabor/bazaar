package role_test

import (
	"bazaar/internal/domain"
	"bazaar/internal/modules/role"
	"bazaar/internal/platform/config"
	"bazaar/internal/platform/database"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"os"
	"slices"
	"testing"
)

// errRollback ends the test transaction, so nothing stays in the database.
var errRollback = errors.New("rollback test data")

// TestRolePermissionSyncOnDatabase runs the role SQL against a real
// database. It needs migrated Postgres and runs only with
// BAZAAR_INTEGRATION=1, e.g. `BAZAAR_INTEGRATION=1 make test`
// (make exports POSTGRES_* from .env).
func TestRolePermissionSyncOnDatabase(t *testing.T) {
	if os.Getenv("BAZAAR_INTEGRATION") == "" {
		t.Skip("set BAZAAR_INTEGRATION=1 to run against Postgres")
	}

	ctx := context.Background()
	db, err := database.New(ctx, config.EnvLoad(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	err = db.WithinTx(ctx, func(ctx context.Context) error {
		roles := role.NewService(role.NewRepository(db), db)
		f := newFixture(ctx, t, db)

		seller := f.role("Seller", "products.view")
		cashier := f.role("Cashier", "users.view")
		admin := f.user(f.role("Admin"), "products.view", "locations.view", "users.view", "categories.view")
		// categories.view is a personal permission, outside the Seller package.
		employee := f.user(seller, "products.view", "categories.view")

		change := domain.RolePermissionRequest{RoleId: seller, CompanyId: f.company, GrantorId: admin}

		change.PermissionId = f.perm("locations.view")
		if err := roles.AddRolePermission(ctx, change); err != nil {
			t.Fatalf("AddRolePermission: %v", err)
		}
		f.wantCodes(employee, "categories.view", "locations.view", "products.view")

		change.PermissionId = f.perm("products.view")
		if err := roles.RemoveRolePermission(ctx, change); err != nil {
			t.Fatalf("RemoveRolePermission: %v", err)
		}
		f.wantCodes(employee, "categories.view", "locations.view")

		assign := domain.AssignRoleRequest{UserId: employee, CompanyId: f.company, RoleId: cashier, AssignerId: admin}
		if err := roles.AssignRole(ctx, assign); err != nil {
			t.Fatalf("AssignRole: %v", err)
		}
		// locations.view came with the Seller package and goes with it;
		// the personal categories.view stays.
		f.wantCodes(employee, "categories.view", "users.view")

		return errRollback
	})
	if !errors.Is(err, errRollback) {
		t.Fatalf("WithinTx: %v", err)
	}
}

type fixture struct {
	ctx     context.Context
	t       *testing.T
	db      *database.Database
	company string
}

func newFixture(ctx context.Context, t *testing.T, db *database.Database) *fixture {
	f := &fixture{ctx: ctx, t: t, db: db}
	f.scan(&f.company, `INSERT INTO companies(id, label) VALUES (gen_random_uuid(), 'integration test') RETURNING id`)
	return f
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

func (f *fixture) perm(code string) string {
	f.t.Helper()
	var id string
	f.scan(&id, `SELECT id FROM permissions WHERE code = $1 AND deleted_at = 0`, code)
	return id
}

// role creates a role whose package holds the given permission codes.
func (f *fixture) role(name string, codes ...string) string {
	f.t.Helper()
	var id string
	f.scan(&id, `INSERT INTO user_roles(company_id, name) VALUES ($1, $2) RETURNING id`, f.company, name)
	for _, code := range codes {
		f.exec(`INSERT INTO role_permissions(user_role_id, permission_id) VALUES ($1, $2)`, id, f.perm(code))
	}
	return id
}

// user creates a user of the role holding exactly the given permissions.
func (f *fixture) user(roleId string, codes ...string) string {
	f.t.Helper()
	var id string
	phone := fmt.Sprintf("+99800%07d", rand.IntN(10_000_000))
	f.scan(&id, `
		INSERT INTO users(id, company_id, user_role_id, first_name, phone, password_hash)
		VALUES (gen_random_uuid(), $1, $2, 'Test', $3, 'x') RETURNING id
	`, f.company, roleId, phone)
	for _, code := range codes {
		f.exec(`INSERT INTO user_permissions(user_id, permission_id) VALUES ($1, $2)`, id, f.perm(code))
	}
	return id
}

func (f *fixture) wantCodes(userId string, want ...string) {
	f.t.Helper()
	rows, err := f.db.Conn(f.ctx).Query(f.ctx, `
		SELECT p.code FROM user_permissions up
		JOIN permissions p ON p.id = up.permission_id
		WHERE up.user_id = $1 ORDER BY p.code
	`, userId)
	if err != nil {
		f.t.Fatalf("list permissions: %v", err)
	}
	defer rows.Close()

	var got []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			f.t.Fatalf("scan: %v", err)
		}
		got = append(got, code)
	}
	if !slices.Equal(got, want) {
		f.t.Fatalf("permissions = %v, want %v", got, want)
	}
}
