package repository

import (
	"context"
	"github.com/squ1ky/flyte/internal/payment/domain"
)

type PaymentRepository interface {
	CreateOrGet(ctx context.Context, payment *domain.Payment) (*domain.CreatePaymentResult, error)
	UpdateStatus(ctx context.Context, paymentID string, status domain.PaymentStatus, errorMessage *string) error
	GetByBookingID(ctx context.Context, bookingID string) (*domain.Payment, error)
}
