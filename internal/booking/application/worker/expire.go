package worker

import (
	"context"
	"github.com/squ1ky/flyte/internal/booking/config"
	"github.com/squ1ky/flyte/internal/booking/domain"
	flightclient "github.com/squ1ky/flyte/internal/booking/infrastructure/clients/grpc/flight"
	"log/slog"
	"time"
)

const expireBatchSize = 50

type ExpiredBookingCleaner struct {
	bookings domain.BookingRepository
	flights  *flightclient.Client
	interval time.Duration
	logger   *slog.Logger
}

func NewExpiredBookingCleaner(
	bookings domain.BookingRepository,
	flights *flightclient.Client,
	cfg config.CleanerConfig,
	log *slog.Logger,
) *ExpiredBookingCleaner {
	return &ExpiredBookingCleaner{
		bookings: bookings,
		flights:  flights,
		interval: cfg.Interval,
		logger:   log,
	}
}

func (c *ExpiredBookingCleaner) Run(ctx context.Context) error {
	c.logger.InfoContext(ctx, "expire cleaner worker started",
		slog.String("interval", c.interval.String()),
	)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.logger.InfoContext(ctx, "expire cleaner worker stopping")
			return ctx.Err()
		case <-ticker.C:
			c.expireBatch(ctx)
		}
	}
}

func (c *ExpiredBookingCleaner) expireBatch(ctx context.Context) {
	expired, err := c.bookings.FindExpired(ctx, expireBatchSize)
	if err != nil {
		c.logger.ErrorContext(ctx, "find expired bookings failed",
			slog.Any("error", err),
		)
		return
	}

	if len(expired) == 0 {
		return
	}

	c.logger.InfoContext(ctx, "expiring bookings",
		slog.Int("count", len(expired)),
	)

	cancelled := 0
	for _, booking := range expired {
		log := c.logger.With("booking_id", booking.ID)

		if err := c.bookings.UpdateStatus(ctx, booking.ID, domain.BookingStatusCancelled); err != nil {
			log.ErrorContext(ctx, "expire booking failed",
				slog.Any("error", err),
			)
			continue
		}

		if err := c.flights.CancelReservation(ctx, booking.ID); err != nil {
			log.ErrorContext(ctx, "compensation failed: cancel reservation",
				slog.Any("error", err),
			)
		}

		cancelled++
	}

	c.logger.InfoContext(ctx, "expire batch done",
		slog.Int("cancelled", cancelled),
	)
}
