package repository

import (
	"context"
	"encoding/json"
	"github.com/squ1ky/flyte/internal/booking/domain"
	"time"
)

const (
	EventTypePaymentRequest = "PAYMENT_REQUEST"
)

type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "PENDING"
	OutboxStatusProcessed OutboxStatus = "PROCESSED"
	OutboxStatusFailed    OutboxStatus = "FAILED"
)

type OutboxEvent struct {
	ID        string          `db:"id"`
	EventType string          `db:"event_type"`
	Payload   json.RawMessage `db:"payload"`
	Status    OutboxStatus    `db:"status"`
	CreatedAt time.Time       `db:"created_at"`
}

type BookingRepository interface {
	Create(ctx context.Context, booking *domain.Booking) (string, error)
	GetByID(ctx context.Context, id string) (*domain.Booking, error)
	UpdateStatus(ctx context.Context, id string, status domain.BookingStatus) error
	ListByUserID(ctx context.Context, userID int64) ([]domain.Booking, error)
	GetExpiredBookings(ctx context.Context, ttl time.Duration) ([]domain.Booking, error)

	GetPendingOutboxEvents(ctx context.Context, limit int) ([]OutboxEvent, error)
	MarkOutboxEventProcessed(ctx context.Context, id string) error
	MarkOutboxEventFailed(ctx context.Context, id string, reason string) error
}
