package domain

import (
	"context"
	"errors"
)

type Role struct {
	Id        string `json:"id"`
	CompanyId string `json:"company_id"`
	Name      string `json:"name"`
	IsSystem  bool   `json:"is_system"`
}

// RoleSummary is a row of the company role list.
type RoleSummary struct {
	Id               string `json:"id"`
	Name             string `json:"name"`
	IsSystem         bool   `json:"is_system"`
	PermissionsCount int    `json:"permissions_count"`
	UserCount        int    `json:"user_count"`
}

// RoleDetails is a role with its permission package.
type RoleDetails struct {
	Role
	Permissions []Permission `json:"permissions"`
}

type CreateRoleRequest struct {
	CompanyId string `json:"company_id"`
	Name      string `json:"name"`
}

type CreateRoleResponse struct {
	RoleId string `json:"role_id"`
}

type UpdateRoleRequest struct {
	RoleId    string `json:"role_id"`
	CompanyId string `json:"company_id"`
	Name      string `json:"name"`
}

// RolePermissionRequest adds a permission to a role package or removes it.
// The grantor must hold the permission.
type RolePermissionRequest struct {
	RoleId       string `json:"role_id"`
	CompanyId    string `json:"company_id"`
	PermissionId string `json:"permission_id"`
	GrantorId    string `json:"-"`
}

// AssignRoleRequest gives a user a new role. The old package is taken
// from the user and the new one is given.
type AssignRoleRequest struct {
	UserId     string `json:"user_id"`
	CompanyId  string `json:"company_id"`
	RoleId     string `json:"role_id"`
	AssignerId string `json:"-"`
}

var (
	ErrRoleNotFound      = errors.New("role not found")
	ErrRoleNotAssignable = errors.New("role cannot be assigned")
	// ErrRoleExceedsPermissions means the role holds a permission the
	// grantor does not have.
	ErrRoleExceedsPermissions = errors.New("role exceeds grantor permissions")
	ErrRoleAlreadyExists      = errors.New("role already exists")
	ErrInvalidRoleName        = errors.New("invalid role name")
	// ErrSystemRole means the action is not allowed on the owner role or on
	// a user holding it.
	ErrSystemRole = errors.New("system role cannot be changed")
	ErrRoleInUse  = errors.New("role is assigned to active users")
)

type RoleRepository interface {
	// GetById returns an active role of the company or ErrRoleNotFound.
	GetById(ctx context.Context, roleId, companyId string) (*Role, error)
	// GetCompanyRoles returns active roles of the company, except system ones.
	GetCompanyRoles(ctx context.Context, companyId string) ([]Role, error)
	// GivePermissions copies the role's permissions to the user.
	GivePermissions(ctx context.Context, userId, roleId string) error
	// CreateSystemRole creates the company's system role named owner.
	CreateSystemRole(ctx context.Context, companyId string) (string, error)
	// GiveAllPermissions puts every active permission into the role.
	GiveAllPermissions(ctx context.Context, roleId string) error
	// IsSubsetOfUser reports whether every active permission of the role
	// is also held by the user.
	IsSubsetOfUser(ctx context.Context, roleId, userId string) (bool, error)

	// ListRoles returns all active roles of the company, system ones too.
	ListRoles(ctx context.Context, companyId string) ([]RoleSummary, error)
	// GetPermissions returns the active permissions of the role package.
	GetPermissions(ctx context.Context, roleId string) ([]Permission, error)
	// CreateRole returns ErrRoleAlreadyExists for a duplicate active name.
	CreateRole(ctx context.Context, companyId, name string) (string, error)
	// RenameRole returns ErrRoleNotFound or ErrRoleAlreadyExists.
	RenameRole(ctx context.Context, roleId, companyId, name string) error
	// ArchiveRole archives a non-system role without active users and
	// returns ErrRoleNotFound when nothing was archived.
	ArchiveRole(ctx context.Context, roleId, companyId string) error
	CountActiveUsers(ctx context.Context, roleId string) (int, error)
	// UserHasPermission reports whether the user holds the active permission.
	UserHasPermission(ctx context.Context, userId, permissionId string) (bool, error)
	AddPermission(ctx context.Context, roleId, permissionId string) error
	// RemovePermission returns ErrPermissionNotFound when the role does not
	// have the permission.
	RemovePermission(ctx context.Context, roleId, permissionId string) error
	// GivePermissionToRoleUsers gives the permission to every active user
	// holding the role.
	GivePermissionToRoleUsers(ctx context.Context, roleId, permissionId string) error
	// RevokePermissionFromRoleUsers takes the permission from every user
	// holding the role.
	RevokePermissionFromRoleUsers(ctx context.Context, roleId, permissionId string) error
	// RevokeCurrentRolePermissions takes from the user the permissions of
	// the user's current role package; personal permissions outside it stay.
	RevokeCurrentRolePermissions(ctx context.Context, userId string) error
	// GetUserRole returns the role of an active user of the company or
	// ErrUserNotFound.
	GetUserRole(ctx context.Context, userId, companyId string) (*Role, error)
	// SetUserRole points the user at the role; permissions are not touched.
	SetUserRole(ctx context.Context, userId, roleId string) error
}

type RoleService interface {
	// CreateOwnerRole creates the system owner role holding all permissions.
	CreateOwnerRole(ctx context.Context, companyId string) (string, error)
	GetCompanyRoles(ctx context.Context, companyId string) ([]Role, error)
	// CheckAssignable returns ErrRoleNotAssignable when the role is missing,
	// belongs to another company, is archived or is a system role.
	CheckAssignable(ctx context.Context, roleId, companyId string) error
	// CheckGrantable returns ErrRoleExceedsPermissions when the role holds a
	// permission the grantor does not have.
	CheckGrantable(ctx context.Context, roleId, grantorId string) error
	GivePermissions(ctx context.Context, userId, roleId string) error

	ListRoles(ctx context.Context, companyId string) ([]RoleSummary, error)
	GetRole(ctx context.Context, roleId, companyId string) (*RoleDetails, error)
	CreateRole(ctx context.Context, rq CreateRoleRequest) (*CreateRoleResponse, error)
	// RenameRole returns ErrSystemRole for the owner role.
	RenameRole(ctx context.Context, rq UpdateRoleRequest) error
	// ArchiveRole returns ErrSystemRole for the owner role and ErrRoleInUse
	// while active users hold the role.
	ArchiveRole(ctx context.Context, roleId, companyId string) error
	// AddRolePermission and RemoveRolePermission change the role package and,
	// in the same transaction, the permissions of every user of the role.
	AddRolePermission(ctx context.Context, rq RolePermissionRequest) error
	RemoveRolePermission(ctx context.Context, rq RolePermissionRequest) error
	// AssignRole checks that the user is in the company and not the owner,
	// that the role is assignable and not wider than the assigner. Then, in
	// one transaction, it takes the old package from the user, sets the role
	// and gives the new package; personal permissions outside the old
	// package stay.
	AssignRole(ctx context.Context, rq AssignRoleRequest) error
}
