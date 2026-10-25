package auth

import (
	"bazaar/internal/domain"
	"context"
	"errors"
	"testing"
)

// txMarker lets fakes see which WithinTx call (if any) their ctx came from.
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
	createUserErr error

	companyTx  int
	userTx     int
	userRoleId string
}

func (f *fakeRepo) CreateCompany(ctx context.Context, name string) (string, error) {
	f.companyTx = txCall(ctx)
	return "company-1", nil
}

func (f *fakeRepo) CreateUser(ctx context.Context, companyId, roleId, firstName, lastName, phone, passwordHash string) (string, error) {
	f.userTx = txCall(ctx)
	f.userRoleId = roleId
	if f.createUserErr != nil {
		return "", f.createUserErr
	}
	return "user-1", nil
}

func (f *fakeRepo) GetUserByPhone(ctx context.Context, phone string) (*domain.User, error) {
	return nil, domain.ErrInvalidCred
}

func (f *fakeRepo) GetUserById(ctx context.Context, userId string) (*domain.User, error) {
	return nil, domain.ErrInvalidCred
}

type fakeRoles struct {
	// Embedded so the fake only implements the methods the test uses.
	domain.RoleService

	giveErr error

	ownerCompanyId string
	ownerTx        int
	gaveUserId     string
	gaveRoleId     string
	gaveTx         int
}

func (f *fakeRoles) GetCompanyRoles(ctx context.Context, companyId string) ([]domain.Role, error) {
	return nil, nil
}

func (f *fakeRoles) CheckAssignable(ctx context.Context, roleId, companyId string) error {
	return nil
}

func (f *fakeRoles) CheckGrantable(ctx context.Context, roleId, grantorId string) error {
	return nil
}

func (f *fakeRoles) CreateOwnerRole(ctx context.Context, companyId string) (string, error) {
	f.ownerCompanyId = companyId
	f.ownerTx = txCall(ctx)
	return "role-owner", nil
}

func (f *fakeRoles) GivePermissions(ctx context.Context, userId, roleId string) error {
	f.gaveUserId, f.gaveRoleId = userId, roleId
	f.gaveTx = txCall(ctx)
	return f.giveErr
}

func validRegister() domain.RegisterRequest {
	return domain.RegisterRequest{
		FirstName:   "Ali",
		CompanyName: "Bazaar",
		Phone:       "+998901234567",
		Password:    "Secret#123",
	}
}

func TestRegisterOwnerRunsEverythingInOneTx(t *testing.T) {
	repo, tx, roles := &fakeRepo{}, &fakeTx{}, &fakeRoles{}

	res, err := NewService(repo, tx, roles, []byte("secret")).RegisterOwner(context.Background(), validRegister())

	if err != nil {
		t.Fatalf("RegisterOwner error = %v", err)
	}
	if res.CompanyId != "company-1" || res.UserId != "user-1" {
		t.Fatalf("response = %+v, want company-1/user-1", res)
	}
	if tx.calls != 1 {
		t.Fatalf("WithinTx calls = %d, want 1", tx.calls)
	}
	if repo.companyTx != 1 || roles.ownerTx != 1 || repo.userTx != 1 || roles.gaveTx != 1 {
		t.Fatalf("tx of company=%d ownerRole=%d user=%d permissions=%d, want all 1",
			repo.companyTx, roles.ownerTx, repo.userTx, roles.gaveTx)
	}
	if roles.ownerCompanyId != "company-1" {
		t.Fatalf("owner role created for %q, want company-1", roles.ownerCompanyId)
	}
	if repo.userRoleId != "role-owner" {
		t.Fatalf("user role = %q, want role-owner", repo.userRoleId)
	}
	if roles.gaveUserId != "user-1" || roles.gaveRoleId != "role-owner" {
		t.Fatalf("GivePermissions(%q, %q), want (user-1, role-owner)", roles.gaveUserId, roles.gaveRoleId)
	}
}

func TestRegisterOwnerFailsInsideTx(t *testing.T) {
	giveErr := errors.New("give permissions failed")
	tests := []struct {
		name  string
		repo  *fakeRepo
		roles *fakeRoles
		want  error
	}{
		{name: "phone taken", repo: &fakeRepo{createUserErr: domain.ErrPhoneTaken}, roles: &fakeRoles{}, want: domain.ErrPhoneTaken},
		{name: "give permissions", repo: &fakeRepo{}, roles: &fakeRoles{giveErr: giveErr}, want: giveErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &fakeTx{}

			res, err := NewService(tt.repo, tx, tt.roles, []byte("secret")).RegisterOwner(context.Background(), validRegister())

			if !errors.Is(err, tt.want) || res != nil {
				t.Fatalf("RegisterOwner = %+v, %v; want nil, %v", res, err, tt.want)
			}
			// The error must come out of fn, so the real TxManager rolls back
			// the company and the owner role too.
			if !errors.Is(tx.lastErr, tt.want) {
				t.Fatalf("WithinTx fn returned %v, want %v", tx.lastErr, tt.want)
			}
		})
	}
}

func TestRegisterOwnerRejectsInvalidPhone(t *testing.T) {
	rq := validRegister()
	rq.Phone = "123"
	tx := &fakeTx{}

	_, err := NewService(&fakeRepo{}, tx, &fakeRoles{}, []byte("secret")).RegisterOwner(context.Background(), rq)

	if !errors.Is(err, domain.ErrInvalidPhone) {
		t.Fatalf("RegisterOwner error = %v, want %v", err, domain.ErrInvalidPhone)
	}
	if tx.calls != 0 {
		t.Fatalf("WithinTx calls = %d, want 0", tx.calls)
	}
}
