package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
)

type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "PENDING"
	PaymentStatusSuccess PaymentStatus = "SUCCESS"
	PaymentStatusFailed  PaymentStatus = "FAILED"
)

type Payment struct {
	ID            string        `db:"id"`
	BookingID     string        `db:"booking_id"`
	UserID        int64         `db:"user_id"`
	AmountCents   int64         `db:"amount_cents"`
	Currency      string        `db:"currency"`
	Status        PaymentStatus `db:"status"`
	ProviderTxnID *string       `db:"provider_txn_id"`
	ErrorMessage  *string       `db:"error_message"`
	CreatedAt     time.Time     `db:"created_at"`
	ProcessedAt   *time.Time    `db:"processed_at"`
}

type CreatePaymentResult struct {
	Payment *Payment
	IsNew   bool
}

type PaymentRepository interface {
	CreateOrGet(ctx context.Context, payment *Payment) (*CreatePaymentResult, error)
	UpdateStatus(ctx context.Context, paymentID string, status PaymentStatus, providerTxnID *string, errorMessage *string) error
	GetByBookingID(ctx context.Context, bookingID string) (*Payment, error)
}
