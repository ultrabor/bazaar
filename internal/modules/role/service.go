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
}

var _ domain.RoleService = (*service)(nil)

func NewService(repo domain.RoleRepository) domain.RoleService {
	return &service{repo: repo}
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
