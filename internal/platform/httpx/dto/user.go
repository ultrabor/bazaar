package dto

type CreateUserRequest struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Phone      string `json:"phone"`
	Password   string `json:"password"`
	UserRoleId string `json:"user_role_id"`
}
