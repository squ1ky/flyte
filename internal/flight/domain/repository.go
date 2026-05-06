package domain

import (
	"context"
	"time"
)

type FlightStorage interface {
	Create(ctx context.Context, flight *Flight, fares []FlightFare) (int64, error)
	UpdateStatus(ctx context.Context, flightID int64, status FlightStatus) error
	GetByID(ctx context.Context, flightID int64) (*FlightDetails, error)
	GetFaresByFlightID(ctx context.Context, flightID int64) ([]FlightFare, error)
	List(ctx context.Context, limit, offset int) ([]FlightDocument, error)
}

type SeatReservationParams struct {
	FlightID    int64
	SeatNumbers []string
	BookingID   string
}

func (p SeatReservationParams) Validate() error {
	if p.FlightID <= 0 {
		return ErrInvalidFlightID
	}
	if p.BookingID == "" {
		return ErrEmptyBookingID
	}
	if len(p.SeatNumbers) == 0 {
		return ErrEmptySeatsList
	}
	return nil
}

type ReservationStorage interface {
	ReserveSeats(ctx context.Context, params SeatReservationParams) error
	ConfirmReservation(ctx context.Context, params SeatReservationParams) error
	CancelReservation(ctx context.Context, bookingID string) error
	GetReservedSeats(ctx context.Context, flightID int64) ([]SeatReservation, error)
	CleanExpiredReservations(ctx context.Context) (int64, error)
}

type AircraftStorage interface {
	Create(ctx context.Context, aircraft *Aircraft) (int64, error)
	GetByID(ctx context.Context, aircraftID int64) (*Aircraft, error)
	List(ctx context.Context, limit, offset int) ([]Aircraft, error)

	AddSeats(ctx context.Context, aircraftID int64, seats []AircraftSeat) error
	GetSeatsByAircraftID(ctx context.Context, aircraftID int64) ([]AircraftSeat, error)
}

type AirportStorage interface {
	GetByCode(ctx context.Context, code string) (*Airport, error)
	SearchAirports(ctx context.Context, searchQuery string) ([]Airport, error)
}

type AirlineStorage interface {
	Create(ctx context.Context, airline *Airline) (int64, error)
	GetByID(ctx context.Context, airlineID int64) (*Airline, error)
	ListAirlines(ctx context.Context, limit, offset int) ([]Airline, error)
}

type FlightSearchCriteria struct {
	Origin        string // City or code
	Destination   string // City or code
	Date          time.Time
	SeatClass     SeatClass
	NumPassengers int
}

func (c FlightSearchCriteria) Validate() error {
	if c.Origin == "" || c.Destination == "" {
		return ErrInvalidSearchCriteria
	}
	if c.Date.IsZero() {
		return ErrInvalidSearchDate
	}
	if c.NumPassengers <= 0 {
		return ErrInvalidPassengerCount
	}
	return nil
}

type FlightDocument struct {
	ID             int64        `json:"id"`
	FlightNumber   string       `json:"flight_number"`
	DepartureTime  time.Time    `json:"departure_time"`
	ArrivalTime    time.Time    `json:"arrival_time"`
	MinPriceCents  int64        `json:"min_price_cents"`
	AvailableSeats int          `json:"available_seats"`
	Status         FlightStatus `json:"status"`

	Airline   Airline `json:"airline"`
	Departure Airport `json:"departure"`
	Arrival   Airport `json:"arrival"`
}

type FlightSearcher interface {
	Search(ctx context.Context, criteria FlightSearchCriteria) ([]FlightDocument, error)
	IndexFlight(ctx context.Context, flight *FlightDocument) error
}
