package repository

import (
	"context"
	"github.com/squ1ky/flyte/internal/booking/domain"
)

type BookingRepository interface {
	Create(ctx context.Context, booking *domain.Booking) (string, error)
	GetByID(ctx context.Context, id string) (*domain.Booking, error)
	UpdateStatus(ctx context.Context, id string, status domain.BookingStatus) error
	ListByUserID(ctx context.Context, userID int64) ([]domain.Booking, error)
}
