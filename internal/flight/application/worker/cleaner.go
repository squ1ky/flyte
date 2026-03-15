package worker

import (
	"context"
	"github.com/squ1ky/flyte/internal/flight/domain"
	"log/slog"
	"time"
)

type SeatCleaner struct {
	reservationRepo domain.ReservationStorage
	logger          *slog.Logger
	interval        time.Duration
}

func NewSeatCleaner(
	reservationRepo domain.ReservationStorage,
	logger *slog.Logger,
	interval time.Duration,
) *SeatCleaner {
	return &SeatCleaner{
		reservationRepo: reservationRepo,
		logger:          logger,
		interval:        interval,
	}
}

func (c *SeatCleaner) Start(ctx context.Context) {
	c.logger.InfoContext(ctx, "starting seat cleaner worker",
		slog.String("interval", c.interval.String()),
	)
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.logger.InfoContext(ctx, "stopping seat cleaner worker")
			return
		case <-ticker.C:
			count, err := c.reservationRepo.CleanExpiredReservations(ctx)
			if err != nil {
				c.logger.ErrorContext(ctx, "failed to clean expired reservations",
					slog.Any("error", err),
				)
				continue
			}
			if count > 0 {
				c.logger.InfoContext(ctx, "cleaned expired reservations",
					slog.Int64("count", count),
				)
			}
		}
	}
}
