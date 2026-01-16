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
	GetSeatsByFlightID(ctx context.Context, flightID int64) ([]domain.Seat, error)
	GetAirports(ctx context.Context) ([]domain.Airport, error)
	DeleteFlight(ctx context.Context, id int64) error

	BookSeat(ctx context.Context, flightID int64, seatNumber string) (int64, error)
	ReleaseSeat(ctx context.Context, flightID int64, seatNumber string) error
	ConfirmSeat(ctx context.Context, flightID int64, seatNumber string) error
}

type FlightSearcher interface {
	Search(ctx context.Context, filter SearchFilter) ([]domain.Flight, error)
	IndexFlight(ctx context.Context, flight *domain.Flight) error
	UpdateAvailableSeats(ctx context.Context, flightID int64, newCount int) error
}
