package service

import (
	"context"
	"fmt"
	"github.com/squ1ky/flyte/internal/flight/domain"
	"log/slog"
)

type ReservationService struct {
	repo   domain.ReservationStorage
	logger *slog.Logger
}

func NewReservationService(
	repo domain.ReservationStorage,
	logger *slog.Logger,
) *ReservationService {
	return &ReservationService{
		repo:   repo,
		logger: logger,
	}
}

func (s *ReservationService) ReserveSeats(ctx context.Context, params domain.SeatReservationParams) error {
	if err := params.Validate(); err != nil {
		return err
	}

	if err := s.repo.ReserveSeats(ctx, params); err != nil {
		s.logger.ErrorContext(ctx, "failed to reserve seats",
			slog.Int64("flight_id", params.FlightID),
			slog.String("booking_id", params.BookingID),
			slog.Any("error", err),
		)
		return fmt.Errorf("reserve seats: %w", err)
	}
	return nil
}

func (s *ReservationService) ConfirmReservation(ctx context.Context, params domain.SeatReservationParams) error {
	if err := params.Validate(); err != nil {
		return err
	}

	if err := s.repo.ConfirmReservation(ctx, params); err != nil {
		s.logger.ErrorContext(ctx, "failed to confirm reservation",
			slog.Int64("flight_id", params.FlightID),
			slog.String("booking_id", params.BookingID),
			slog.Any("error", err),
		)
		return fmt.Errorf("confirm reservation: %w", err)
	}
	return nil
}

func (s *ReservationService) CancelReservation(ctx context.Context, bookingID string) error {
	if bookingID == "" {
		return domain.ErrEmptyBookingID
	}

	if err := s.repo.CancelReservation(ctx, bookingID); err != nil {
		s.logger.ErrorContext(ctx, "failed to cancel reservation",
			slog.String("booking_id", bookingID),
			slog.Any("error", err),
		)
		return fmt.Errorf("cancel reservation: %w", err)
	}
	return nil
}

func (s *ReservationService) GetReservedSeats(ctx context.Context, flightID int64) ([]domain.SeatReservation, error) {
	if flightID <= 0 {
		return nil, domain.ErrInvalidFlightID
	}

	seats, err := s.repo.GetReservedSeats(ctx, flightID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get reserved seats",
			slog.Int64("flight_id", flightID),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("get reserved seats: %w", err)
	}
	return seats, nil
}
