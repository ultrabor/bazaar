package permission

import "errors"

type Permission struct {
	Id    string `json:"id"`
	Label string `json:"label"`
	Code  string `json:"code"`
}

type ChangePermissionRequest struct {
	GuarantorUserId string `json:"guarantor_user_id"`
	ApplyUserId     string `json:"apply_user_id"`
	PermissionId    string `json:"permission_id"`
}

var ErrNotFound = errors.New("no rows in result set")
var ErrPermissionAlreadyExists = errors.New("permission already exists")
var ErrInvalidCred = errors.New("dont have permission to set")
