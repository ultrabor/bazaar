package httpx

import (
	"bazaar/internal/domain"
	"bazaar/internal/platform/httpx/dto"
	"bazaar/internal/platform/httpx/middleware"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

type fakeUserService struct {
	createErr error
	getErr    error

	created      *domain.CreateUserRequest
	getUserId    string
	getCompanyId string
}

func (f *fakeUserService) GetUserById(ctx context.Context, userId, companyId string) (*domain.User, error) {
	f.getUserId, f.getCompanyId = userId, companyId
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &domain.User{Id: userId, CompanyId: companyId}, nil
}

func (f *fakeUserService) CreateUser(ctx context.Context, req *domain.CreateUserRequest) (*domain.CreateUserResponse, error) {
	f.created = req
	if f.createErr != nil {
		return nil, f.createErr
	}
	return &domain.CreateUserResponse{UserID: "user-1"}, nil
}

// newUserRouter wires user handlers without Auth/Permission middleware;
// the current user is put into ctx directly.
func newUserRouter(svc domain.UserService) http.Handler {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := New(log, chi.NewRouter(), &ServiceManager{userService: svc})

	r := chi.NewRouter()
	r.Post("/user", h.CreateUser)
	r.Get("/user/{id}", h.GetUserById)
	return r
}

func serve(t *testing.T, router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	ctx := middleware.WithCurrentUser(req.Context(), &domain.User{Id: "admin-1", CompanyId: "company-from-token"})
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req.WithContext(ctx))
	return res
}

func errorCode(t *testing.T, res *httptest.ResponseRecorder) string {
	t.Helper()

	var body dto.Error
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	return body.Error.Code
}

func TestCreateUserTakesCompanyFromToken(t *testing.T) {
	svc := &fakeUserService{}
	body := `{"first_name":"Ali","phone":"+998901234567","password":"Secret#123",
		"user_role_id":"role-1","company_id":"company-from-body"}`

	res := serve(t, newUserRouter(svc), http.MethodPost, "/user", body)

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
	svc := &fakeUserService{createErr: domain.ErrRoleExceedsPermissions}

	res := serve(t, newUserRouter(svc), http.MethodPost, "/user", `{}`)

	if res.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusForbidden)
	}
	if code := errorCode(t, res); code != "role_exceeds_permissions" {
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
		{name: "invalid role", body: `{}`, err: domain.ErrInvalidRole, wantCode: "invalid_role"},
		{name: "invalid phone", body: `{}`, err: domain.ErrInvalidPhone, wantCode: "invalid_phone"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := serve(t, newUserRouter(&fakeUserService{createErr: tt.err}), http.MethodPost, "/user", tt.body)

			if res.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
			}
			if code := errorCode(t, res); code != tt.wantCode {
				t.Fatalf("code = %q, want %q", code, tt.wantCode)
			}
		})
	}
}

func TestGetUserByIdNotFound(t *testing.T) {
	svc := &fakeUserService{getErr: domain.ErrUserNotFound}

	res := serve(t, newUserRouter(svc), http.MethodGet, "/user/user-42", "")

	if res.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNotFound)
	}
	if code := errorCode(t, res); code != "user_not_found" {
		t.Fatalf("code = %q, want user_not_found", code)
	}
	if svc.getUserId != "user-42" || svc.getCompanyId != "company-from-token" {
		t.Fatalf("GetUserById(%q, %q), want (user-42, company-from-token)", svc.getUserId, svc.getCompanyId)
	}
}
