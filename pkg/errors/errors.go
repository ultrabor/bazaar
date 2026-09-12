package errors

import (
	"errors"
	"fmt"
)

type DependencyError struct {
	Dependency string
	Err        error
}

func New(msg string, err error) *DependencyError {
	return &DependencyError{Dependency: msg, Err: err}
}

func (e *DependencyError) Error() string {
	return fmt.Sprintf("%s %v", e.Dependency, e.Err)
}

func (e *DependencyError) Unwrap() error {
	return errors.Unwrap(e.Err)
}
