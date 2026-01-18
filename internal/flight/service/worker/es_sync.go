package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/squ1ky/flyte/internal/flight/domain"
	"github.com/squ1ky/flyte/internal/flight/repository"
	"log/slog"
	"time"
)

type OutboxEventStatus string

const (
	OutboxStatusPending = "PENDING"
	OutboxStatusFailed  = "FAILED"
)

type ElasticSyncWorker struct {
	db             *sqlx.DB
	flightRepo     repository.FlightStorage
	flightSearcher repository.FlightSearcher
	logger         *slog.Logger
}

func NewElasticSyncWorker(
	db *sqlx.DB,
	flightRepo repository.FlightStorage,
	searcher repository.FlightSearcher,
	logger *slog.Logger,
) *ElasticSyncWorker {
	return &ElasticSyncWorker{
		db:             db,
		flightRepo:     flightRepo,
		flightSearcher: searcher,
		logger:         logger,
	}
}

func (w *ElasticSyncWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	w.logger.Info("starting outbox processor")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("stopping elastic sync worker")
			return
		case <-ticker.C:
			if err := w.processBatch(ctx); err != nil {
				w.logger.Error("batch processing error", "error", err)
			}
		}
	}
}

func (w *ElasticSyncWorker) processBatch(ctx context.Context) error {
	query := `
		SELECT id, event_type, payload
		FROM flight_outbox
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT 50
		FOR UPDATE SKIP LOCKED
	`

	tx, err := w.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("fetch events: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			w.logger.Error("failed to close rows", "error", err)
		}
	}(rows)

	type OutboxEvent struct {
		ID        string           `db:"id"`
		EventType domain.EventType `db:"event_type"`
		Payload   json.RawMessage  `db:"payload"`
	}
	var events []OutboxEvent

	for rows.Next() {
		var evt OutboxEvent
		if err := rows.Scan(&evt.ID, &evt.EventType, &evt.Payload); err != nil {
			w.logger.Error("scan event failed", "error", err)
			continue
		}
		events = append(events, evt)
	}
	rows.Close()

	if len(events) == 0 {
		return nil
	}

	stmtFailed, err := tx.PrepareContext(ctx, "UPDATE flight_outbox SET status = $1 WHERE id = $2")
	if err != nil {
		return err
	}
	defer stmtFailed.Close()
	stmtDelete, err := tx.PrepareContext(ctx, "DELETE FROM flight_outbox WHERE id = $1")
	if err != nil {
		return err
	}
	defer stmtDelete.Close()

	for _, evt := range events {
		if err := w.handleEvent(ctx, evt.EventType, evt.Payload); err != nil {
			w.logger.Error("handle event failed",
				"id", evt.ID,
				"type", evt.EventType,
				"error", err)
			if _, execErr := stmtFailed.ExecContext(ctx, OutboxStatusFailed, evt.ID); execErr != nil {
				w.logger.Error("failed to mark event as failed", "id", evt.ID, "error", execErr)
			}
			continue
		}

		if _, execErr := stmtDelete.ExecContext(ctx, evt.ID); execErr != nil {
			return fmt.Errorf("delete event %s: %w", evt.ID, err)
		}
	}

	return tx.Commit()
}

func (w *ElasticSyncWorker) handleEvent(ctx context.Context, eventType domain.EventType, payload []byte) error {
	switch eventType {
	case domain.EventFlightCreated:
		var flight domain.Flight
		if err := json.Unmarshal(payload, &flight); err != nil {
			return fmt.Errorf("unmarshal flight: %w", err)
		}
		return w.flightSearcher.IndexFlight(ctx, &flight)
	case domain.EventSeatsChanged:
		var eventData struct {
			FlightID int64 `json:"flight_id"`
		}
		if err := json.Unmarshal(payload, &eventData); err != nil {
			return fmt.Errorf("unmarshal seats event: %w", err)
		}

		flight, err := w.flightRepo.GetByID(ctx, eventData.FlightID)
		if err != nil {
			return fmt.Errorf("get fresh flight data: %w", err)
		}
		return w.flightSearcher.UpdateAvailableSeats(ctx, flight.ID, flight.AvailableSeats)
	default:
		return fmt.Errorf("unknown event type: %s", eventType)
	}
}
