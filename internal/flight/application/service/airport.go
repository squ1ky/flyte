package service

import (
	"context"
	"fmt"
	"github.com/squ1ky/flyte/internal/flight/domain"
	"log/slog"
)

type AirportService struct {
	repo   domain.AirportStorage
	logger *slog.Logger
}

func NewAirportService(
	repo domain.AirportStorage,
	logger *slog.Logger,
) *AirportService {
	return &AirportService{
		repo:   repo,
		logger: logger,
	}
}

func (s *AirportService) GetByCode(ctx context.Context, code string) (*domain.Airport, error) {
	if code == "" {
		return nil, domain.ErrInvalidAirportCode
	}

	airport, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get airport",
			slog.String("code", code),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("get airport: %w", err)
	}
	return airport, nil
}

func (s *AirportService) Search(ctx context.Context, query string) ([]domain.Airport, error) {
	if query == "" {
		return nil, domain.ErrEmptySearchQuery
	}

	airports, err := s.repo.SearchAirports(ctx, query)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to search airports",
			slog.String("query", query),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("search airports: %w", err)
	}
	return airports, nil
}
