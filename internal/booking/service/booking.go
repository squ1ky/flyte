package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/squ1ky/flyte/internal/booking/clients/grpc/flight"
	"github.com/squ1ky/flyte/internal/booking/domain"
	"github.com/squ1ky/flyte/internal/booking/domain/events"
	"github.com/squ1ky/flyte/internal/booking/kafka"
	"github.com/squ1ky/flyte/internal/booking/repository"
	"log/slog"
)

type BookingService struct {
	repo         repository.BookingRepository
	producer     *kafka.PaymentEventProducer
	flightClient *flight.Client
	log          *slog.Logger
}

func NewBookingService(
	repo repository.BookingRepository,
	producer *kafka.PaymentEventProducer,
	flightClient *flight.Client,
	log *slog.Logger,
) *BookingService {
	return &BookingService{
		repo:         repo,
		producer:     producer,
		flightClient: flightClient,
		log:          log,
	}
}

type CreateBookingDTO struct {
	UserID            int64
	FlightID          int64
	SeatNumber        string
	PriceCents        int64
	Currency          string
	PassengerName     string
	PassengerPassport string
}

func (s *BookingService) CreateBooking(ctx context.Context, dto CreateBookingDTO) (string, error) {
	log := s.log.With("user_id", dto.UserID, "flight_id", dto.FlightID)

	if err := s.flightClient.ReserveSeat(ctx, dto.FlightID, dto.SeatNumber); err != nil {
		log.Error("failed to reserve seat", "error", err)
		return "", fmt.Errorf("failed to reserve seat: %w", err)
	}

	booking := &domain.Booking{
		UserID:            dto.UserID,
		FlightID:          dto.FlightID,
		SeatNumber:        dto.SeatNumber,
		PriceCents:        dto.PriceCents,
		Currency:          dto.Currency,
		PassengerName:     dto.PassengerName,
		PassengerPassport: dto.PassengerPassport,
		Status:            domain.StatusPending,
	}

	id, err := s.repo.Create(ctx, booking)
	if err != nil {
		log.Error("failed to create booking, releasing seat", "error", err)
		if releaseErr := s.flightClient.ReleaseSeat(ctx, dto.FlightID, dto.SeatNumber); releaseErr != nil {
			log.Error("failed to release seat during rollback", "error", releaseErr)
		}
		return "", fmt.Errorf("failed to create booking: %w", err)
	}

	return id, nil
}

func (s *BookingService) GetBooking(ctx context.Context, id string) (*domain.Booking, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *BookingService) ListBookings(ctx context.Context, userID int64) ([]domain.Booking, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *BookingService) CancelBooking(ctx context.Context, bookingID string) error {
	log := s.log.With("booking_id", bookingID)

	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		log.Error("failed to fetch booking: %w", err)
		return fmt.Errorf("failed to fetch booking: %w", err)
	}

	if booking.Status.IsTerminal() {
		log.Warn("cannot cancel booking in terminal status", "status", booking.Status)
		return fmt.Errorf("cannot cancel booking with status %s", booking.Status)
	}

	if err := s.repo.UpdateStatus(ctx, bookingID, domain.StatusCancelled); err != nil {
		log.Error("failed to update status to cancelled", "error", err)
		return fmt.Errorf("failed to update status to cancelled: %w", err)
	}

	if err := s.flightClient.ReleaseSeat(ctx, booking.FlightID, booking.SeatNumber); err != nil {
		log.Error("failed to release seat during cancellation", "error", err)
		return fmt.Errorf("booking cancelled locally but failed to release seat in flight-service: %w", err)
	}

	log.Info("booking cancelled successfully")
	return nil
}

func (s *BookingService) ProcessPaymentResult(ctx context.Context, bookingID string, status events.PaymentStatus) error {
	log := s.log.With("booking_id", bookingID, "status", status)

	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, domain.ErrBookingNotFound) {
			log.Error("booking not found for paymentResult")
			return nil
		}
		return fmt.Errorf("failed to get booking: %w", err)
	}

	switch status {
	case events.PaymentStatusSuccess:
		if booking.Status.IsTerminal() {
			log.Info("booking already in terminal state", "current_status", booking.Status)
			return nil
		}

		err = s.repo.UpdateStatus(ctx, bookingID, domain.StatusPaid)
		if err != nil {
			log.Warn("failed to update local status to PAID", "error", err)
			return fmt.Errorf("failed to update status: %w", err)
		}
		log.Info("booking successfully confirmed and paid")

		err = s.flightClient.ConfirmSeat(ctx, booking.FlightID, booking.SeatNumber)
		if err != nil {
			log.Error("status updated to PAID but failed to confirm seat", "error", err)
		}
	case events.PaymentStatusFailed:
		log.Info("payment failed, cancelling booking")

		err := s.repo.UpdateStatus(ctx, bookingID, domain.StatusFailed)
		if err != nil {
			log.Warn("booking cancellation skipped", "error", err)
		}

		err = s.flightClient.ReleaseSeat(ctx, booking.FlightID, booking.SeatNumber)
		if err != nil {
			log.Warn("failed to release seat", "error", err)
		}
	}

	return nil
}
