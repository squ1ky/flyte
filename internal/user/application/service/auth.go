package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/squ1ky/flyte/internal/user/application/formatter"
	"github.com/squ1ky/flyte/internal/user/config"
	"github.com/squ1ky/flyte/internal/user/domain"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type UserClaims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	userRepo domain.UserRepository
	cfg      config.JWTConfig
	logger   *slog.Logger
}

func NewAuthService(
	repo domain.UserRepository,
	cfg config.JWTConfig,
	logger *slog.Logger,
) *AuthService {
	return &AuthService{
		userRepo: repo,
		cfg:      cfg,
		logger:   logger,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, phone string) (int64, error) {
	if err := domain.ValidateRegister(email, password, phone); err != nil {
		return 0, err
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to generate password hash",
			slog.Any("error", err),
		)
		return 0, fmt.Errorf("failed to generate password hash: %w", err)
	}

	cleanPhone := formatter.FormatPhoneNumber(phone)

	user := &domain.User{
		Email:        email,
		PasswordHash: string(passHash),
		PhoneNumber:  cleanPhone,
		Role:         domain.RoleUser,
	}

	id, err := s.userRepo.Create(ctx, user)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			return 0, domain.ErrUserAlreadyExists
		}
		s.logger.ErrorContext(ctx, "failed to create user",
			slog.String("email", email),
			slog.Any("error", err),
		)
		return 0, fmt.Errorf("failed to register user: %w", err)
	}

	s.logger.InfoContext(ctx, "user registered",
		slog.Int64("user_id", id),
		slog.String("email", email),
	)

	return id, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (userID int64, token string, err error) {
	if err := domain.ValidateLogin(email, password); err != nil {
		return 0, "", err
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return 0, "", domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return 0, "", domain.ErrInvalidCredentials
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &UserClaims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.TTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})

	tokenString, err := jwtToken.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to sign token",
			slog.Int64("user_id", user.ID),
			slog.Any("error", err),
		)
		return 0, "", fmt.Errorf("failed to sign token: %w", err)
	}

	s.logger.InfoContext(ctx, "user logged in",
		slog.Int64("user_id", user.ID),
	)
	
	return user.ID, tokenString, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (*UserClaims, error) {
	if token == "" {
		return nil, domain.NewValidationError("token", "is required")
	}

	jwtToken, err := jwt.ParseWithClaims(token, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := jwtToken.Claims.(*UserClaims)
	if !ok || !jwtToken.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (s *AuthService) GetUser(ctx context.Context, userID int64) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}
