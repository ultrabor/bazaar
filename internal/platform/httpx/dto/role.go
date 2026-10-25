package dto

type CreateRoleRequest struct {
	Name string `json:"name"`
}

type UpdateRoleRequest struct {
	Name string `json:"name"`
}

type AddRolePermissionRequest struct {
	PermissionId string `json:"permission_id"`
}

type AssignRoleRequest struct {
	RoleId string `json:"role_id"`
}
