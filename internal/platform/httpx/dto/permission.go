package dto

type ChangePermissionRequest struct {
	UserId       string `json:"user_id"`
	PermissionId string `json"json:"permission_id"`
}
