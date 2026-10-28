package permission

import (
	"bazaar/internal/platform/support/apperror"
	"bazaar/internal/user"
	"context"
	"errors"

	"github.com/google/uuid"
)

// codeView lets a user see the permissions of other users in the company.
const codeView = "permissions.view"

type Service struct {
	repo  repository
	users userService
}

func NewService(repo repository, users userService) *Service {
	return &Service{repo: repo, users: users}
}

func (s *Service) GetUserPermissions(ctx context.Context, viewerId, companyId, userId string) ([]Permission, error) {
	if userId == "" || userId == viewerId {
		return s.repo.GetListByUserId(ctx, viewerId)
	}

	if _, err := uuid.Parse(userId); err != nil {
		return nil, ErrInvalidUserID
	}

	if _, err := s.repo.GetPermissionByCode(ctx, viewerId, codeView); err != nil {
		if errors.Is(err, ErrPermissionNotFound) {
			return nil, apperror.ErrForbidden
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

// HasPermission reports whether the user holds the permission code.
func (s *Service) HasPermission(ctx context.Context, userId, code string) (bool, error) {
	_, err := s.repo.GetPermissionByCode(ctx, userId, code)
	if errors.Is(err, ErrPermissionNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) SetPermission(ctx context.Context, rq ChangePermissionRequest) error {
	if err := s.checkTarget(ctx, rq); err != nil {
		return err
	}

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
		return ErrPermissionNotOwned
	}

	return s.repo.SetPermission(ctx, rq.ApplyUserId, rq.PermissionId)
}

func (s *Service) RemovePermission(ctx context.Context, rq ChangePermissionRequest) error {
	if err := s.checkTarget(ctx, rq); err != nil {
		return err
	}

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
		return ErrPermissionNotOwned
	}

	return s.repo.RemovePermission(ctx, rq.ApplyUserId, rq.PermissionId)
}

// checkTarget makes sure the apply user is a valid id of the guarantor's
// company, so permissions cannot be changed across companies.
func (s *Service) checkTarget(ctx context.Context, rq ChangePermissionRequest) error {
	if _, err := uuid.Parse(rq.ApplyUserId); err != nil {
		return ErrInvalidUserID
	}

	_, err := s.users.GetUserById(ctx, rq.ApplyUserId, rq.CompanyId)
	return err
}

// userService finds a user of a company; it returns user.ErrUserNotFound
// for a user of another company.
type userService interface {
	GetUserById(ctx context.Context, userId, companyId string) (*user.User, error)
}

// repository is the storage Service needs; Repository implements it.
type repository interface {
	GetListByUserId(ctx context.Context, userId string) ([]Permission, error)
	// GetPermissionByCode returns ErrPermissionNotFound when the user
	// does not have the permission.
	GetPermissionByCode(ctx context.Context, userId, code string) (*Permission, error)
	SetPermission(ctx context.Context, userId, permissionId string) error
	// RemovePermission returns ErrPermissionNotFound when nothing was removed.
	RemovePermission(ctx context.Context, userId, permissionId string) error
}
