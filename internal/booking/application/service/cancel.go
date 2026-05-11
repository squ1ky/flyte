package service

import (
	"context"
	"fmt"
	"github.com/squ1ky/flyte/internal/booking/domain"
	"log/slog"
)

type CancelService struct {
	bookings domain.BookingRepository
	flights  FlightClient
	logger   *slog.Logger
}

func NewCancelService(
	bookings domain.BookingRepository,
	flights FlightClient,
	logger *slog.Logger,
) *CancelService {
	return &CancelService{
		bookings: bookings,
		flights:  flights,
		logger:   logger,
	}
}

func (s *CancelService) CancelBooking(ctx context.Context, bookingID string) error {
	log := s.logger.With(slog.String("booking_id", bookingID))

	booking, err := s.bookings.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.Status == domain.BookingStatusCancelled {
		return domain.ErrBookingAlreadyCancelled
	}
	if booking.Status == domain.BookingStatusPaid {
		return domain.ErrBookingAlreadyPaid
	}

	if err := s.bookings.UpdateStatus(ctx, bookingID, domain.BookingStatusCancelled); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	if err := s.flights.CancelReservation(ctx, bookingID); err != nil {
		log.ErrorContext(ctx, "compensation failed: cancel reservation",
			slog.Any("error", err),
		)
	}

	log.InfoContext(ctx, "booking cancelled")
	return nil
}
