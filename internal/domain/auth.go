package domain

import (
	"context"
	"errors"
)

type RegisterRequest struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	CompanyName string `json:"company_name"`
	Phone       string `json:"phone"`
	Password    string `json:"password"`
}

type RegisterResponse struct {
	CompanyId string `json:"company_id"`
	UserId    string `json:"user_id"`
}

type LoginRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// Auth also returns ErrInvalidCred, ErrInvalidPhone and ErrPhoneTaken.
var ErrUnauthorized = errors.New("unauthorized")

type AuthRepository interface {
	CreateCompany(ctx context.Context, name string) (string, error)
	// CreateUser returns ErrPhoneTaken when an active user has the phone.
	CreateUser(ctx context.Context, companyId, roleId, firstName, lastName, phone, passwordHash string) (string, error)
	// GetUserByPhone and GetUserById return ErrInvalidCred for a missing user.
	GetUserByPhone(ctx context.Context, phone string) (*User, error)
	GetUserById(ctx context.Context, userId string) (*User, error)
}

type AuthService interface {
	// RegisterOwner creates the company, its owner role with all permissions
	// and the owner user in one transaction.
	RegisterOwner(ctx context.Context, rq RegisterRequest) (*RegisterResponse, error)
	Login(ctx context.Context, rq LoginRequest) (*LoginResponse, error)
	GetUserById(ctx context.Context, userId string) (*User, error)
}
