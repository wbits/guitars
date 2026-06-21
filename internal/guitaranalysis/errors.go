package guitaranalysis

import "errors"

// ValidationError indicates invalid analysis input.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + " " + e.Message
}

func InvalidField(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}

func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}

// ErrBYOKNotConfigured indicates the owner has no assistant API key stored.
var ErrBYOKNotConfigured = errors.New("assistant BYOK is not configured")

// ErrPhotoAnalysisDisabled indicates the owner has not opted in to photo analysis.
var ErrPhotoAnalysisDisabled = errors.New("photo analysis is not enabled")
