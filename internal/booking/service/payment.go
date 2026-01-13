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

type PaymentProcessor struct {
	repo         repository.BookingRepository
	flightClient *flight.Client
	log          *slog.Logger
}

func NewPaymentProcessor(
	repo repository.BookingRepository,
	flightClient *flight.Client,
	log *slog.Logger,
) *PaymentProcessor {
	return &PaymentProcessor{
		repo:         repo,
		flightClient: flightClient,
		log:          log,
	}
}

func (p *PaymentProcessor) ProcessResult(ctx context.Context, bookingID string, status kafka.PaymentStatus) error {
	log := p.log.With("booking_id", bookingID, "status", status)

	switch status {
	case kafka.PaymentStatusSuccess:
		err := p.repo.UpdateStatus(ctx, bookingID, domain.StatusPaid)
		if err != nil {
			log.Warn("booking update skipped (already processed or not found)")
			return nil
		}
		log.Info("booking confirmed")
	case kafka.PaymentStatusFailed:
		log.Info("payment failed, compensating")

		err := p.repo.UpdateStatus(ctx, bookingID, domain.StatusCancelled)
		if err != nil {
			log.Warn("booking cancellation skipped", "error", err)
		}

		booking, err := p.repo.GetByID(ctx, bookingID)
		if err != nil {
			return fmt.Errorf("failed to fetch booking data for compensation: %w", err)
		}

		if err := p.flightClient.ReleaseSeat(ctx, booking.FlightID, booking.SeatNumber); err != nil {
			log.Error("compensation failed", "error", err)
			return fmt.Errorf("compensation failed: %w", err)
		}
	}

	return nil
}
