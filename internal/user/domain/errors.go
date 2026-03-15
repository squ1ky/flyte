package domain

import "errors"

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrPassengerAlreadyExists = errors.New("passenger already exists")
	ErrPassengerNotFound      = errors.New("passenger not found")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return e.Field + ": " + e.Message
	}
	return e.Message
}

func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}

func AsValidationError(err error) (*ValidationError, bool) {
	var ve *ValidationError
	ok := errors.As(err, &ve)
	return ve, ok
}

func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
