package domain

import (
	"errors"
)

var (
	ErrBookingNotFound         = errors.New("booking not found")
	ErrBookingAlreadyPaid      = errors.New("booking is already paid")
	ErrBookingAlreadyCancelled = errors.New("booking is already cancelled")
	ErrBookingExpired          = errors.New("booking has expired")
)

var (
	ErrSeatsUnavailable = errors.New("one or more seats are unavailable")
)

var (
	ErrNoPassengers      = errors.New("at least one passenger is required")
	ErrInvalidFlightID   = errors.New("flight_id is required")
	ErrInvalidUserID     = errors.New("user_id is required")
	ErrInvalidEmail      = errors.New("contact_email is required")
	ErrInvalidSeatNumber = errors.New("seat_number is required for each passenger")
	ErrInvalidSeatClass  = errors.New("seat_class is invalid")
	ErrInvalidDocument   = errors.New("document_number is required for each passenger")
	ErrInvalidName       = errors.New("first_name and last_name are required for each passenger")
)
