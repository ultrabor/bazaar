package permission

import (
	"bazaar/internal/domain"
	"context"
	"errors"
	"testing"
)

const (
	viewerId  = "11111111-1111-1111-1111-111111111111"
	otherId   = "22222222-2222-2222-2222-222222222222"
	companyId = "company-1"
)

// fakeRepo keeps permissions per user.
type fakeRepo struct {
	perms   map[string][]domain.Permission
	listFor []string
	set     []string
	removed []string
}

func (f *fakeRepo) GetListByUserId(ctx context.Context, userId string) ([]domain.Permission, error) {
	f.listFor = append(f.listFor, userId)
	return f.perms[userId], nil
}

func (f *fakeRepo) GetPermissionByCode(ctx context.Context, userId, code string) (*domain.Permission, error) {
	for _, p := range f.perms[userId] {
		if p.Code == code {
			return &p, nil
		}
	}
	return nil, domain.ErrPermissionNotFound
}

func (f *fakeRepo) SetPermission(ctx context.Context, userId, permissionId string) error {
	f.set = append(f.set, userId+":"+permissionId)
	return nil
}

func (f *fakeRepo) RemovePermission(ctx context.Context, userId, permissionId string) error {
	f.removed = append(f.removed, userId+":"+permissionId)
	return nil
}

// fakeUsers knows which company each user id belongs to.
type fakeUsers struct {
	company map[string]string
}

func (f *fakeUsers) GetUserById(ctx context.Context, userId, companyId string) (*domain.User, error) {
	if f.company[userId] != companyId {
		return nil, domain.ErrUserNotFound
	}
	return &domain.User{Id: userId, CompanyId: companyId}, nil
}

func (f *fakeUsers) CreateUser(ctx context.Context, req *domain.CreateUserRequest) (*domain.CreateUserResponse, error) {
	return nil, errors.New("not used")
}

var (
	viewPerm  = domain.Permission{Id: "p-view", Code: "permissions.view"}
	salesPerm = domain.Permission{Id: "p-sales", Code: "products.view"}
)

func TestGetUserPermissions(t *testing.T) {
	tests := []struct {
		name       string
		viewer     []domain.Permission
		otherIn    string
		userId     string
		wantErr    error
		wantListed string
	}{
		{name: "own without user_id", viewer: []domain.Permission{salesPerm}, userId: "", wantListed: viewerId},
		{name: "own by id", viewer: []domain.Permission{salesPerm}, userId: viewerId, wantListed: viewerId},
		{name: "other without permissions.view", viewer: []domain.Permission{salesPerm}, otherIn: companyId, userId: otherId, wantErr: domain.ErrForbidden},
		{name: "other in same company", viewer: []domain.Permission{viewPerm}, otherIn: companyId, userId: otherId, wantListed: otherId},
		{name: "other in another company", viewer: []domain.Permission{viewPerm}, otherIn: "company-2", userId: otherId, wantErr: domain.ErrUserNotFound},
		{name: "not a uuid", viewer: []domain.Permission{viewPerm}, userId: "abc", wantErr: domain.ErrInvalidUserID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{perms: map[string][]domain.Permission{
				viewerId: tt.viewer,
				otherId:  {salesPerm},
			}}
			users := &fakeUsers{company: map[string]string{viewerId: companyId, otherId: tt.otherIn}}

			_, err := NewService(repo, users).GetUserPermissions(context.Background(), viewerId, companyId, tt.userId)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GetUserPermissions error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if len(repo.listFor) != 0 {
					t.Fatalf("listed permissions of %v after error", repo.listFor)
				}
				return
			}
			if len(repo.listFor) != 1 || repo.listFor[0] != tt.wantListed {
				t.Fatalf("listed permissions of %v, want [%s]", repo.listFor, tt.wantListed)
			}
		})
	}
}

func TestChangePermissionChecksTargetCompany(t *testing.T) {
	tests := []struct {
		name    string
		otherIn string
		applyTo string
		permId  string
		wantErr error
	}{
		{name: "same company, owned", otherIn: companyId, applyTo: otherId, permId: salesPerm.Id},
		{name: "another company", otherIn: "company-2", applyTo: otherId, permId: salesPerm.Id, wantErr: domain.ErrUserNotFound},
		{name: "not a uuid", otherIn: companyId, applyTo: "abc", permId: salesPerm.Id, wantErr: domain.ErrInvalidUserID},
		{name: "permission not owned", otherIn: companyId, applyTo: otherId, permId: "p-other", wantErr: domain.ErrPermissionNotOwned},
	}

	ops := map[string]func(s domain.PermissionService, rq domain.ChangePermissionRequest) error{
		"set": func(s domain.PermissionService, rq domain.ChangePermissionRequest) error {
			return s.SetPermission(context.Background(), rq)
		},
		"remove": func(s domain.PermissionService, rq domain.ChangePermissionRequest) error {
			return s.RemovePermission(context.Background(), rq)
		},
	}

	for opName, op := range ops {
		for _, tt := range tests {
			t.Run(opName+"/"+tt.name, func(t *testing.T) {
				repo := &fakeRepo{perms: map[string][]domain.Permission{viewerId: {salesPerm}}}
				users := &fakeUsers{company: map[string]string{otherId: tt.otherIn}}
				rq := domain.ChangePermissionRequest{
					GuarantorUserId: viewerId,
					CompanyId:       companyId,
					ApplyUserId:     tt.applyTo,
					PermissionId:    tt.permId,
				}

				err := op(NewService(repo, users), rq)

				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				changed := len(repo.set) + len(repo.removed)
				if tt.wantErr != nil && changed != 0 {
					t.Fatalf("permissions changed after error: set=%v removed=%v", repo.set, repo.removed)
				}
				if tt.wantErr == nil && changed != 1 {
					t.Fatalf("changes = %d, want 1", changed)
				}
			})
		}
	}
}
