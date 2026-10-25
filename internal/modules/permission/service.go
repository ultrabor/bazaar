package permission

import (
	"bazaar/internal/domain"
	"context"
)

type service struct {
	repo domain.PermissionRepository
}

var _ domain.PermissionService = (*service)(nil)

func NewService(repo domain.PermissionRepository) domain.PermissionService {
	return &service{repo: repo}
}

func (s *service) GetListByUserId(ctx context.Context, userId string) ([]domain.Permission, error) {

	res, err := s.repo.GetListByUserId(ctx, userId)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *service) GetPermissionByCode(ctx context.Context, userId, code string) (*domain.Permission, error) {

	res, err := s.repo.GetPermissionByCode(ctx, userId, code)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *service) SetPermission(ctx context.Context, rq domain.ChangePermissionRequest) error {

	res, err := s.repo.GetListByUserId(ctx, rq.GuarantorUserId)

	if err != nil {
		return err
	}

	has := false
	for _, v := range res {
		if v.Id == rq.PermissionId {
			has = true
			break
		}
	}

	if !has {
		return domain.ErrPermissionNotOwned
	}

	return s.repo.SetPermission(ctx, rq.ApplyUserId, rq.PermissionId)
}

func (s *service) RemovePermission(ctx context.Context, rq domain.ChangePermissionRequest) error {

	res, err := s.repo.GetListByUserId(ctx, rq.GuarantorUserId)

	if err != nil {
		return err
	}

	has := false
	for _, v := range res {
		if v.Id == rq.PermissionId {
			has = true
			break
		}
	}

	if !has {
		return domain.ErrPermissionNotOwned
	}

	return s.repo.RemovePermission(ctx, rq.ApplyUserId, rq.PermissionId)
}
