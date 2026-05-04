package service

import (
	"context"
	"fmt"
	"github.com/squ1ky/flyte/internal/flight/domain"
	"log/slog"
)

type AircraftService struct {
	repo   domain.AircraftStorage
	logger *slog.Logger
}

func NewAircraftService(
	repo domain.AircraftStorage,
	logger *slog.Logger,
) *AircraftService {
	return &AircraftService{
		repo:   repo,
		logger: logger,
	}
}

func (s *AircraftService) Create(ctx context.Context, aircraft *domain.Aircraft) (int64, error) {
	if err := aircraft.Validate(); err != nil {
		return 0, err
	}

	id, err := s.repo.Create(ctx, aircraft)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create aircraft",
			slog.String("model", aircraft.Model),
			slog.Any("error", err),
		)
		return 0, fmt.Errorf("create aircraft: %w", err)
	}

	return id, nil
}

func (s *AircraftService) GetByID(ctx context.Context, aircraftID int64) (*domain.Aircraft, error) {
	aircraft, err := s.repo.GetByID(ctx, aircraftID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get aircraft",
			slog.Int64("aircraft_id", aircraftID),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("get aircraft: %w", err)
	}
	return aircraft, nil
}

func (s *AircraftService) AddSeats(ctx context.Context, aircraftID int64, seats []domain.AircraftSeat) error {
	if len(seats) == 0 {
		return domain.ErrEmptySeatsList
	}

	if _, err := s.repo.GetByID(ctx, aircraftID); err != nil {
		s.logger.ErrorContext(ctx, "failed to get aircraft",
			slog.Int64("aircraft_id", aircraftID),
			slog.Any("error", err),
		)
		return fmt.Errorf("aircraft lookup: %w", err)
	}

	if err := s.repo.AddSeats(ctx, aircraftID, seats); err != nil {
		s.logger.ErrorContext(ctx, "failed to add aircraft seats",
			slog.Int64("aircraft_id", aircraftID),
			slog.Any("error", err),
		)
		return fmt.Errorf("aircraft add seats: %w", err)
	}
	return nil
}

func (s *AircraftService) GetSeats(ctx context.Context, aircraftID int64) ([]domain.AircraftSeat, error) {
	seats, err := s.repo.GetSeatsByAircraftID(ctx, aircraftID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get aircraft seats",
			slog.Int64("aircraft_id", aircraftID),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("get aircraft seats: %w", err)
	}
	return seats, nil
}

func (s *AircraftService) List(ctx context.Context, limit, offset int) ([]domain.Aircraft, error) {
	aircrafts, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to list aircrafts", slog.Any("error", err))
		return nil, fmt.Errorf("list aircrafts: %w", err)
	}
	return aircrafts, nil
}
