package user

import (
	"bazaar/internal/platform/httpx/httpxtest"
	"bazaar/internal/role"
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
)

type fakeUserService struct {
	createErr error
	getErr    error

	created      *CreateUserRequest
	getUserId    string
	getCompanyId string
}

func (f *fakeUserService) GetUserById(ctx context.Context, userId, companyId string) (*User, error) {
	f.getUserId, f.getCompanyId = userId, companyId
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &User{Id: userId, CompanyId: companyId}, nil
}

func (f *fakeUserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	f.created = req
	if f.createErr != nil {
		return nil, f.createErr
	}
	return &CreateUserResponse{UserID: "user-1"}, nil
}

// newUserRouter wires user handlers without Auth/Permission middleware;
// the current user is put into ctx directly.
func newUserRouter(svc userService) http.Handler {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := NewHandler(log, svc)

	r := chi.NewRouter()
	r.Post("/user", h.CreateUser)
	r.Get("/user/{id}", h.GetUserById)
	return r
}

func TestCreateUserTakesCompanyFromToken(t *testing.T) {
	svc := &fakeUserService{}
	body := `{"first_name":"Ali","phone":"+998901234567","password":"Secret#123",
		"user_role_id":"role-1","company_id":"company-from-body"}`

	res := httpxtest.Serve(t, newUserRouter(svc), http.MethodPost, "/user", body)

	if res.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusCreated)
	}
	if svc.created.CompanyId != "company-from-token" {
		t.Fatalf("CompanyId = %q, want company-from-token", svc.created.CompanyId)
	}
	if svc.created.UserRoleId != "role-1" {
		t.Fatalf("UserRoleId = %q, want role-1", svc.created.UserRoleId)
	}
	if svc.created.CreatorId != "admin-1" {
		t.Fatalf("CreatorId = %q, want admin-1 from token", svc.created.CreatorId)
	}
}

func TestCreateUserRoleExceedsPermissions(t *testing.T) {
	svc := &fakeUserService{createErr: role.ErrRoleExceedsPermissions}

	res := httpxtest.Serve(t, newUserRouter(svc), http.MethodPost, "/user", `{}`)

	if res.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusForbidden)
	}
	if code := httpxtest.ErrorCode(t, res); code != "role_exceeds_permissions" {
		t.Fatalf("code = %q, want role_exceeds_permissions", code)
	}
}

func TestCreateUserBadRequest(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		err      error
		wantCode string
	}{
		{name: "invalid json", body: `{`, wantCode: "invalid_json"},
		{name: "invalid role", body: `{}`, err: ErrInvalidRole, wantCode: "invalid_role"},
		{name: "invalid phone", body: `{}`, err: ErrInvalidPhone, wantCode: "invalid_phone"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := httpxtest.Serve(t, newUserRouter(&fakeUserService{createErr: tt.err}), http.MethodPost, "/user", tt.body)

			if res.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
			}
			if code := httpxtest.ErrorCode(t, res); code != tt.wantCode {
				t.Fatalf("code = %q, want %q", code, tt.wantCode)
			}
		})
	}
}

func TestGetUserByIdNotFound(t *testing.T) {
	svc := &fakeUserService{getErr: ErrUserNotFound}

	res := httpxtest.Serve(t, newUserRouter(svc), http.MethodGet, "/user/user-42", "")

	if res.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNotFound)
	}
	if code := httpxtest.ErrorCode(t, res); code != "user_not_found" {
		t.Fatalf("code = %q, want user_not_found", code)
	}
	if svc.getUserId != "user-42" || svc.getCompanyId != "company-from-token" {
		t.Fatalf("GetUserById(%q, %q), want (user-42, company-from-token)", svc.getUserId, svc.getCompanyId)
	}
}
