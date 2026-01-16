package service

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/squ1ky/flyte/internal/flight/repository"
	"log/slog"
	"time"
)

const (
	CleanupInterval = 1 * time.Minute
	ReservationTTL  = 15 * time.Minute
)

type SeatCleaner struct {
	db             *sqlx.DB
	flightRepo     repository.FlightStorage
	flightSearcher repository.FlightSearcher
	logger         *slog.Logger
}

func NewSeatCleaner(
	db *sqlx.DB,
	flightRepo repository.FlightStorage,
	flightSearcher repository.FlightSearcher,
	logger *slog.Logger,
) *SeatCleaner {
	return &SeatCleaner{
		db:             db,
		flightRepo:     flightRepo,
		flightSearcher: flightSearcher,
		logger:         logger,
	}
}

func (c *SeatCleaner) Start(ctx context.Context) {
	c.logger.Info("starting seat cleaner worker", "interval", CleanupInterval, "ttl", ReservationTTL)
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("stopping seat cleaner worker")
			return
		case <-ticker.C:
			c.processExpiredReservations(ctx)
		}
	}
}

func (c *SeatCleaner) processExpiredReservations(ctx context.Context) {
	cutoffTime := time.Now().Add(-ReservationTTL)

	query := `
		UPDATE seats
		SET is_booked = FALSE, reserved_at = NULL
		WHERE is_booked = TRUE
		  AND reserved_at IS NOT NULL
		  AND reserved_at < $1
		RETURNING flight_id
	`

	rows, err := c.db.QueryContext(ctx, query, cutoffTime)
	if err != nil {
		c.logger.Error("failed to clean expired reservations", "error", err)
		return
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			c.logger.Error("failed to close rows", "error", err)
		}
	}(rows)

	affectedFlightIDs := make(map[int64]struct{})
	for rows.Next() {
		var flightID int64
		if err := rows.Scan(&flightID); err != nil {
			c.logger.Error("failed to scan flight id from cleaner", "error", err)
			continue
		}
		affectedFlightIDs[flightID] = struct{}{}
	}

	if len(affectedFlightIDs) > 0 {
		c.logger.Info("cleaned expired seats", "affected_flights_count", len(affectedFlightIDs))
	}

	for flightID := range affectedFlightIDs {
		c.syncElastic(ctx, flightID)
	}
}

func (c *SeatCleaner) syncElastic(ctx context.Context, flightID int64) {
	flight, err := c.flightRepo.GetByID(ctx, flightID)
	if err != nil {
		c.logger.Error("cleaner: failed to get flight for sync", "flight_id", flightID, "error", err)
		return
	}

	err = c.flightSearcher.UpdateAvailableSeats(ctx, flightID, flight.AvailableSeats)
	if err != nil {
		c.logger.Error("cleaner: failed to update elastic", "flight_id", flightID, "error", err)
	} else {
		c.logger.Debug("cleaner: synced flight with elastic", "flight_id", flightID, "seats", flight.AvailableSeats)
	}
}
