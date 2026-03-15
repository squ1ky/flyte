package domain

import "errors"

var (
	ErrFlightNotFound   = errors.New("flight not found")
	ErrAirportNotFound  = errors.New("airport not found")
	ErrAircraftNotFound = errors.New("aircraft not found")
	ErrAirlineNotFound  = errors.New("airline not found")

	ErrFlightAlreadyExists   = errors.New("flight already exists")
	ErrAircraftAlreadyExists = errors.New("aircraft already exists")
	ErrAirlineAlreadyExists  = errors.New("airline with this IATA code already exists")

	ErrSeatUnavailable     = errors.New("seat is already reserved or sold")
	ErrReservationExpired  = errors.New("reservation is expired or does not exist")
	ErrReservationNotFound = errors.New("reservation not found")

	ErrInvalidFlightID       = errors.New("invalid flight id")
	ErrInvalidAircraftID     = errors.New("invalid aircraft id")
	ErrInvalidAirlineID      = errors.New("invalid airline id")
	ErrInvalidAircraftModel  = errors.New("aircraft model is required")
	ErrInvalidTotalSeats     = errors.New("total seats must be positive")
	ErrEmptySearchQuery      = errors.New("search query is required")
	ErrInvalidIATACode       = errors.New("IATA code is required")
	ErrInvalidAirlineName    = errors.New("airline name is required")
	ErrEmptySeatsList        = errors.New("seats list is empty")
	ErrEmptyBookingID        = errors.New("booking id is required")
	ErrEmptyFlightNumber     = errors.New("flight number is required")
	ErrInvalidAirportCode    = errors.New("airport code is required")
	ErrSameAirports          = errors.New("departure and arrival airports must differ")
	ErrInvalidFlightTime     = errors.New("arrival time must be after departure time")
	ErrInvalidSearchCriteria = errors.New("origin and destination are required")
	ErrInvalidSearchDate     = errors.New("search date is required")
	ErrInvalidPassengerCount = errors.New("passenger count must be positive")
)
