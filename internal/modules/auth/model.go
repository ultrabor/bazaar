package auth

import "errors"

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

var ErrPhoneTaken = errors.New("phone already registered")
var ErrInvalidPhone = errors.New("invalid phone")
var ErrInvalidCred = errors.New("invalid credentials")
var ErrUnauthorized = errors.New("unauthorized")
