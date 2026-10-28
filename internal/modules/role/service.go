package role

import (
	"bazaar/internal/platform/database"
	"bazaar/internal/platform/support/apperror"
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

/*
Маршрут	Право
GET /role, GET /role/{id}	roles.view
POST /role	roles.create
PUT /role/{id}	roles.edit
PATCH /role/{id}/archive	roles.archive
POST /role/{id}/permission	roles.edit
DELETE /role/{id}/permission/{permissionId}	roles.edit
PUT /user/{id}/role	roles.assign
*/

type Service struct {
	repo repository
	tx   database.TxManager
}

func NewService(repo repository, tx database.TxManager) *Service {
	return &Service{repo: repo, tx: tx}
}

func (s *Service) CreateOwnerRole(ctx context.Context, companyId string) (string, error) {
	var roleId string
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		roleId, err = s.repo.CreateSystemRole(ctx, companyId)
		if err != nil {
			return err
		}

		return s.repo.GiveAllPermissions(ctx, roleId)
	})
	if err != nil {
		return "", err
	}

	return roleId, nil
}

func (s *Service) GetCompanyRoles(ctx context.Context, companyId string) ([]Role, error) {
	return s.repo.GetCompanyRoles(ctx, companyId)
}

func (s *Service) CheckAssignable(ctx context.Context, roleId, companyId string) error {
	role, err := s.repo.GetById(ctx, roleId, companyId)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			return ErrRoleNotAssignable
		}
		return err
	}

	if role.IsSystem {
		return ErrRoleNotAssignable
	}

	return nil
}

func (s *Service) CheckGrantable(ctx context.Context, roleId, grantorId string) error {
	ok, err := s.repo.IsSubsetOfUser(ctx, roleId, grantorId)
	if err != nil {
		return err
	}
	if !ok {
		return ErrRoleExceedsPermissions
	}
	return nil
}

func (s *Service) GivePermissions(ctx context.Context, userId, roleId string) error {
	return s.repo.GivePermissions(ctx, userId, roleId)
}

const maxRoleNameLen = 100

func (s *Service) ListRoles(ctx context.Context, companyId string) ([]RoleSummary, error) {
	if companyId == "" {
		return nil, apperror.ErrInvalidCred
	}
	return s.repo.ListRoles(ctx, companyId)
}

func (s *Service) GetRole(ctx context.Context, roleId, companyId string) (*RoleDetails, error) {
	role, err := s.repo.GetById(ctx, roleId, companyId)
	if err != nil {
		return nil, err
	}

	permissions, err := s.repo.GetPermissions(ctx, role.Id)
	if err != nil {
		return nil, err
	}

	return &RoleDetails{Role: *role, Permissions: permissions}, nil
}

func (s *Service) CreateRole(ctx context.Context, rq CreateRoleRequest) (*CreateRoleResponse, error) {
	if rq.CompanyId == "" {
		return nil, apperror.ErrInvalidCred
	}

	name, err := roleName(rq.Name)
	if err != nil {
		return nil, err
	}

	id, err := s.repo.CreateRole(ctx, rq.CompanyId, name)
	if err != nil {
		return nil, err
	}

	return &CreateRoleResponse{RoleId: id}, nil
}

func (s *Service) RenameRole(ctx context.Context, rq UpdateRoleRequest) error {
	if _, err := s.editableRole(ctx, rq.RoleId, rq.CompanyId); err != nil {
		return err
	}

	name, err := roleName(rq.Name)
	if err != nil {
		return err
	}

	return s.repo.RenameRole(ctx, rq.RoleId, rq.CompanyId, name)
}

func (s *Service) ArchiveRole(ctx context.Context, roleId, companyId string) error {
	if _, err := s.editableRole(ctx, roleId, companyId); err != nil {
		return err
	}

	users, err := s.repo.CountActiveUsers(ctx, roleId)
	if err != nil {
		return err
	}
	if users > 0 {
		return ErrRoleInUse
	}

	return s.repo.ArchiveRole(ctx, roleId, companyId)
}

func (s *Service) AddRolePermission(ctx context.Context, rq RolePermissionRequest) error {
	if err := s.checkRolePermissionChange(ctx, rq); err != nil {
		return err
	}
	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.repo.AddPermission(ctx, rq.RoleId, rq.PermissionId); err != nil {
			return err
		}
		return s.repo.GivePermissionToRoleUsers(ctx, rq.RoleId, rq.PermissionId)
	})
}

func (s *Service) RemoveRolePermission(ctx context.Context, rq RolePermissionRequest) error {
	if err := s.checkRolePermissionChange(ctx, rq); err != nil {
		return err
	}
	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.repo.RemovePermission(ctx, rq.RoleId, rq.PermissionId); err != nil {
			return err
		}
		return s.repo.RevokePermissionFromRoleUsers(ctx, rq.RoleId, rq.PermissionId)
	})
}

func (s *Service) AssignRole(ctx context.Context, rq AssignRoleRequest) error {
	if _, err := uuid.Parse(rq.UserId); err != nil {
		return ErrInvalidUserID
	}

	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		current, err := s.repo.GetUserRole(ctx, rq.UserId, rq.CompanyId)
		if err != nil {
			return err
		}
		// The owner keeps the system role; otherwise a company could lose
		// its only user with all permissions.
		if current.IsSystem {
			return ErrSystemRole
		}

		if err := s.CheckAssignable(ctx, rq.RoleId, rq.CompanyId); err != nil {
			return err
		}
		if err := s.CheckGrantable(ctx, rq.RoleId, rq.AssignerId); err != nil {
			return err
		}

		// Revoke reads the current role, so it must run before SetUserRole.
		if err := s.repo.RevokeCurrentRolePermissions(ctx, rq.UserId); err != nil {
			return err
		}
		if err := s.repo.SetUserRole(ctx, rq.UserId, rq.RoleId); err != nil {
			return err
		}
		return s.repo.GivePermissions(ctx, rq.UserId, rq.RoleId)
	})
}

// editableRole returns an active role of the company that is not the
// system role.
func (s *Service) editableRole(ctx context.Context, roleId, companyId string) (*Role, error) {
	role, err := s.repo.GetById(ctx, roleId, companyId)
	if err != nil {
		return nil, err
	}
	if role.IsSystem {
		return nil, ErrSystemRole
	}
	return role, nil
}

// checkRolePermissionChange allows changing a role package only for
// permissions the grantor holds, so nobody can build a role wider than
// themselves.
func (s *Service) checkRolePermissionChange(ctx context.Context, rq RolePermissionRequest) error {
	if _, err := s.editableRole(ctx, rq.RoleId, rq.CompanyId); err != nil {
		return err
	}

	has, err := s.repo.UserHasPermission(ctx, rq.GrantorId, rq.PermissionId)
	if err != nil {
		return err
	}
	if !has {
		return ErrPermissionNotOwned
	}
	return nil
}

func roleName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > maxRoleNameLen {
		return "", ErrInvalidRoleName
	}
	return name, nil
}

// repository is the storage Service needs; Repository implements it.
type repository interface {
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
