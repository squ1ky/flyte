package domain

import "errors"

var (
	ErrFlightAlreadyExists = errors.New("flight already exists")
	ErrFlightNotFound      = errors.New("flight not found")
	ErrSeatNotFound        = errors.New("seat not found")
	ErrSeatAlreadyBooked   = errors.New("seat already booked")
)
