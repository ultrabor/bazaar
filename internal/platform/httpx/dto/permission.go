package dto

type ChangePermissionRequest struct {
	UserId       string `json:"user_id"`
	PermissionId string `json:"permission_id"`
}
