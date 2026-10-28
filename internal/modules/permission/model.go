package permission

import (
	"errors"
)

type Permission struct {
	Id    string `json:"id"`
	Label string `json:"label"`
	Code  string `json:"code"`
}

// ChangePermissionRequest grants or revokes a permission. The guarantor must
// hold the permission, and the apply user must belong to CompanyId.
type ChangePermissionRequest struct {
	GuarantorUserId string `json:"guarantor_user_id"`
	CompanyId       string `json:"company_id"`
	ApplyUserId     string `json:"apply_user_id"`
	PermissionId    string `json:"permission_id"`
}

var (
	ErrPermissionNotFound      = errors.New("permission not found")
	ErrPermissionAlreadyExists = errors.New("permission already exists")
	ErrPermissionNotOwned      = errors.New("dont have permission to set")
	ErrInvalidUserID           = errors.New("invalid user id")
)
