package user

import "errors"

type User struct {
	Id           string `json:"id"`
	CompanyId    string `json:"company_id"`
	UserRoleId   string `json:"user_role_id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Phone        string `json:"phone"`
	PasswordHash string `json:"-"`
}

type UserRole struct {
	Id        string `json:"id"`
	CompanyId string `json:"company_id"`
	Name      string `json:"name"`
}

type CreateUserRequest struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Phone        string `json:"phone"`
	Password     string `json:"password"`
	PasswordHash string `json:"-"`
	CompanyId    string `json:"company_id"`
	UserRoleId   string `json:"user_role_id"`
}

type CreateUserResponse struct {
	UserID string `json:"user_id"`
}

var ErrPhoneTaken = errors.New("phone already registered")
var ErrInvalidPhone = errors.New("invalid phone")
var ErrInvalidCred = errors.New("invalid credentials")
var ErrNotFound = errors.New("no rows in result set")
var ErrInvalidRole = errors.New("invalid role for company")
