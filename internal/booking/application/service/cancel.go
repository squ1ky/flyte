package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/squ1ky/flyte/internal/booking/domain"
	"github.com/squ1ky/flyte/internal/booking/infrastructure/outbox"
)

type CancelService struct {
	bookings domain.BookingRepository
	flights  FlightClient
	outbox   outbox.Repository
	logger   *slog.Logger
}

func NewCancelService(
	bookings domain.BookingRepository,
	flights FlightClient,
	outbox outbox.Repository,
	logger *slog.Logger,
) *CancelService {
	return &CancelService{
		bookings: bookings,
		flights:  flights,
		outbox:   outbox,
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

	reason := domain.CancelReasonUserCancelled
	if err := s.bookings.UpdateStatus(ctx, bookingID, domain.BookingStatusCancelled, &reason); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	if err := s.publishBookingCancelled(ctx, bookingID, reason); err != nil {
		log.ErrorContext(ctx, "publish booking cancelled event failed",
			slog.Any("error", err),
		)
	}

	if err := s.flights.CancelReservation(ctx, bookingID); err != nil {
		log.ErrorContext(ctx, "compensation failed: cancel reservation",
			slog.Any("error", err),
		)
	}

	log.InfoContext(ctx, "booking cancelled")
	return nil
}

func (s *CancelService) publishBookingCancelled(ctx context.Context, bookingID string, reason domain.CancelReason) error {
	event := domain.BookingCancelledEvent{
		BookingID:   bookingID,
		Reason:      reason,
		CancelledAt: time.Now(),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal booking cancelled event: %w", err)
	}

	return s.outbox.Insert(ctx, outbox.Event{
		ID:        uuid.New().String(),
		EventType: outbox.EventBookingCancelled,
		Payload:   payload,
		Status:    outbox.StatusPending,
		CreatedAt: time.Now(),
	})
}
