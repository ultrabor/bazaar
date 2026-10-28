package role

import (
	"bazaar/internal/platform/httpx/httpxtest"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
)

type fakeRoleService struct {
	roleService

	err      error
	created  CreateRoleRequest
	assigned AssignRoleRequest
	archived string
}

func (f *fakeRoleService) CreateRole(ctx context.Context, rq CreateRoleRequest) (*CreateRoleResponse, error) {
	f.created = rq
	if f.err != nil {
		return nil, f.err
	}
	return &CreateRoleResponse{RoleId: "role-1"}, nil
}

func (f *fakeRoleService) ArchiveRole(ctx context.Context, roleId, companyId string) error {
	f.archived = roleId
	return f.err
}

func (f *fakeRoleService) AssignRole(ctx context.Context, rq AssignRoleRequest) error {
	f.assigned = rq
	return f.err
}

func newRoleRouter(svc roleService) http.Handler {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := NewHandler(log, svc)

	r := chi.NewRouter()
	r.Post("/role", h.CreateRole)
	r.Patch("/role/{id}/archive", h.ArchiveRole)
	r.Put("/user/{id}/role", h.AssignRole)
	return r
}

func TestCreateRoleTakesCompanyFromToken(t *testing.T) {
	svc := &fakeRoleService{}

	res := httpxtest.Serve(t, newRoleRouter(svc), http.MethodPost, "/role", `{"name":"Seller","company_id":"company-from-body"}`)

	if res.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusCreated)
	}
	if svc.created.CompanyId != "company-from-token" || svc.created.Name != "Seller" {
		t.Fatalf("CreateRole(%+v), want company-from-token and Seller", svc.created)
	}
}

func TestAssignRoleUsesPathAndToken(t *testing.T) {
	svc := &fakeRoleService{}

	res := httpxtest.Serve(t, newRoleRouter(svc), http.MethodPut, "/user/user-42/role", `{"role_id":"role-1"}`)

	if res.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNoContent)
	}
	want := AssignRoleRequest{UserId: "user-42", CompanyId: "company-from-token", RoleId: "role-1", AssignerId: "admin-1"}
	if svc.assigned != want {
		t.Fatalf("AssignRole(%+v), want %+v", svc.assigned, want)
	}
}

func TestRoleErrorResponses(t *testing.T) {
	tests := []struct {
		err        error
		wantStatus int
		wantCode   string
	}{
		{ErrSystemRole, http.StatusForbidden, "system_role"},
		{ErrRoleInUse, http.StatusConflict, "role_in_use"},
		{ErrRoleNotFound, http.StatusNotFound, "role_not_found"},
		{ErrRoleExceedsPermissions, http.StatusForbidden, "role_exceeds_permissions"},
		{ErrRoleNotAssignable, http.StatusBadRequest, "invalid_role"},
		{ErrUserNotFound, http.StatusNotFound, "user_not_found"},
		{errors.New("db down"), http.StatusInternalServerError, "internal_error"},
	}

	for _, tt := range tests {
		t.Run(tt.wantCode, func(t *testing.T) {
			res := httpxtest.Serve(t, newRoleRouter(&fakeRoleService{err: tt.err}), http.MethodPatch, "/role/role-1/archive", "")

			if res.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", res.Code, tt.wantStatus)
			}
			if code := httpxtest.ErrorCode(t, res); code != tt.wantCode {
				t.Fatalf("code = %q, want %q", code, tt.wantCode)
			}
		})
	}
}
