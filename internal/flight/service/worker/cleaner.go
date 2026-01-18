package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/squ1ky/flyte/internal/flight/domain"
	"log/slog"
	"time"
)

type SeatCleaner struct {
	db             *sqlx.DB
	logger         *slog.Logger
	interval       time.Duration
	reservationTTL time.Duration
}

func NewSeatCleaner(
	db *sqlx.DB,
	logger *slog.Logger,
	interval time.Duration,
	reservationTTL time.Duration,
) *SeatCleaner {
	return &SeatCleaner{
		db:             db,
		logger:         logger,
		interval:       interval,
		reservationTTL: reservationTTL,
	}
}

func (c *SeatCleaner) Start(ctx context.Context) {
	c.logger.Info("starting seat cleaner worker", "interval", c.interval, "ttl", c.reservationTTL)
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("stopping seat cleaner worker")
			return
		case <-ticker.C:
			if err := c.processExpiredReservations(ctx); err != nil {
				c.logger.Error("failed to process expired reservations", "error", err)
			}
		}
	}
}

func (c *SeatCleaner) processExpiredReservations(ctx context.Context) error {
	cutoffTime := time.Now().Add(-c.reservationTTL)

	tx, err := c.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	queryUpdate := `
		UPDATE seats
		SET is_booked = FALSE, reserved_at = NULL
		WHERE is_booked = TRUE
		  AND reserved_at IS NOT NULL
		  AND reserved_at < $1
		RETURNING flight_id
	`
	rows, err := c.db.QueryContext(ctx, queryUpdate, cutoffTime)
	if err != nil {
		return fmt.Errorf("update expired seats: %w", err)
	}

	affectedFlightIDs := make(map[int64]struct{})
	for rows.Next() {
		var flightID int64
		if err := rows.Scan(&flightID); err != nil {
			rows.Close()
			c.logger.Error("failed to scan flight id from cleaner", "error", err)
			return fmt.Errorf("scan flight id from cleaner: %w", err)
		}
		affectedFlightIDs[flightID] = struct{}{}
	}
	rows.Close()

	if len(affectedFlightIDs) == 0 {
		return nil
	}

	stmtOutbox, err := tx.PrepareContext(ctx, "INSERT INTO flight_outbox (event_type, payload) VALUES ($1, $2)")
	if err != nil {
		return fmt.Errorf("prepare outbox stmt: %w", err)
	}
	defer stmtOutbox.Close()

	for flightID := range affectedFlightIDs {
		payload := map[string]int64{
			"flight_id": flightID,
		}
		data, _ := json.Marshal(payload)

		if _, err := stmtOutbox.Exec(ctx, domain.EventSeatsChanged, data); err != nil {
			return fmt.Errorf("insert outbox event for flight %d: %w", flightID, err)
		}
	}

	return tx.Commit()
}
