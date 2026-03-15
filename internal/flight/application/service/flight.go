package service

import (
	"context"
	"fmt"
	"github.com/squ1ky/flyte/internal/flight/domain"
	"log/slog"
)

type FlightService struct {
	repo     domain.FlightStorage
	searcher domain.FlightSearcher
	logger   *slog.Logger
}

func NewFlightService(
	repo domain.FlightStorage,
	searcher domain.FlightSearcher,
	logger *slog.Logger,
) *FlightService {
	return &FlightService{
		repo:     repo,
		searcher: searcher,
		logger:   logger,
	}
}

func (s *FlightService) Search(ctx context.Context, criteria domain.FlightSearchCriteria) ([]domain.FlightDocument, error) {
	if err := criteria.Validate(); err != nil {
		return nil, err
	}

	results, err := s.searcher.Search(ctx, criteria)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to search flights",
			slog.String("origin", criteria.Origin),
			slog.String("destination", criteria.Destination),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("search flights: %w", err)
	}
	return results, nil
}

func (s *FlightService) Create(ctx context.Context, flight *domain.Flight, fares []domain.FlightFare) (int64, error) {
	if err := flight.Validate(); err != nil {
		return 0, err
	}

	id, err := s.repo.Create(ctx, flight, fares)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create flight",
			slog.String("flight_number", flight.FlightNumber),
			slog.Any("error", err),
		)
		return 0, fmt.Errorf("create flight: %w", err)
	}
	return id, nil
}

func (s *FlightService) UpdateStatus(ctx context.Context, flightID int64, status domain.FlightStatus) error {
	if flightID <= 0 {
		return domain.ErrInvalidFlightID
	}

	if err := s.repo.UpdateStatus(ctx, flightID, status); err != nil {
		s.logger.ErrorContext(ctx, "failed to update flight status",
			slog.Int64("flight_id", flightID),
			slog.String("status", string(status)),
			slog.Any("error", err),
		)
		return fmt.Errorf("update flight status: %w", err)
	}
	return nil
}

func (s *FlightService) GetByID(ctx context.Context, flightID int64) (*domain.Flight, error) {
	if flightID <= 0 {
		return nil, domain.ErrInvalidFlightID
	}

	flight, err := s.repo.GetByID(ctx, flightID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get flight",
			slog.Int64("flight_id", flightID),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("get flight: %w", err)
	}
	return flight, err
}

func (s *FlightService) GetFares(ctx context.Context, flightID int64) ([]domain.FlightFare, error) {
	if flightID <= 0 {
		return nil, domain.ErrInvalidFlightID
	}

	fares, err := s.repo.GetFaresByFlightID(ctx, flightID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get flight fares",
			slog.Int64("flight_id", flightID),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("get flight fares: %w", err)
	}
	return fares, err
}
