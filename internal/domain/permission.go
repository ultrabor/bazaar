package domain

import (
	"context"
	"errors"
)

type Permission struct {
	Id    string `json:"id"`
	Label string `json:"label"`
	Code  string `json:"code"`
}

// ChangePermissionRequest grants or revokes a permission. The guarantor must
// hold the permission to pass it to or take it from the apply user.
type ChangePermissionRequest struct {
	GuarantorUserId string `json:"guarantor_user_id"`
	ApplyUserId     string `json:"apply_user_id"`
	PermissionId    string `json:"permission_id"`
}

var (
	ErrPermissionNotFound      = errors.New("permission not found")
	ErrPermissionAlreadyExists = errors.New("permission already exists")
	ErrPermissionNotOwned      = errors.New("dont have permission to set")
)

type PermissionRepository interface {
	GetListByUserId(ctx context.Context, userId string) ([]Permission, error)
	// GetPermissionByCode returns ErrPermissionNotFound when the user
	// does not have the permission.
	GetPermissionByCode(ctx context.Context, userId, code string) (*Permission, error)
	SetPermission(ctx context.Context, userId, permissionId string) error
	// RemovePermission returns ErrPermissionNotFound when nothing was removed.
	RemovePermission(ctx context.Context, userId, permissionId string) error
}

type PermissionService interface {
	// GetUserPermissions returns the viewer's own permissions when userId is
	// empty or the viewer's id. For another user the viewer needs
	// permissions.view (else ErrForbidden) and the user must belong to
	// companyId (else ErrUserNotFound).
	GetUserPermissions(ctx context.Context, viewerId, companyId, userId string) ([]Permission, error)
	GetPermissionByCode(ctx context.Context, userId, code string) (*Permission, error)
	// SetPermission and RemovePermission return ErrPermissionNotOwned when
	// the guarantor does not hold the permission.
	SetPermission(ctx context.Context, rq ChangePermissionRequest) error
	RemovePermission(ctx context.Context, rq ChangePermissionRequest) error
}
