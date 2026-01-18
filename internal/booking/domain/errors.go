package domain

import (
	"errors"
)

var (
	ErrBookingNotFound         = errors.New("booking not found")
	ErrSeatAlreadyBooked       = errors.New("seat is already booked")
	ErrBookingAlreadyPaid      = errors.New("booking is already paid")
	ErrBookingAlreadyCancelled = errors.New("booking is already cancelled")
	ErrInvalidBookingStatus    = errors.New("invalid booking status transition")
)
