package auth

import (
	"errors"
	"time"
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

// LoginRequest carries the client's user agent and IP. The handler takes
// them from the HTTP request, never from the body.
type LoginRequest struct {
	Phone     string
	Password  string
	UserAgent string
	Ip        string
}

type LoginResponse struct {
	Token string `json:"token"`
}

type Session struct {
	Id        string     `json:"id"`
	UserId    string     `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at"`
	CreatedAt time.Time  `json:"-"`
	UserAgent *string    `json:"user_agent,omitempty"`
	Ip        *string    `json:"ip,omitempty"`
}

// User is a user as auth sees it: with the password hash for login.
type User struct {
	Id           string
	CompanyId    string
	UserRoleId   string
	FirstName    string
	LastName     string
	Phone        string
	PasswordHash string
}

// Auth also returns apperror.ErrInvalidCred. ErrInvalidPhone and
// ErrPhoneTaken are auth's own: registration writes the users table itself
// instead of calling the user module.
var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalidPhone = errors.New("invalid phone")
	ErrPhoneTaken   = errors.New("phone already registered")
)
