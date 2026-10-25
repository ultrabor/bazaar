package role

import (
	"bazaar/internal/modules/permission"
	"errors"
)

type Role struct {
	Id          string                  `json:"id"`
	Name        string                  `json:"name"`
	CompanyId   string                  `json:"company_id"`
	IsSystem    string                  `json:"is_system"`
	Permissions []permission.Permission `json:"permissions"`
}

type GetListRoleResponse struct {
	Id               string `json:"id"`
	Name             string `json:"name"`
	CompanyId        string `json:"company_id"`
	IsSystem         string `json:"is_system"`
	PermissionsCount int    `json:"permissions_count"`
	UserCount        int    `json:"user_count"`
}

var ErrNotFound = errors.New("no rows in result set")
var ErrInvalidCred = errors.New("dont have permission to set")
