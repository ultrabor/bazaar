package auth

import (
	"bazaar/internal/platform/support/apperror"
	"bazaar/internal/platform/support/validator"
	"context"
	"errors"
	"testing"
	"time"
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
	user          *User
	session       *Session

	companyTx  int
	userTx     int
	userRoleId string

	createdHash  string
	createdUA    string
	createdIP    string
	lookedUpHash string
	revokedId    string
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

func (f *fakeRepo) GetUserByPhone(ctx context.Context, phone string) (*User, error) {
	if f.user == nil || f.user.Phone != phone {
		return nil, apperror.ErrInvalidCred
	}
	return f.user, nil
}

func (f *fakeRepo) GetUserById(ctx context.Context, userId string) (*User, error) {
	if f.user == nil || f.user.Id != userId {
		return nil, apperror.ErrInvalidCred
	}
	return f.user, nil
}

func (f *fakeRepo) CreateSession(ctx context.Context, userId, tokenHash string, expiresAt time.Time, userAgent, ip string) (string, error) {
	f.createdHash, f.createdUA, f.createdIP = tokenHash, userAgent, ip
	return "session-1", nil
}

func (f *fakeRepo) GetActiveSessionByTokenHash(ctx context.Context, tokenHash string) (*Session, error) {
	f.lookedUpHash = tokenHash
	if f.session == nil || f.session.TokenHash != tokenHash {
		return nil, apperror.ErrInvalidCred
	}
	return f.session, nil
}

func (f *fakeRepo) RevokeSession(ctx context.Context, sessionId string) error {
	if f.session == nil || f.session.Id != sessionId {
		return apperror.ErrInvalidCred
	}
	f.revokedId = sessionId
	return nil
}

type fakeRoles struct {
	giveErr error

	ownerCompanyId string
	ownerTx        int
	gaveUserId     string
	gaveRoleId     string
	gaveTx         int
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

func validRegister() RegisterRequest {
	return RegisterRequest{
		FirstName:   "Ali",
		CompanyName: "Bazaar",
		Phone:       "+998901234567",
		Password:    "Secret#123",
	}
}

func TestRegisterOwnerRunsEverythingInOneTx(t *testing.T) {
	repo, tx, roles := &fakeRepo{}, &fakeTx{}, &fakeRoles{}

	res, err := NewService(repo, tx, roles, time.Hour).RegisterOwner(context.Background(), validRegister())

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
		{name: "phone taken", repo: &fakeRepo{createUserErr: ErrPhoneTaken}, roles: &fakeRoles{}, want: ErrPhoneTaken},
		{name: "give permissions", repo: &fakeRepo{}, roles: &fakeRoles{giveErr: giveErr}, want: giveErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &fakeTx{}

			res, err := NewService(tt.repo, tx, tt.roles, time.Hour).RegisterOwner(context.Background(), validRegister())

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

	_, err := NewService(&fakeRepo{}, tx, &fakeRoles{}, time.Hour).RegisterOwner(context.Background(), rq)

	if !errors.Is(err, ErrInvalidPhone) {
		t.Fatalf("RegisterOwner error = %v, want %v", err, ErrInvalidPhone)
	}
	if tx.calls != 0 {
		t.Fatalf("WithinTx calls = %d, want 0", tx.calls)
	}
}

func loginUser(t *testing.T) *User {
	t.Helper()
	hash, err := validator.HashPassword("Secret123!")
	if err != nil {
		t.Fatalf("HashPassword error = %v", err)
	}
	return &User{Id: "user-1", CompanyId: "company-1", Phone: "+998901234567", PasswordHash: hash}
}

func TestLoginReturnsRawTokenAndStoresOnlyItsHash(t *testing.T) {
	repo := &fakeRepo{user: loginUser(t)}
	rq := LoginRequest{Phone: "+998901234567", Password: "Secret123!", UserAgent: "curl", Ip: "10.0.0.1"}

	res, err := NewService(repo, &fakeTx{}, &fakeRoles{}, time.Hour).Login(context.Background(), rq)

	if err != nil {
		t.Fatalf("Login error = %v", err)
	}
	if res.Token == "" || res.Token == repo.createdHash {
		t.Fatalf("token = %q, stored hash = %q; want raw token, not the hash", res.Token, repo.createdHash)
	}
	if repo.createdHash != hashToken(res.Token) {
		t.Fatalf("stored hash = %q, want hashToken(token)", repo.createdHash)
	}
	if repo.createdUA != "curl" || repo.createdIP != "10.0.0.1" {
		t.Fatalf("session ua/ip = %q/%q", repo.createdUA, repo.createdIP)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	repo := &fakeRepo{user: loginUser(t)}

	_, err := NewService(repo, &fakeTx{}, &fakeRoles{}, time.Hour).Login(context.Background(), LoginRequest{Phone: "+998901234567", Password: "wrong"})

	if !errors.Is(err, apperror.ErrInvalidCred) || repo.createdHash != "" {
		t.Fatalf("Login error = %v, session created = %v; want ErrInvalidCred, none", err, repo.createdHash != "")
	}
}

func TestLoadSessionFindsUserByTokenHash(t *testing.T) {
	user := loginUser(t)
	repo := &fakeRepo{user: user, session: &Session{Id: "session-1", UserId: user.Id, TokenHash: hashToken("tok")}}
	svc := NewService(repo, &fakeTx{}, &fakeRoles{}, time.Hour)

	u, s, err := svc.LoadSession(context.Background(), "tok")

	if err != nil || u.CompanyId != "company-1" || s.Id != "session-1" {
		t.Fatalf("LoadSession = %+v, %+v, %v", u, s, err)
	}
	if repo.lookedUpHash != hashToken("tok") {
		t.Fatalf("looked up %q, want the token hash", repo.lookedUpHash)
	}

	if _, _, err := svc.LoadSession(context.Background(), "other"); !errors.Is(err, apperror.ErrInvalidCred) {
		t.Fatalf("LoadSession(unknown) error = %v, want ErrInvalidCred", err)
	}
}

func TestLogoutRevokesGivenSession(t *testing.T) {
	repo := &fakeRepo{session: &Session{Id: "session-1"}}
	svc := NewService(repo, &fakeTx{}, &fakeRoles{}, time.Hour)

	if err := svc.Logout(context.Background(), "session-1"); err != nil || repo.revokedId != "session-1" {
		t.Fatalf("Logout error = %v, revoked = %q", err, repo.revokedId)
	}
	if err := svc.Logout(context.Background(), ""); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("Logout(\"\") error = %v, want ErrUnauthorized", err)
	}
}
