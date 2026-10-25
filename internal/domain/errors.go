package domain

import "errors"

// ErrInvalidCred means required input is missing or malformed. Several
// modules return it, so it lives here instead of a module file.
var ErrInvalidCred = errors.New("invalid credentials")

// ErrForbidden means the caller lacks the permission an action requires.
var ErrForbidden = errors.New("forbidden")
