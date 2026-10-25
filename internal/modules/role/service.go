package role

import (
	"bazaar/internal/domain"
	"context"
	"errors"
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

func (s *service) GivePermissions(ctx context.Context, userId, roleId string) error {
	return s.repo.GivePermissions(ctx, userId, roleId)
}
