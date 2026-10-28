package apperror

import "errors"

// ErrInvalidCred means required input is missing or malformed. Every
// module returns it, so it lives in the platform instead of a module.
var ErrInvalidCred = errors.New("invalid credentials")

// ErrForbidden means the caller lacks the permission an action requires.
var ErrForbidden = errors.New("forbidden")
