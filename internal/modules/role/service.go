package role

import (
	"bazaar/internal/domain"
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

type service struct {
	repo domain.RoleRepository
	tx   domain.TxManager
}

var _ domain.RoleService = (*service)(nil)

func NewService(repo domain.RoleRepository, tx domain.TxManager) domain.RoleService {
	return &service{repo: repo, tx: tx}
}

func (s *service) CreateOwnerRole(ctx context.Context, companyId string) (string, error) {
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

func (s *service) GetCompanyRoles(ctx context.Context, companyId string) ([]domain.Role, error) {
	return s.repo.GetCompanyRoles(ctx, companyId)
}

func (s *service) CheckAssignable(ctx context.Context, roleId, companyId string) error {
	role, err := s.repo.GetById(ctx, roleId, companyId)
	if err != nil {
		if errors.Is(err, domain.ErrRoleNotFound) {
			return domain.ErrRoleNotAssignable
		}
		return err
	}

	if role.IsSystem {
		return domain.ErrRoleNotAssignable
	}

	return nil
}

func (s *service) CheckGrantable(ctx context.Context, roleId, grantorId string) error {
	ok, err := s.repo.IsSubsetOfUser(ctx, roleId, grantorId)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrRoleExceedsPermissions
	}
	return nil
}

func (s *service) GivePermissions(ctx context.Context, userId, roleId string) error {
	return s.repo.GivePermissions(ctx, userId, roleId)
}

const maxRoleNameLen = 100

func (s *service) ListRoles(ctx context.Context, companyId string) ([]domain.RoleSummary, error) {
	if companyId == "" {
		return nil, domain.ErrInvalidCred
	}
	return s.repo.ListRoles(ctx, companyId)
}

func (s *service) GetRole(ctx context.Context, roleId, companyId string) (*domain.RoleDetails, error) {
	role, err := s.repo.GetById(ctx, roleId, companyId)
	if err != nil {
		return nil, err
	}

	permissions, err := s.repo.GetPermissions(ctx, role.Id)
	if err != nil {
		return nil, err
	}

	return &domain.RoleDetails{Role: *role, Permissions: permissions}, nil
}

func (s *service) CreateRole(ctx context.Context, rq domain.CreateRoleRequest) (*domain.CreateRoleResponse, error) {
	if rq.CompanyId == "" {
		return nil, domain.ErrInvalidCred
	}

	name, err := roleName(rq.Name)
	if err != nil {
		return nil, err
	}

	id, err := s.repo.CreateRole(ctx, rq.CompanyId, name)
	if err != nil {
		return nil, err
	}

	return &domain.CreateRoleResponse{RoleId: id}, nil
}

func (s *service) RenameRole(ctx context.Context, rq domain.UpdateRoleRequest) error {
	if _, err := s.editableRole(ctx, rq.RoleId, rq.CompanyId); err != nil {
		return err
	}

	name, err := roleName(rq.Name)
	if err != nil {
		return err
	}

	return s.repo.RenameRole(ctx, rq.RoleId, rq.CompanyId, name)
}

func (s *service) ArchiveRole(ctx context.Context, roleId, companyId string) error {
	if _, err := s.editableRole(ctx, roleId, companyId); err != nil {
		return err
	}

	users, err := s.repo.CountActiveUsers(ctx, roleId)
	if err != nil {
		return err
	}
	if users > 0 {
		return domain.ErrRoleInUse
	}

	return s.repo.ArchiveRole(ctx, roleId, companyId)
}

func (s *service) AddRolePermission(ctx context.Context, rq domain.RolePermissionRequest) error {
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

func (s *service) RemoveRolePermission(ctx context.Context, rq domain.RolePermissionRequest) error {
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

func (s *service) AssignRole(ctx context.Context, rq domain.AssignRoleRequest) error {
	if _, err := uuid.Parse(rq.UserId); err != nil {
		return domain.ErrInvalidUserID
	}

	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		current, err := s.repo.GetUserRole(ctx, rq.UserId, rq.CompanyId)
		if err != nil {
			return err
		}
		// The owner keeps the system role; otherwise a company could lose
		// its only user with all permissions.
		if current.IsSystem {
			return domain.ErrSystemRole
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
func (s *service) editableRole(ctx context.Context, roleId, companyId string) (*domain.Role, error) {
	role, err := s.repo.GetById(ctx, roleId, companyId)
	if err != nil {
		return nil, err
	}
	if role.IsSystem {
		return nil, domain.ErrSystemRole
	}
	return role, nil
}

// checkRolePermissionChange allows changing a role package only for
// permissions the grantor holds, so nobody can build a role wider than
// themselves.
func (s *service) checkRolePermissionChange(ctx context.Context, rq domain.RolePermissionRequest) error {
	if _, err := s.editableRole(ctx, rq.RoleId, rq.CompanyId); err != nil {
		return err
	}

	has, err := s.repo.UserHasPermission(ctx, rq.GrantorId, rq.PermissionId)
	if err != nil {
		return err
	}
	if !has {
		return domain.ErrPermissionNotOwned
	}
	return nil
}

func roleName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > maxRoleNameLen {
		return "", domain.ErrInvalidRoleName
	}
	return name, nil
}
