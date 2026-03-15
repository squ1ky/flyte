package domain

import (
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) (int64, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
}

type PassengerRepository interface {
	Create(ctx context.Context, passenger *Passenger) (int64, error)
	GetByUserID(ctx context.Context, userID int64) ([]Passenger, error)
	Delete(ctx context.Context, userID, passengerID int64) error
}
