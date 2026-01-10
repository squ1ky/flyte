package pgrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/squ1ky/flyte/internal/payment/domain"
	"time"
)

type PaymentRepo struct {
	db *sqlx.DB
}

func NewPaymentRepo(db *sqlx.DB) *PaymentRepo {
	return &PaymentRepo{db: db}
}

func (r *PaymentRepo) CreateOrGet(ctx context.Context, p *domain.Payment) (*domain.CreatePaymentResult, error) {
	if p.Status == "" {
		p.Status = domain.PaymentStatusPending
	}
	now := time.Now()

	insertQuery := `
		INSERT INTO payments (booking_id, user_id, amount, currency, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (booking_id) DO NOTHING
		RETURNING id, created_at
	`

	var createdID string
	var createdAt time.Time

	err := r.db.QueryRowContext(ctx, insertQuery,
		p.BookingID,
		p.UserID,
		p.Amount,
		p.Currency,
		p.Status,
		now,
	).Scan(&createdID, &createdAt)

	if err == nil {
		p.ID = createdID
		p.CreatedAt = createdAt

		return &domain.CreatePaymentResult{
			Payment: p,
			IsNew:   true,
		}, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	existingPayment, err := r.GetByBookingID(ctx, p.BookingID)
	if err != nil {
		return nil, err
	}

	return &domain.CreatePaymentResult{
		Payment: existingPayment,
		IsNew:   false,
	}, nil
}

func (r *PaymentRepo) UpdateStatus(ctx context.Context, paymentID string, status domain.PaymentStatus, errorMsg *string) error {
	query := `
		UPDATE payments
		SET status = $1, error_message = $2, processed_at = NOW()
		WHERE id = $3
	`

	res, err := r.db.ExecContext(ctx, query, status, errorMsg, paymentID)
	if err != nil {
		return fmt.Errorf("failed to execute update: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("payment with id %s not found for update", paymentID)
	}

	return nil
}

func (r *PaymentRepo) GetByBookingID(ctx context.Context, bookingID string) (*domain.Payment, error) {
	query := `
		SELECT id, booking_id, user_id, amount, currency, status, error_message, created_at, processed_at
		FROM payments
		WHERE booking_id = $1
	`

	var p domain.Payment
	if err := r.db.GetContext(ctx, &p, query, bookingID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, err
	}

	return &p, nil
}
