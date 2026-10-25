package user

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
	createErr error
	created   *domain.CreateUserRequest
	createdTx int
}

func (f *fakeRepo) GetUserById(ctx context.Context, userId, companyId string) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

func (f *fakeRepo) CreateUser(ctx context.Context, req *domain.CreateUserRequest) (*domain.CreateUserResponse, error) {
	f.created = req
	f.createdTx = txCall(ctx)
	if f.createErr != nil {
		return nil, f.createErr
	}
	return &domain.CreateUserResponse{UserID: "user-1"}, nil
}

type fakeRoles struct {
	// Embedded so the fake only implements the methods the test uses.
	domain.RoleService

	checkErr error
	grantErr error
	giveErr  error

	grantor string

	gaveUserId string
	gaveRoleId string
	gaveTx     int
}

func (f *fakeRoles) GetCompanyRoles(ctx context.Context, companyId string) ([]domain.Role, error) {
	return nil, nil
}

func (f *fakeRoles) CreateOwnerRole(ctx context.Context, companyId string) (string, error) {
	return "", nil
}

func (f *fakeRoles) CheckAssignable(ctx context.Context, roleId, companyId string) error {
	return f.checkErr
}

func (f *fakeRoles) CheckGrantable(ctx context.Context, roleId, grantorId string) error {
	f.grantor = grantorId
	return f.grantErr
}

func (f *fakeRoles) GivePermissions(ctx context.Context, userId, roleId string) error {
	f.gaveUserId = userId
	f.gaveRoleId = roleId
	f.gaveTx = txCall(ctx)
	return f.giveErr
}

func validRequest() *domain.CreateUserRequest {
	return &domain.CreateUserRequest{
		FirstName:  "Ali",
		Phone:      "+998901234567",
		Password:   "Secret#123",
		CompanyId:  "company-1",
		UserRoleId: "role-1",
		CreatorId:  "admin-1",
	}
}

func TestCreateUserRejectsRoleWiderThanCreator(t *testing.T) {
	repo, tx := &fakeRepo{}, &fakeTx{}
	roles := &fakeRoles{grantErr: domain.ErrRoleExceedsPermissions}

	_, err := NewService(repo, tx, roles).CreateUser(context.Background(), validRequest())

	if !errors.Is(err, domain.ErrRoleExceedsPermissions) {
		t.Fatalf("CreateUser error = %v, want %v", err, domain.ErrRoleExceedsPermissions)
	}
	if roles.grantor != "admin-1" {
		t.Fatalf("CheckGrantable grantor = %q, want admin-1", roles.grantor)
	}
	if tx.calls != 0 || repo.created != nil {
		t.Fatalf("WithinTx calls = %d, created = %v; want no insert", tx.calls, repo.created)
	}
}

func TestCreateUserRejectsNotAssignableRole(t *testing.T) {
	repo, tx := &fakeRepo{}, &fakeTx{}
	roles := &fakeRoles{checkErr: domain.ErrRoleNotAssignable}

	_, err := NewService(repo, tx, roles).CreateUser(context.Background(), validRequest())

	if !errors.Is(err, domain.ErrInvalidRole) {
		t.Fatalf("CreateUser error = %v, want %v", err, domain.ErrInvalidRole)
	}
	if tx.calls != 0 {
		t.Fatalf("WithinTx calls = %d, want 0", tx.calls)
	}
	if repo.created != nil {
		t.Fatal("CreateUser reached the repository")
	}
}

func TestCreateUserRunsInsertAndPermissionsInOneTx(t *testing.T) {
	repo, tx, roles := &fakeRepo{}, &fakeTx{}, &fakeRoles{}

	res, err := NewService(repo, tx, roles).CreateUser(context.Background(), validRequest())

	if err != nil {
		t.Fatalf("CreateUser error = %v", err)
	}
	if res.UserID != "user-1" {
		t.Fatalf("UserID = %q, want %q", res.UserID, "user-1")
	}
	if tx.calls != 1 {
		t.Fatalf("WithinTx calls = %d, want 1", tx.calls)
	}
	if repo.createdTx != 1 || roles.gaveTx != 1 {
		t.Fatalf("tx of CreateUser = %d, GivePermissions = %d, want both 1", repo.createdTx, roles.gaveTx)
	}
	if roles.gaveUserId != "user-1" || roles.gaveRoleId != "role-1" {
		t.Fatalf("GivePermissions(%q, %q), want (user-1, role-1)", roles.gaveUserId, roles.gaveRoleId)
	}
	if repo.created.PasswordHash == "" || repo.created.PasswordHash == repo.created.Password {
		t.Fatal("password was not hashed before insert")
	}
}

func TestCreateUserFailsWhenGivePermissionsFails(t *testing.T) {
	giveErr := errors.New("give permissions failed")
	repo, tx := &fakeRepo{}, &fakeTx{}
	roles := &fakeRoles{giveErr: giveErr}

	res, err := NewService(repo, tx, roles).CreateUser(context.Background(), validRequest())

	if !errors.Is(err, giveErr) {
		t.Fatalf("CreateUser error = %v, want %v", err, giveErr)
	}
	if res != nil {
		t.Fatalf("CreateUser response = %+v, want nil", res)
	}
	// The error must come out of fn, so the real TxManager rolls back the insert.
	if !errors.Is(tx.lastErr, giveErr) {
		t.Fatalf("WithinTx fn returned %v, want %v", tx.lastErr, giveErr)
	}
}

func TestCreateUserReturnsPhoneTaken(t *testing.T) {
	repo := &fakeRepo{createErr: domain.ErrPhoneTaken}
	tx, roles := &fakeTx{}, &fakeRoles{}

	_, err := NewService(repo, tx, roles).CreateUser(context.Background(), validRequest())

	if !errors.Is(err, domain.ErrPhoneTaken) {
		t.Fatalf("CreateUser error = %v, want %v", err, domain.ErrPhoneTaken)
	}
	if roles.gaveUserId != "" {
		t.Fatal("GivePermissions called after failed insert")
	}
}
