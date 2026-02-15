package domain

import "errors"

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrPassengerAlreadyExists = errors.New("passenger already exists")
	ErrPassengerNotFound      = errors.New("passenger not found")
)
