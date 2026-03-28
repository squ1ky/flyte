package outbox

import (
	"context"
	"time"
)

type EventType string

const (
	EventPaymentRequest EventType = "payment_request"
)

type Status string

const (
	StatusPending Status = "PENDING"
	StatusSent    Status = "SENT"
)

type Event struct {
	ID        string    `db:"id"`
	EventType EventType `db:"event_type"`
	Payload   []byte    `db:"payload"`
	Status    Status    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
}

type Repository interface {
	Insert(ctx context.Context, event Event) error
	FetchPending(ctx context.Context, limit int) ([]Event, error)
	MarkSent(ctx context.Context, id string) error
}
