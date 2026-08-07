package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/squ1ky/flyte/internal/booking/config"
	"github.com/squ1ky/flyte/internal/booking/domain"
	"github.com/squ1ky/flyte/internal/booking/infrastructure/outbox"
)

const expireBatchSize = 50

type ExpiredBookingCleaner struct {
	bookings domain.BookingRepository
	flights  FlightClient
	outbox   outbox.Repository
	interval time.Duration
	logger   *slog.Logger
}

func NewExpiredBookingCleaner(
	bookings domain.BookingRepository,
	flights FlightClient,
	outbox outbox.Repository,
	cfg config.CleanerConfig,
	log *slog.Logger,
) *ExpiredBookingCleaner {
	return &ExpiredBookingCleaner{
		bookings: bookings,
		flights:  flights,
		outbox:   outbox,
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

		reason := domain.CancelReasonExpired
		if err := c.bookings.UpdateStatus(ctx, booking.ID, domain.BookingStatusCancelled, &reason); err != nil {
			log.ErrorContext(ctx, "expire booking failed",
				slog.Any("error", err),
			)
			continue
		}

		if err := c.publishBookingCancelled(ctx, booking.ID, reason); err != nil {
			log.ErrorContext(ctx, "publish booking cancelled event failed",
				slog.Any("error", err),
			)
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

func (c *ExpiredBookingCleaner) publishBookingCancelled(ctx context.Context, bookingID string, reason domain.CancelReason) error {
	event := domain.BookingCancelledEvent{
		BookingID:   bookingID,
		Reason:      reason,
		CancelledAt: time.Now(),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal booking cancelled event: %w", err)
	}

	return c.outbox.Insert(ctx, outbox.Event{
		ID:        uuid.New().String(),
		EventType: outbox.EventBookingCancelled,
		Payload:   payload,
		Status:    outbox.StatusPending,
		CreatedAt: time.Now(),
	})
}
