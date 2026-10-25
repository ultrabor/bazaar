package role

import "context"

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
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetList(ctx context.Context, companyId string) ([]Role, error) {
	return s.repo.GetList(ctx, companyId)
}

func (s *Service) GetById(ctx context.Context, userId string) ([]Role, error)

func (s *Service) CreateRole(ctx context.Context, userId string) ([]Role, error)
