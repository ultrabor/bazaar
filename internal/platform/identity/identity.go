// Package identity carries the authenticated user through a request.
// It belongs to the platform so that middleware and module handlers share
// the type without depending on any module.
package identity

import "context"

// User is the authenticated caller. Auth middleware loads it once per
// request; handlers read it to scope work to the caller's company.
type User struct {
	Id         string `json:"id"`
	CompanyId  string `json:"company_id"`
	UserRoleId string `json:"user_role_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Phone      string `json:"phone"`
}

type contextKey uint8

const currentUserKey contextKey = iota

// WithUser returns a copy of ctx that carries u. Auth middleware uses it
// after loading the user; tests use it to skip token checks.
func WithUser(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, currentUserKey, u)
}

func FromContext(ctx context.Context) (*User, bool) {
	if ctx == nil {
		return nil, false
	}

	user, ok := ctx.Value(currentUserKey).(*User)
	if !ok {
		return nil, false
	}

	return user, true
}
