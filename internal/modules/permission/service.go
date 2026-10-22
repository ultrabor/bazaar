package permission

import (
	"context"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetListByUserId(ctx context.Context, userId string) ([]Permission, error) {

	res, err := s.repo.GetListByUserId(ctx, userId)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Service) GetPermissionByCode(ctx context.Context, userId, code string) (*Permission, error) {

	res, err := s.repo.GetPermissionByCode(ctx, userId, code)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Service) SetPermission(ctx context.Context, rq ChangePermissionRequest) error {

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
		return ErrInvalidCred
	}

	return s.repo.SetPermission(ctx, rq.ApplyUserId, rq.PermissionId)
}

func (s *Service) RemovePermission(ctx context.Context, rq ChangePermissionRequest) error {

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
		return ErrInvalidCred
	}

	return s.repo.RemovePermission(ctx, rq.ApplyUserId, rq.PermissionId)
}
