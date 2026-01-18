package domain

import "errors"

var (
	ErrFlightNotFound      = errors.New("flight not found")
	ErrFlightAlreadyExists = errors.New("flight already exists")

	ErrSeatNotFound      = errors.New("seat not found")
	ErrSeatAlreadyBooked = errors.New("seat already booked")

	ErrAircraftNotFound = errors.New("aircraft not found")
)
