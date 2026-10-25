package permission

import (
	"bazaar/internal/domain"
	"context"
	"errors"

	"github.com/google/uuid"
)

// codeView lets a user see the permissions of other users in the company.
const codeView = "permissions.view"

type service struct {
	repo  domain.PermissionRepository
	users domain.UserService
}

var _ domain.PermissionService = (*service)(nil)

func NewService(repo domain.PermissionRepository, users domain.UserService) domain.PermissionService {
	return &service{repo: repo, users: users}
}

func (s *service) GetUserPermissions(ctx context.Context, viewerId, companyId, userId string) ([]domain.Permission, error) {
	if userId == "" || userId == viewerId {
		return s.repo.GetListByUserId(ctx, viewerId)
	}

	if _, err := uuid.Parse(userId); err != nil {
		return nil, domain.ErrInvalidUserID
	}

	if _, err := s.repo.GetPermissionByCode(ctx, viewerId, codeView); err != nil {
		if errors.Is(err, domain.ErrPermissionNotFound) {
			return nil, domain.ErrForbidden
		}
		return nil, err
	}

	// GetUserById filters by company, so a user of another company is
	// reported as not found.
	if _, err := s.users.GetUserById(ctx, userId, companyId); err != nil {
		return nil, err
	}

	return s.repo.GetListByUserId(ctx, userId)
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
