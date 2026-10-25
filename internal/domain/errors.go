package domain

import "errors"

// ErrInvalidCred means required input is missing or malformed. Several
// modules return it, so it lives here instead of a module file.
var ErrInvalidCred = errors.New("invalid credentials")
