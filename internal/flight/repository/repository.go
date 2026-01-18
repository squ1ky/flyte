package repository

import (
	"context"
	"github.com/squ1ky/flyte/internal/flight/domain"
	"time"
)

type SearchFilter struct {
	FromAirport    string
	ToAirport      string
	Date           time.Time
	PassengerCount int
}

type FlightStorage interface {
	CreateFlight(ctx context.Context, flight *domain.Flight) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Flight, error)
	DeleteFlight(ctx context.Context, id int64) error

	GetSeatsByFlightID(ctx context.Context, flightID int64) ([]domain.Seat, error)
	BookSeat(ctx context.Context, flightID int64, seatNumber string) (int64, error)
	ReleaseSeat(ctx context.Context, flightID int64, seatNumber string) error
	ConfirmSeat(ctx context.Context, flightID int64, seatNumber string) error

	GetAirports(ctx context.Context) ([]domain.Airport, error)
}

type AircraftStorage interface {
	CreateAircraft(ctx context.Context, model string, totalSeats int) (int64, error)
	GetAircraftByID(ctx context.Context, aircraftID int64) (*domain.Aircraft, error)
	GetAircrafts(ctx context.Context) ([]domain.Aircraft, error)

	AddAircraftSeats(ctx context.Context, aircraftID int64, seats []domain.AircraftSeat) error
	GetAircraftSeats(ctx context.Context, aircraftID int64) ([]domain.AircraftSeat, error)
}

type FlightSearcher interface {
	Search(ctx context.Context, filter SearchFilter) ([]domain.Flight, error)
	IndexFlight(ctx context.Context, flight *domain.Flight) error
	UpdateAvailableSeats(ctx context.Context, flightID int64, newCount int) error
}
