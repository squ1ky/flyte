package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"github.com/squ1ky/flyte/internal/flight/domain"
	"github.com/squ1ky/flyte/internal/flight/repository"
	"log/slog"
	"time"
)

type ElasticOutboxProcessor struct {
	db             *sqlx.DB
	flightSearcher repository.FlightSearcher
	logger         *slog.Logger
}

func NewElasticOutboxProcessor(
	db *sqlx.DB,
	searcher repository.FlightSearcher,
	logger *slog.Logger,
) *ElasticOutboxProcessor {
	return &ElasticOutboxProcessor{
		db:             db,
		flightSearcher: searcher,
		logger:         logger,
	}
}

func (p *ElasticOutboxProcessor) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	p.logger.Info("starting outbox processor")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.processBatch(ctx)
		}
	}
}

func (p *ElasticOutboxProcessor) processBatch(ctx context.Context) {
	query := `
		SELECT id, payload
		FROM flight_outbox
		WHERE status = 'PENDING'
		ORDER BY created_at DESC
		LIMIT 10
		FOR UPDATE SKIP LOCKED
	`

	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		p.logger.Error("failed to begin tx", "error", err)
		return
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		p.logger.Warn("failed to fetch outbox events", "error", err)
		return
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			p.logger.Error("failed to close rows", "error", err)
		}
	}(rows)

	for rows.Next() {
		var id string
		var payload []byte
		if err := rows.Scan(&id, &payload); err != nil {
			continue
		}

		var flight domain.Flight
		if err := json.Unmarshal(payload, &flight); err != nil {
			p.logger.Error("failed to unmarshal flight", "id", id, "error", err)

			_, _ = tx.ExecContext(ctx, "UPDATE flight_outbox SET STATUS = 'FAILED' WHERE id = $1", id)
			continue
		}

		if err := p.flightSearcher.IndexFlight(ctx, &flight); err != nil {
			p.logger.Error("failed to index flight", "id", id, "error", err)
			return
		}

		if _, err := tx.ExecContext(ctx, "DELETE FROM flight_outbox WHERE id = $1", id); err != nil {
			p.logger.Error("failed to delete processed event", "id", id, "error", err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		p.logger.Error("failed to commit outbox processing", "error", err)
	}
}
