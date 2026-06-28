package shared

import (
	"errors"
	"fmt"
)

// ErrNotFound is a sentinel error for lookups that find no matching row.
var ErrNotFound = errors.New("not found")

// ErrValidation is a sentinel error for query validation failures.
var ErrValidation = errors.New("validation error")

// ValidationError is a custom error type for validation failures that can be
// detected with errors.Is(err, ErrValidation) but doesn't include "validation error"
// in the message.
type ValidationError struct {
	msg string
}

func (e *ValidationError) Error() string {
	return e.msg
}

func (e *ValidationError) Is(target error) bool {
	return target == ErrValidation
}

func (e *ValidationError) Unwrap() error {
	return ErrValidation
}

// ValidationErrorf creates a ValidationError with a formatted message.
// go vet automatically detects this as a printf wrapper via fmt.Sprintf.
func ValidationErrorf(format string, args ...any) error {
	return &ValidationError{msg: fmt.Sprintf(format, args...)}
}
