package outbox

import (
	"context"
	"encoding/json"
	"github.com/squ1ky/flyte/internal/flight/domain"
	"time"
)

type Status string

const (
	StatusPending    Status = "PENDING"
	StatusProcessing Status = "PROCESSING"
	StatusProcessed  Status = "PROCESSED"
	StatusFailed     Status = "FAILED"
)

type Event struct {
	ID         string           `db:"id"`
	EventType  domain.EventType `db:"event_type"`
	Payload    json.RawMessage  `db:"payload"`
	Status     Status           `db:"status"`
	RetryCount int              `db:"retry_count"`
	CreatedAt  time.Time        `db:"created_at"`
}

type Storage interface {
	FetchPending(ctx context.Context, limit int) ([]Event, error)
	MarkProcessed(ctx context.Context, ids []string) error
	MarkFailed(ctx context.Context, ids []string) error
}
