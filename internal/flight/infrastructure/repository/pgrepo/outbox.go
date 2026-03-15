package pgrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/squ1ky/flyte/internal/flight/domain"
	"github.com/squ1ky/flyte/internal/flight/infrastructure/repository/outbox"
)

type OutboxRepo struct {
	db *sqlx.DB
}

func NewOutboxRepo(db *sqlx.DB) *OutboxRepo {
	return &OutboxRepo{db: db}
}

func (r *OutboxRepo) FetchPending(ctx context.Context, limit int) ([]outbox.Event, error) {
	query := `
		UPDATE flight_outbox
		SET status = :processing
		WHERE id IN (
		    SELECT id FROM flight_outbox
		    WHERE status = :pending
		    ORDER BY created_at ASC
		    LIMIT :limit
		    FOR UPDATE SKIP LOCKED
		)
		RETURNING id, event_type, payload, status, created_at
	`

	rows, err := r.db.NamedQueryContext(ctx, query, map[string]any{
		"processing": outbox.StatusProcessing,
		"pending":    outbox.StatusPending,
		"limit":      limit,
	})
	if err != nil {
		return nil, fmt.Errorf("fetch pending events: %w", err)
	}
	defer rows.Close()

	var events []outbox.Event
	for rows.Next() {
		var event outbox.Event
		if err := rows.StructScan(&event); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

func (r *OutboxRepo) MarkProcessed(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	query, args, err := sqlx.In(
		`UPDATE flight_outbox
		SET status = ? WHERE id IN (?)
		`,
		outbox.StatusProcessed,
		ids,
	)
	if err != nil {
		return fmt.Errorf("build mark processed query: %w", err)
	}

	if _, err := r.db.ExecContext(ctx, r.db.Rebind(query), args...); err != nil {
		return fmt.Errorf("mark processed: %w", err)
	}

	return nil
}

func (r *OutboxRepo) MarkFailed(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	query, args, err := sqlx.In(
		`UPDATE flight_outbox
    		SET status = ?
    		WHERE id IN (?)
    		`,
		outbox.StatusFailed,
		ids,
	)
	if err != nil {
		return fmt.Errorf("build mark failed query: %w", err)
	}

	if _, err := r.db.ExecContext(ctx, r.db.Rebind(query), args...); err != nil {
		return fmt.Errorf("mark failed: %w", err)
	}

	return nil
}

func insertOutboxEvent(ctx context.Context, tx *sqlx.Tx, eventType domain.EventType, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal outbox payload: %w", err)
	}

	query := `
       INSERT INTO flight_outbox (event_type, payload, status)
       VALUES ($1, $2, $3)
    `

	if _, err := tx.ExecContext(ctx, query, eventType, data, outbox.StatusPending); err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}
	return nil
}
