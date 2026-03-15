package service

import (
	"context"
	"fmt"
	"github.com/squ1ky/flyte/internal/flight/domain"
	"log/slog"
)

type AirlineService struct {
	repo   domain.AirlineStorage
	logger *slog.Logger
}

func NewAirlineService(
	repo domain.AirlineStorage,
	logger *slog.Logger,
) *AirlineService {
	return &AirlineService{
		repo:   repo,
		logger: logger,
	}
}

func (s *AirlineService) Create(ctx context.Context, airline *domain.Airline) (int64, error) {
	if err := airline.Validate(); err != nil {
		return 0, err
	}

	id, err := s.repo.Create(ctx, airline)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create airline",
			slog.String("iata_code", airline.IATACode),
			slog.Any("error", err),
		)
		return 0, fmt.Errorf("create airline: %w", err)
	}
	return id, nil
}

func (s *AirlineService) GetByID(ctx context.Context, airlineID int64) (*domain.Airline, error) {
	if airlineID == 0 {
		return nil, domain.ErrInvalidAirlineID
	}

	airline, err := s.repo.GetByID(ctx, airlineID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get airline",
			slog.Int64("airline_id", airlineID),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("get airline: %w", err)
	}
	return airline, nil
}

func (s *AirlineService) List(ctx context.Context, limit, offset int) ([]domain.Airline, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	airlines, err := s.repo.ListAirlines(ctx, limit, offset)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to list airlines",
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("list airlines: %w", err)
	}
	return airlines, nil
}
