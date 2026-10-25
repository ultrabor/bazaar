package role

import (
	"bazaar/internal/domain"
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
)

const (
	companyId = "company-1"
	adminId   = "admin-1"
	userId    = "33333333-3333-3333-3333-333333333333"
)

// crudRepo is a role repository fake with one role and a target user.
type crudRepo struct {
	domain.RoleRepository

	role      *domain.Role // what GetById returns; nil means not found
	users     int          // active users of the role
	held      bool         // whether the grantor holds the permission
	subset    bool         // whether the role is within the grantor's permissions
	userRole  *domain.Role // current role of the target user; nil means not found
	createErr error

	writes []string
}

func (f *crudRepo) GetById(ctx context.Context, roleId, companyId string) (*domain.Role, error) {
	if f.role == nil {
		return nil, domain.ErrRoleNotFound
	}
	return f.role, nil
}

func (f *crudRepo) CreateRole(ctx context.Context, companyId, name string) (string, error) {
	f.writes = append(f.writes, "create:"+name)
	return "role-new", f.createErr
}

func (f *crudRepo) RenameRole(ctx context.Context, roleId, companyId, name string) error {
	f.writes = append(f.writes, "rename:"+name)
	return nil
}

func (f *crudRepo) ArchiveRole(ctx context.Context, roleId, companyId string) error {
	f.writes = append(f.writes, "archive")
	return nil
}

func (f *crudRepo) CountActiveUsers(ctx context.Context, roleId string) (int, error) {
	return f.users, nil
}

func (f *crudRepo) UserHasPermission(ctx context.Context, userId, permissionId string) (bool, error) {
	return f.held, nil
}

// write records a repository write with the WithinTx call it ran in.
func (f *crudRepo) write(ctx context.Context, w string) {
	f.writes = append(f.writes, fmt.Sprintf("%s@%d", w, txCall(ctx)))
}

func (f *crudRepo) AddPermission(ctx context.Context, roleId, permissionId string) error {
	f.write(ctx, "add:"+permissionId)
	return nil
}

func (f *crudRepo) RemovePermission(ctx context.Context, roleId, permissionId string) error {
	f.write(ctx, "remove:"+permissionId)
	return nil
}

func (f *crudRepo) GivePermissionToRoleUsers(ctx context.Context, roleId, permissionId string) error {
	f.write(ctx, "give-role-users:"+permissionId)
	return nil
}

func (f *crudRepo) RevokePermissionFromRoleUsers(ctx context.Context, roleId, permissionId string) error {
	f.write(ctx, "revoke-role-users:"+permissionId)
	return nil
}

func (f *crudRepo) RevokeCurrentRolePermissions(ctx context.Context, userId string) error {
	f.write(ctx, "revoke-current-package")
	return nil
}

func (f *crudRepo) GivePermissions(ctx context.Context, userId, roleId string) error {
	f.write(ctx, "give-package:"+roleId)
	return nil
}

func (f *crudRepo) IsSubsetOfUser(ctx context.Context, roleId, userId string) (bool, error) {
	return f.subset, nil
}

func (f *crudRepo) GetUserRole(ctx context.Context, userId, companyId string) (*domain.Role, error) {
	if f.userRole == nil {
		return nil, domain.ErrUserNotFound
	}
	return f.userRole, nil
}

func (f *crudRepo) SetUserRole(ctx context.Context, userId, roleId string) error {
	f.write(ctx, "set-role:"+roleId)
	return nil
}

var (
	plainRole  = &domain.Role{Id: "role-1", CompanyId: companyId, Name: "Seller"}
	systemRole = &domain.Role{Id: "role-owner", CompanyId: companyId, Name: "owner", IsSystem: true}
)

func newCrud(repo *crudRepo) domain.RoleService {
	return NewService(repo, &fakeTx{})
}

func TestCreateRole(t *testing.T) {
	tests := []struct {
		name      string
		roleName  string
		createErr error
		want      error
		wantWrite string
	}{
		{name: "trims name", roleName: "  Seller  ", wantWrite: "create:Seller"},
		{name: "empty name", roleName: "   ", want: domain.ErrInvalidRoleName},
		{name: "duplicate", roleName: "Seller", createErr: domain.ErrRoleAlreadyExists, want: domain.ErrRoleAlreadyExists, wantWrite: "create:Seller"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &crudRepo{createErr: tt.createErr}

			_, err := newCrud(repo).CreateRole(context.Background(), domain.CreateRoleRequest{CompanyId: companyId, Name: tt.roleName})

			if !errors.Is(err, tt.want) {
				t.Fatalf("CreateRole error = %v, want %v", err, tt.want)
			}
			if tt.wantWrite == "" && len(repo.writes) != 0 {
				t.Fatalf("writes = %v, want none", repo.writes)
			}
			if tt.wantWrite != "" && (len(repo.writes) != 1 || repo.writes[0] != tt.wantWrite) {
				t.Fatalf("writes = %v, want [%s]", repo.writes, tt.wantWrite)
			}
		})
	}
}

func TestSystemRoleCannotBeChanged(t *testing.T) {
	ctx := context.Background()
	perm := domain.RolePermissionRequest{RoleId: systemRole.Id, CompanyId: companyId, PermissionId: "p-1", GrantorId: adminId}
	ops := map[string]func(s domain.RoleService) error{
		"rename": func(s domain.RoleService) error {
			return s.RenameRole(ctx, domain.UpdateRoleRequest{RoleId: systemRole.Id, CompanyId: companyId, Name: "Boss"})
		},
		"archive":           func(s domain.RoleService) error { return s.ArchiveRole(ctx, systemRole.Id, companyId) },
		"add permission":    func(s domain.RoleService) error { return s.AddRolePermission(ctx, perm) },
		"remove permission": func(s domain.RoleService) error { return s.RemoveRolePermission(ctx, perm) },
	}

	for name, op := range ops {
		t.Run(name, func(t *testing.T) {
			repo := &crudRepo{role: systemRole, held: true}

			if err := op(newCrud(repo)); !errors.Is(err, domain.ErrSystemRole) {
				t.Fatalf("error = %v, want %v", err, domain.ErrSystemRole)
			}
			if len(repo.writes) != 0 {
				t.Fatalf("writes = %v, want none", repo.writes)
			}
		})
	}
}

func TestArchiveRole(t *testing.T) {
	tests := []struct {
		name  string
		repo  *crudRepo
		want  error
		wrote bool
	}{
		{name: "no users", repo: &crudRepo{role: plainRole}, wrote: true},
		{name: "in use", repo: &crudRepo{role: plainRole, users: 2}, want: domain.ErrRoleInUse},
		{name: "not found", repo: &crudRepo{}, want: domain.ErrRoleNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := newCrud(tt.repo).ArchiveRole(context.Background(), "role-1", companyId)

			if !errors.Is(err, tt.want) {
				t.Fatalf("ArchiveRole error = %v, want %v", err, tt.want)
			}
			if wrote := len(tt.repo.writes) > 0; wrote != tt.wrote {
				t.Fatalf("archived = %v, want %v", wrote, tt.wrote)
			}
		})
	}
}

func TestRolePermissionNeedsGrantorPermission(t *testing.T) {
	rq := domain.RolePermissionRequest{RoleId: plainRole.Id, CompanyId: companyId, PermissionId: "p-1", GrantorId: adminId}

	for _, held := range []bool{true, false} {
		repo := &crudRepo{role: plainRole, held: held}

		err := newCrud(repo).AddRolePermission(context.Background(), rq)

		if !held && (!errors.Is(err, domain.ErrPermissionNotOwned) || len(repo.writes) != 0) {
			t.Fatalf("not held: error = %v, writes = %v; want %v, none", err, repo.writes, domain.ErrPermissionNotOwned)
		}
		if held && err != nil {
			t.Fatalf("held: error = %v", err)
		}
	}
}

// A permission added to the package reaches the role users, and a removed
// one is taken from them, in the same transaction as the package change.
func TestRolePackageChangeReachesRoleUsers(t *testing.T) {
	rq := domain.RolePermissionRequest{RoleId: plainRole.Id, CompanyId: companyId, PermissionId: "p-1", GrantorId: adminId}
	tests := []struct {
		name string
		op   func(s domain.RoleService) error
		want []string
	}{
		{
			name: "add",
			op:   func(s domain.RoleService) error { return s.AddRolePermission(context.Background(), rq) },
			want: []string{"add:p-1@1", "give-role-users:p-1@1"},
		},
		{
			name: "remove",
			op:   func(s domain.RoleService) error { return s.RemoveRolePermission(context.Background(), rq) },
			want: []string{"remove:p-1@1", "revoke-role-users:p-1@1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &crudRepo{role: plainRole, held: true}

			if err := tt.op(newCrud(repo)); err != nil {
				t.Fatalf("error = %v", err)
			}
			if !slices.Equal(repo.writes, tt.want) {
				t.Fatalf("writes = %v, want %v", repo.writes, tt.want)
			}
		})
	}
}

func TestAssignRole(t *testing.T) {
	tests := []struct {
		name   string
		userId string
		repo   *crudRepo
		want   error
	}{
		{name: "assigns", userId: userId, repo: &crudRepo{role: plainRole, userRole: plainRole, subset: true}},
		{name: "not a uuid", userId: "abc", repo: &crudRepo{role: plainRole, userRole: plainRole, subset: true}, want: domain.ErrInvalidUserID},
		{name: "user of another company", userId: userId, repo: &crudRepo{role: plainRole, subset: true}, want: domain.ErrUserNotFound},
		{name: "target is the owner", userId: userId, repo: &crudRepo{role: plainRole, userRole: systemRole, subset: true}, want: domain.ErrSystemRole},
		{name: "role not assignable", userId: userId, repo: &crudRepo{role: systemRole, userRole: plainRole, subset: true}, want: domain.ErrRoleNotAssignable},
		{name: "role wider than assigner", userId: userId, repo: &crudRepo{role: plainRole, userRole: plainRole}, want: domain.ErrRoleExceedsPermissions},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rq := domain.AssignRoleRequest{UserId: tt.userId, CompanyId: companyId, RoleId: plainRole.Id, AssignerId: adminId}

			err := newCrud(tt.repo).AssignRole(context.Background(), rq)

			if !errors.Is(err, tt.want) {
				t.Fatalf("AssignRole error = %v, want %v", err, tt.want)
			}
			// The old package goes before the role changes, because the
			// revoke reads the current role; all in one transaction.
			want := []string{"revoke-current-package@1", "set-role:" + plainRole.Id + "@1", "give-package:" + plainRole.Id + "@1"}
			if tt.want != nil {
				want = nil
			}
			if !slices.Equal(tt.repo.writes, want) {
				t.Fatalf("writes = %v, want %v", tt.repo.writes, want)
			}
		})
	}
}
