package role

import (
	"bazaar/internal/domain"
	"context"
	"errors"
	"testing"
)

type txMarker struct{}

func txCall(ctx context.Context) int {
	n, _ := ctx.Value(txMarker{}).(int)
	return n
}

type fakeTx struct {
	calls   int
	lastErr error
}

func (f *fakeTx) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	f.calls++
	f.lastErr = fn(context.WithValue(ctx, txMarker{}, f.calls))
	return f.lastErr
}

type fakeRepo struct {
	role      *domain.Role
	getErr    error
	giveAllEr error
	subset    bool
	subsetErr error

	createdTx   int
	giveAllTx   int
	giveAllRole string
}

func (f *fakeRepo) GetById(ctx context.Context, roleId, companyId string) (*domain.Role, error) {
	return f.role, f.getErr
}

func (f *fakeRepo) GetCompanyRoles(ctx context.Context, companyId string) ([]domain.Role, error) {
	return nil, nil
}

func (f *fakeRepo) GivePermissions(ctx context.Context, userId, roleId string) error {
	return nil
}

func (f *fakeRepo) CreateSystemRole(ctx context.Context, companyId string) (string, error) {
	f.createdTx = txCall(ctx)
	return "role-owner", nil
}

func (f *fakeRepo) IsSubsetOfUser(ctx context.Context, roleId, userId string) (bool, error) {
	return f.subset, f.subsetErr
}

func (f *fakeRepo) GiveAllPermissions(ctx context.Context, roleId string) error {
	f.giveAllTx = txCall(ctx)
	f.giveAllRole = roleId
	return f.giveAllEr
}

func TestCreateOwnerRoleGivesAllPermissionsInOneTx(t *testing.T) {
	repo, tx := &fakeRepo{}, &fakeTx{}

	roleId, err := NewService(repo, tx).CreateOwnerRole(context.Background(), "company-1")

	if err != nil || roleId != "role-owner" {
		t.Fatalf("CreateOwnerRole = %q, %v; want role-owner, nil", roleId, err)
	}
	if tx.calls != 1 || repo.createdTx != 1 || repo.giveAllTx != 1 {
		t.Fatalf("WithinTx calls = %d, create tx = %d, give all tx = %d; want 1, 1, 1",
			tx.calls, repo.createdTx, repo.giveAllTx)
	}
	if repo.giveAllRole != "role-owner" {
		t.Fatalf("GiveAllPermissions(%q), want role-owner", repo.giveAllRole)
	}
}

func TestCreateOwnerRoleFailsWhenPermissionsFail(t *testing.T) {
	giveErr := errors.New("give all failed")
	tx := &fakeTx{}

	_, err := NewService(&fakeRepo{giveAllEr: giveErr}, tx).CreateOwnerRole(context.Background(), "company-1")

	if !errors.Is(err, giveErr) || !errors.Is(tx.lastErr, giveErr) {
		t.Fatalf("CreateOwnerRole error = %v, fn error = %v; want %v", err, tx.lastErr, giveErr)
	}
}

func TestCheckAssignable(t *testing.T) {
	dbErr := errors.New("db down")
	tests := []struct {
		name string
		repo *fakeRepo
		want error
	}{
		{name: "regular role", repo: &fakeRepo{role: &domain.Role{Id: "r"}}, want: nil},
		{name: "system role", repo: &fakeRepo{role: &domain.Role{Id: "r", IsSystem: true}}, want: domain.ErrRoleNotAssignable},
		{name: "missing, foreign or archived", repo: &fakeRepo{getErr: domain.ErrRoleNotFound}, want: domain.ErrRoleNotAssignable},
		{name: "database error", repo: &fakeRepo{getErr: dbErr}, want: dbErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewService(tt.repo, &fakeTx{}).CheckAssignable(context.Background(), "r", "company-1")
			if !errors.Is(err, tt.want) {
				t.Fatalf("CheckAssignable error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestCheckGrantable(t *testing.T) {
	dbErr := errors.New("db down")
	tests := []struct {
		name string
		repo *fakeRepo
		want error
	}{
		{name: "subset", repo: &fakeRepo{subset: true}, want: nil},
		{name: "role has more", repo: &fakeRepo{subset: false}, want: domain.ErrRoleExceedsPermissions},
		{name: "database error", repo: &fakeRepo{subsetErr: dbErr}, want: dbErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewService(tt.repo, &fakeTx{}).CheckGrantable(context.Background(), "role-1", "admin-1")
			if !errors.Is(err, tt.want) {
				t.Fatalf("CheckGrantable error = %v, want %v", err, tt.want)
			}
		})
	}
}
