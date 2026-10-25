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

var (
	ErrRoleNotFound      = errors.New("role not found")
	ErrRoleNotAssignable = errors.New("role cannot be assigned")
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
}

type RoleService interface {
	// CreateOwnerRole creates the system owner role holding all permissions.
	CreateOwnerRole(ctx context.Context, companyId string) (string, error)
	GetCompanyRoles(ctx context.Context, companyId string) ([]Role, error)
	// CheckAssignable returns ErrRoleNotAssignable when the role is missing,
	// belongs to another company, is archived or is a system role.
	CheckAssignable(ctx context.Context, roleId, companyId string) error
	GivePermissions(ctx context.Context, userId, roleId string) error
}
