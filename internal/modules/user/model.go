package user

import (
	"errors"
)

type User struct {
	Id           string `json:"id"`
	CompanyId    string `json:"company_id"`
	UserRoleId   string `json:"user_role_id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Phone        string `json:"phone"`
	PasswordHash string `json:"-"`
}

type CreateUserRequest struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Phone        string `json:"phone"`
	Password     string `json:"password"`
	PasswordHash string `json:"-"`
	CompanyId    string `json:"company_id"`
	UserRoleId   string `json:"user_role_id"`
	// CreatorId is the user from the token; the new user's role may not
	// hold permissions the creator lacks.
	CreatorId string `json:"-"`
}

type CreateUserResponse struct {
	UserID string `json:"user_id"`
}

var (
	ErrUserNotFound = errors.New("user not found")
	ErrPhoneTaken   = errors.New("phone already registered")
	ErrInvalidPhone = errors.New("invalid phone")
	ErrInvalidRole  = errors.New("invalid role for company")
)
