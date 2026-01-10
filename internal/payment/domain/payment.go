package domain

import (
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
	ID           string        `db:"id"`
	BookingID    string        `db:"booking_id"`
	UserID       string        `db:"user_id"`
	Amount       float64       `db:"amount"`
	Currency     string        `db:"currency"`
	Status       PaymentStatus `db:"status"`
	ErrorMessage *string       `db:"error_message"`
	CreatedAt    time.Time     `db:"created_at"`
	ProcessedAt  *time.Time    `db:"processed_at"`
}

type CreatePaymentResult struct {
	Payment *Payment
	IsNew   bool
}
