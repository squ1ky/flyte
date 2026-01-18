package service

import (
	"context"
	"errors"
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

	booking, err := p.repo.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, domain.ErrBookingNotFound) {
			log.Error("booking not found for paymentResult")
			return nil
		}
		return fmt.Errorf("failed to get booking: %w", err)
	}

	switch status {
	case kafka.PaymentStatusSuccess:
		if booking.Status == domain.StatusPaid {
			log.Info("booking already paid")
			return nil
		}

		err = p.flightClient.ConfirmSeat(ctx, booking.FlightID, booking.SeatNumber)
		if err != nil {
			log.Error("payment success but failed to confirm seat", "error", err)
			return fmt.Errorf("failed to confirm seat: %w", err)
		}

		err = p.repo.UpdateStatus(ctx, bookingID, domain.StatusPaid)
		if err != nil {
			log.Warn("failed to update local status to PAID", "error", err)
			return fmt.Errorf("failed to update status: %w", err)
		}
		log.Info("booking successfully confirmed and paid")

	case kafka.PaymentStatusFailed:
		log.Info("payment failed, cancelling booking")

		err := p.repo.UpdateStatus(ctx, bookingID, domain.StatusCancelled)
		if err != nil {
			log.Warn("booking cancellation skipped", "error", err)
		}

		err = p.flightClient.ReleaseSeat(ctx, booking.FlightID, booking.SeatNumber)
		if err != nil {
			log.Warn("failed to release seat", "error", err)
		}
	}

	return nil
}
