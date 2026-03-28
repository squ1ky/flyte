package pgrepo

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/squ1ky/flyte/internal/booking/infrastructure/outbox"
)

type OutboxRepo struct {
	db *sqlx.DB
}

func NewOutboxRepo(db *sqlx.DB) *OutboxRepo {
	return &OutboxRepo{db: db}
}

func (r *OutboxRepo) Insert(ctx context.Context, event outbox.Event) error {
	const query = `
		INSERT INTO booking_outbox (id, event_type, payload, status, created_at)
		VALUES (:id, :event_type, :payload, :status, :created_at)
	`

	if _, err := r.db.NamedQueryContext(ctx, query, &event); err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}
	return nil
}

func (r *OutboxRepo) FetchPending(ctx context.Context, limit int) ([]outbox.Event, error) {
	var events []outbox.Event
	const query = `
		SELECT id, event_type, payload, status, created_at
		FROM booking_outbox
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT $1
	`

	if err := r.db.SelectContext(ctx, &events, query, limit); err != nil {
		return nil, fmt.Errorf("fetch pending outbox events: %w", err)
	}
	return events, nil
}

func (r *OutboxRepo) MarkSent(ctx context.Context, id string) error {
	const query = `
		UPDATE booking_outbox
		SET status = 'SENT'
		WHERE id = $1
	`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("mark outbox event sent: %w", err)
	}

	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("outbox event %s not found", id)
	}
	return nil
}
