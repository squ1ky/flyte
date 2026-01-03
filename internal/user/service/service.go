package service

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/squ1ky/flyte/internal/user/domain"
)

type UserClaims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type Auth interface {
	Register(ctx context.Context, email, password, phone string) (int64, error)
	Login(ctx context.Context, email, password string) (string, error)
	ValidateToken(ctx context.Context, token string) (*UserClaims, error)
	GetUser(ctx context.Context, userID int64) (*domain.User, error)
}

type Passenger interface {
	AddPassenger(ctx context.Context, p *domain.Passenger) (int64, error)
	GetPassengers(ctx context.Context, userID int64) ([]domain.Passenger, error)
}
