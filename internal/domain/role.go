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
}

type RoleService interface {
	GetCompanyRoles(ctx context.Context, companyId string) ([]Role, error)
	// CheckAssignable returns ErrRoleNotAssignable when the role is missing,
	// belongs to another company, is archived or is a system role.
	CheckAssignable(ctx context.Context, roleId, companyId string) error
	GivePermissions(ctx context.Context, userId, roleId string) error
}
