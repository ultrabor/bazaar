package role

import (
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

// Permission is an entry of a role package.
type Permission struct {
	Id    string `json:"id"`
	Label string `json:"label"`
	Code  string `json:"code"`
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

	// The role module changes role packages and user roles itself, so it
	// keeps its own errors for permissions and users instead of importing
	// the permission and user modules.
	ErrPermissionNotFound = errors.New("permission not found")
	ErrPermissionNotOwned = errors.New("dont have permission to set")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidUserID      = errors.New("invalid user id")
)
