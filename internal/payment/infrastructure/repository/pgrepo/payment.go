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

type createOrGetRow struct {
	ID            string     `db:"id"`
	BookingID     string     `db:"booking_id"`
	UserID        int64      `db:"user_id"`
	AmountCents   int64      `db:"amount_cents"`
	Currency      string     `db:"currency"`
	Status        string     `db:"status"`
	ProviderTxnID *string    `db:"provider_txn_id"`
	ErrorMessage  *string    `db:"error_message"`
	CreatedAt     time.Time  `db:"created_at"`
	ProcessedAt   *time.Time `db:"processed_at"`
	IsNew         bool       `db:"is_new"`
}

func (r *PaymentRepo) CreateOrGet(ctx context.Context, p *domain.Payment) (*domain.CreatePaymentResult, error) {
	if p.Status == "" {
		p.Status = domain.PaymentStatusPending
	}

	query := `
		WITH ins AS (
			INSERT INTO payments (booking_id, user_id, amount_cents, currency, status)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (booking_id) DO NOTHING
			RETURNING id, booking_id, user_id, amount_cents, currency, status,
			          provider_txn_id, error_message, created_at, processed_at
		)
		SELECT id, booking_id, user_id, amount_cents, currency, status,
		       provider_txn_id, error_message, created_at, processed_at,
		       true AS is_new
		FROM ins
		UNION ALL
		SELECT id, booking_id, user_id, amount_cents, currency, status,
		       provider_txn_id, error_message, created_at, processed_at,
		       false AS is_new
		FROM payments
		WHERE booking_id = $1 AND NOT EXISTS (SELECT 1 FROM ins)
	`

	var row createOrGetRow
	if err := r.db.GetContext(ctx, &row, query,
		p.BookingID,
		p.UserID,
		p.AmountCents,
		p.Currency,
		p.Status,
	); err != nil {
		return nil, fmt.Errorf("create or get payment: %w", err)
	}

	payment := &domain.Payment{
		ID:            row.ID,
		BookingID:     row.BookingID,
		UserID:        row.UserID,
		AmountCents:   row.AmountCents,
		Currency:      row.Currency,
		Status:        domain.PaymentStatus(row.Status),
		ProviderTxnID: row.ProviderTxnID,
		ErrorMessage:  row.ErrorMessage,
		CreatedAt:     row.CreatedAt,
		ProcessedAt:   row.ProcessedAt,
	}

	return &domain.CreatePaymentResult{
		Payment: payment,
		IsNew:   row.IsNew,
	}, nil
}

func (r *PaymentRepo) UpdateStatus(
	ctx context.Context,
	paymentID string,
	status domain.PaymentStatus,
	providerTxnID *string,
	errorMsg *string,
) error {
	query := `
		UPDATE payments
		SET status = $1,
		    provider_txn_id = COALESCE($2, provider_txn_id),
		    error_message = $3,
		    processed_at = COALESCE(processed_at, NOW())
		WHERE id = $4
	`

	res, err := r.db.ExecContext(ctx, query, status, providerTxnID, errorMsg, paymentID)
	if err != nil {
		return fmt.Errorf("failed to execute update: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrPaymentNotFound
	}

	return nil
}

func (r *PaymentRepo) GetByBookingID(ctx context.Context, bookingID string) (*domain.Payment, error) {
	query := `
		SELECT id, booking_id, user_id, amount_cents, currency, status,
		       provider_txn_id, error_message, created_at, processed_at
		FROM payments
		WHERE booking_id = $1
	`

	var p domain.Payment
	if err := r.db.GetContext(ctx, &p, query, bookingID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("get payment by booking id: %w", err)
	}

	return &p, nil
}
