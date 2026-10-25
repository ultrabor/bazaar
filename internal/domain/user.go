package domain

import (
	"context"
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
}

type CreateUserResponse struct {
	UserID string `json:"user_id"`
}

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrInvalidUserID = errors.New("invalid user id")
	ErrPhoneTaken   = errors.New("phone already registered")
	ErrInvalidPhone = errors.New("invalid phone")
	ErrInvalidRole  = errors.New("invalid role for company")
)

type UserRepository interface {
	GetUserById(ctx context.Context, userId, companyId string) (*User, error)
	// CreateUser expects req.PasswordHash to be set. Returns ErrPhoneTaken
	// when an active user with the same phone exists.
	CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error)
}

type UserService interface {
	GetUserById(ctx context.Context, userId, companyId string) (*User, error)
	CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error)
}
