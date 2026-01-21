package pgrepo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/squ1ky/flyte/internal/booking/domain"
	"github.com/squ1ky/flyte/internal/booking/domain/events"
	"github.com/squ1ky/flyte/internal/booking/repository"
	"time"
)

type BookingRepo struct {
	db *sqlx.DB
}

func NewBookingRepo(db *sqlx.DB) *BookingRepo {
	return &BookingRepo{db: db}
}

func (r *BookingRepo) Create(ctx context.Context, b *domain.Booking) (string, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	queryBooking := `
		INSERT INTO bookings (
			user_id, flight_id, seat_number,
		    passenger_name, passenger_passport,
		    price_cents, currency, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id
	`

	var id string
	err = tx.QueryRowContext(ctx, queryBooking,
		b.UserID, b.FlightID, b.SeatNumber,
		b.PassengerName, b.PassengerPassport,
		b.PriceCents, b.Currency, b.Status,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to create booking: %w", err)
	}

	payload := events.PaymentRequestEvent{
		BookingID:   id,
		UserID:      b.UserID,
		AmountCents: b.PriceCents,
		Currency:    b.Currency,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal outbox payload: %w", err)
	}

	queryOutbox := `
		INSERT INTO booking_outbox (event_type, payload, status)
		VALUES ($1, $2, $3)
	`
	if _, err := tx.ExecContext(ctx, queryOutbox, repository.EventTypePaymentRequest, payloadBytes, repository.OutboxStatusPending); err != nil {
		return "", fmt.Errorf("failed to insert outbox event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit tx: %w", err)
	}

	return id, nil
}

func (r *BookingRepo) GetByID(ctx context.Context, id string) (*domain.Booking, error) {
	var booking domain.Booking
	query := `SELECT * FROM bookings WHERE id = $1`

	if err := r.db.GetContext(ctx, &booking, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("booking %s: %w", id, domain.ErrBookingNotFound)
		}
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	return &booking, nil
}

func (r *BookingRepo) UpdateStatus(ctx context.Context, id string, status domain.BookingStatus) error {
	query := `
		UPDATE bookings
		SET status = $1, updated_at = NOW()
		WHERE id = $2 AND status = 'PENDING'
	`

	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to execute update status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("booking %s not found or already processed: %w", id, domain.ErrBookingNotFound)
	}

	return nil
}

func (r *BookingRepo) ListByUserID(ctx context.Context, userID int64) ([]domain.Booking, error) {
	var bookings []domain.Booking
	query := `SELECT * FROM bookings WHERE user_id = $1 ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &bookings, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookings: %w", err)
	}

	if bookings == nil {
		bookings = []domain.Booking{}
	}

	return bookings, nil
}

func (r *BookingRepo) GetExpiredBookings(ctx context.Context, ttl time.Duration) ([]domain.Booking, error) {
	var bookings []domain.Booking

	cutoffTime := time.Now().Add(-ttl)
	query := `
		SELECT * FROM bookings
		WHERE status = 'PENDING'
		AND created_at < $1
	`
	if err := r.db.SelectContext(ctx, &bookings, query, cutoffTime); err != nil {
		return nil, fmt.Errorf("failed to get expired bookings: %w", err)
	}

	if bookings == nil {
		bookings = []domain.Booking{}
	}

	return bookings, nil
}

func (r *BookingRepo) GetPendingOutboxEvents(ctx context.Context, limit int) ([]repository.OutboxEvent, error) {
	var outboxEvents []repository.OutboxEvent

	query := `
		SELECT id, event_type, payload, status
		FROM booking_outbox
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`

	if err := r.db.SelectContext(ctx, &outboxEvents, query, repository.OutboxStatusPending, limit); err != nil {
		return nil, fmt.Errorf("failed to fetch pending outbox events: %w", err)
	}

	if outboxEvents == nil {
		outboxEvents = []repository.OutboxEvent{}
	}
	return outboxEvents, nil
}

func (r *BookingRepo) MarkOutboxEventProcessed(ctx context.Context, id string) error {
	query := `
		UPDATE booking_outbox
 		SET status = $1, processed_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, repository.OutboxStatusProcessed, id)
	if err != nil {
		return fmt.Errorf("failed to mark outbox event as processeed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrBookingNotFound
	}

	return nil
}

func (r *BookingRepo) MarkOutboxEventFailed(ctx context.Context, id string, reason string) error {
	query := `
		UPDATE booking_outbox
		SET status = $1, processed_at = NOW(), error_message = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, repository.OutboxStatusFailed, reason, id)
	if err != nil {
		return fmt.Errorf("failed to mark outbox event as failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrBookingNotFound
	}

	return nil
}
