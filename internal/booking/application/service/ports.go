package service

import (
	"context"
	"github.com/squ1ky/flyte/internal/booking/domain"
	"time"
)

// FlightClient is the subset of the gRPC flight client used by booking services.
type FlightClient interface {
	ReserveSeats(ctx context.Context, flightID int64, seatNumbers []string, bookingID string) (unavailableSeats []string, expiresAt time.Time, err error)
	ConfirmReservation(ctx context.Context, flightID int64, seatNumbers []string, bookingID string) error
	CancelReservation(ctx context.Context, bookingID string) error
	GetFlightFares(ctx context.Context, flightID int64) (map[domain.SeatClass]int64, error)
}
