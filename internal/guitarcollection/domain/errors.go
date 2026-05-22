package domain

import "errors"

// ErrGuitarNotFound is returned by a Repository when no guitar exists for the
// requested identifier. Application/interface layers translate this into the
// appropriate transport-level signal (e.g. HTTP 404).
var ErrGuitarNotFound = errors.New("guitar not found")

// ValidationError signals that a Guitar (or one of its value objects) violates
// an invariant of the GuitarCollection domain. Validation errors are part of
// the domain language and are surfaced to API clients as 400 Bad Request.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return e.Field + ": " + e.Message
}

func newValidationError(field, message string) *ValidationError {
	return &ValidationError{Field: field, Message: message}
}

// IsValidationError reports whether err (or any error it wraps) is a
// *ValidationError. It is the canonical way for outer layers to discriminate
// between domain validation failures and infrastructure failures.
func IsValidationError(err error) bool {
	var v *ValidationError
	return errors.As(err, &v)
}
