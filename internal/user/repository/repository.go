package repository

import (
	"context"
	"github.com/squ1ky/flyte/internal/user/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (int64, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
}

type PassengerRepository interface {
	Create(ctx context.Context, passenger *domain.Passenger) (int64, error)
	GetByUserID(ctx context.Context, userID int64) ([]domain.Passenger, error)
}
