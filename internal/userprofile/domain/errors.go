package domain

import "errors"

var ErrUsernameTaken = errors.New("username is already taken")

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

func InvalidField(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}

func IsValidationError(err error) bool {
	var v *ValidationError
	return errors.As(err, &v)
}
