package grpcserver

import (
	"context"
	pb "github.com/squ1ky/flyte/gen/proto/flight"
	"github.com/squ1ky/flyte/internal/flight/domain"
)

type AircraftService interface {
	Create(ctx context.Context, aircraft *domain.Aircraft) (int64, error)
	GetByID(ctx context.Context, aircraftID int64) (*domain.Aircraft, error)
	AddSeats(ctx context.Context, aircraftID int64, seats []domain.AircraftSeat) error
	GetSeats(ctx context.Context, aircraftID int64) ([]domain.AircraftSeat, error)
}

type AirlineService interface {
	Create(ctx context.Context, airline *domain.Airline) (int64, error)
	GetByID(ctx context.Context, airlineID int64) (*domain.Airline, error)
	List(ctx context.Context, limit, offset int) ([]domain.Airline, error)
}

type AirportService interface {
	GetByCode(ctx context.Context, code string) (*domain.Airport, error)
	Search(ctx context.Context, query string) ([]domain.Airport, error)
}

type FlightService interface {
	Search(ctx context.Context, criteria domain.FlightSearchCriteria) ([]domain.FlightDocument, error)
	Create(ctx context.Context, flight *domain.Flight, fares []domain.FlightFare) (int64, error)
	UpdateStatus(ctx context.Context, flightID int64, status domain.FlightStatus) error
	GetByID(ctx context.Context, flightID int64) (*domain.Flight, error)
	GetFares(ctx context.Context, flightID int64) ([]domain.FlightFare, error)
}

type ReservationService interface {
	ReserveSeats(ctx context.Context, params domain.SeatReservationParams) error
	ConfirmReservation(ctx context.Context, params domain.SeatReservationParams) error
	CancelReservation(ctx context.Context, bookingID string) error
	GetReservedSeats(ctx context.Context, flightID int64) ([]domain.SeatReservation, error)
}

type Server struct {
	pb.UnimplementedFlightServiceServer

	aircraft    AircraftService
	airline     AirlineService
	airport     AirportService
	flight      FlightService
	reservation ReservationService
}

func NewServer(
	aircraft AircraftService,
	airline AirlineService,
	airport AirportService,
	flight FlightService,
	reservation ReservationService,
) *Server {
	return &Server{
		aircraft:    aircraft,
		airline:     airline,
		airport:     airport,
		flight:      flight,
		reservation: reservation,
	}
}
