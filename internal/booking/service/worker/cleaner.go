package worker

import (
	"context"
	"fmt"
	"github.com/squ1ky/flyte/internal/booking/clients/grpc/flight"
	"github.com/squ1ky/flyte/internal/booking/domain"
	"github.com/squ1ky/flyte/internal/booking/repository"
	"log/slog"
	"time"
)

type ExpiredBookingCleaner struct {
	repo         repository.BookingRepository
	flightClient *flight.Client
	log          *slog.Logger
	interval     time.Duration
	bookingTTL   time.Duration
}

func NewExpiredBookingCleaner(
	repo repository.BookingRepository,
	flightClient *flight.Client,
	log *slog.Logger,
	interval time.Duration,
	bookingTTL time.Duration,
) *ExpiredBookingCleaner {
	return &ExpiredBookingCleaner{
		repo:         repo,
		flightClient: flightClient,
		log:          log,
		interval:     interval,
		bookingTTL:   bookingTTL,
	}
}

func (c *ExpiredBookingCleaner) Start(ctx context.Context) {
	c.log.Info("starting expired booking cleaner worker",
		"interval", c.interval,
		"bookingTTL", c.bookingTTL)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.log.Info("stopping expired booking cleaner worker")
			return
		case <-ticker.C:
			if err := c.process(ctx); err != nil {
				c.log.Error("failed to process expired bookings", "error", err)
			}
		}
	}
}

func (c *ExpiredBookingCleaner) process(ctx context.Context) error {
	bookings, err := c.repo.GetExpiredBookings(ctx, c.bookingTTL)
	if err != nil {
		return fmt.Errorf("fetch expired: %w", err)
	}

	if len(bookings) == 0 {
		return nil
	}

	c.log.Info("found expired bookings", "count", len(bookings))

	for _, b := range bookings {
		log := c.log.With("booking_id", b.ID)
		if err := c.repo.UpdateStatus(ctx, b.ID, domain.StatusTimeout); err != nil {
			log.Error("failed to update booking status to TIMEOUT", "error", err)
			continue
		}

		if err := c.flightClient.ReleaseSeat(ctx, b.FlightID, b.SeatNumber); err != nil {
			log.Warn("failed to release seat in flight service", "error", err)
		} else {
			log.Info("booking expired and seat released successfully")
		}
	}

	return nil
}
