package service

import (
	"context"
	"fmt"
	"github.com/squ1ky/flyte/internal/booking/clients/grpc/flight"
	"github.com/squ1ky/flyte/internal/booking/domain"
	"github.com/squ1ky/flyte/internal/booking/kafka"
	"github.com/squ1ky/flyte/internal/booking/repository"
	"log/slog"
)

type BookingService struct {
	repo         repository.BookingRepository
	producer     *kafka.BookingProducer
	flightClient *flight.Client
	log          *slog.Logger
}

func NewBookingService(
	repo repository.BookingRepository,
	producer *kafka.BookingProducer,
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
	Price             float64
	Currency          string
	PassengerName     string
	PassengerPassport string
}

func (s *BookingService) CreateBooking(ctx context.Context, dto CreateBookingDTO) (string, error) {
	log := s.log.With("user_id", dto.UserID, "flight_id", dto.FlightID)

	booking := &domain.Booking{
		UserID:            dto.UserID,
		FlightID:          dto.FlightID,
		SeatNumber:        dto.SeatNumber,
		Price:             dto.Price,
		Currency:          dto.Currency,
		PassengerName:     dto.PassengerName,
		PassengerPassport: dto.PassengerPassport,
		Status:            domain.StatusPending,
	}

	id, err := s.repo.Create(ctx, booking)
	if err != nil {
		return "", fmt.Errorf("failed to create booking: %w", err)
	}
	booking.ID = id

	if err := s.flightClient.ReserveSeat(ctx, dto.FlightID, dto.SeatNumber); err != nil {
		log.Error("failed to reserve seat", "error", err)
		_ = s.repo.UpdateStatus(ctx, id, domain.StatusFailed)
		return "", fmt.Errorf("failed to reserve seat: %w", err)
	}

	if err := s.producer.SendPaymentRequest(ctx, booking); err != nil {
		log.Error("failed to send payment request", "error", err)
		_ = s.flightClient.ReleaseSeat(ctx, dto.FlightID, dto.SeatNumber)
		_ = s.repo.UpdateStatus(ctx, id, domain.StatusFailed)
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

	err := s.repo.UpdateStatus(ctx, bookingID, domain.StatusCancelled)
	if err != nil {
		log.Warn("cancel booking status update skipped", "error", err)
	}

	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		log.Error("failed to fetch booking: %w", err)
	}

	if err := s.flightClient.ReleaseSeat(ctx, booking.FlightID, booking.SeatNumber); err != nil {
		log.Error("failed to release seat during cancellation", "error", err)
		return fmt.Errorf("failed to release seat: %w", err)
	}

	log.Info("booking cancelled successfully")
	return nil
}
