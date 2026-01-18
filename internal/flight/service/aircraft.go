package service

import (
	"context"
	"fmt"
	"github.com/squ1ky/flyte/internal/flight/domain"
	"github.com/squ1ky/flyte/internal/flight/repository"
	"log/slog"
)

type AircraftService struct {
	repo   repository.AircraftStorage
	logger *slog.Logger
}

func NewAircraftService(
	repo repository.AircraftStorage,
	logger *slog.Logger,
) *AircraftService {
	return &AircraftService{
		repo:   repo,
		logger: logger,
	}
}

func (s *AircraftService) CreateAircraft(ctx context.Context, model string, totalSeats int) (int64, error) {
	if totalSeats <= 0 {
		return 0, fmt.Errorf("total seats must be positive")
	}

	id, err := s.repo.CreateAircraft(ctx, model, totalSeats)
	if err != nil {
		s.logger.Error("failed to create aircraft", "model", model, "error", err)
		return 0, fmt.Errorf("create aircraft: %w", err)
	}

	return id, nil
}

func (s *AircraftService) ConfigureSeats(ctx context.Context, aircraftID int64, seats []domain.AircraftSeat) error {
	if len(seats) == 0 {
		return fmt.Errorf("seats list is empty")
	}

	if err := s.repo.AddAircraftSeats(ctx, aircraftID, seats); err != nil {
		s.logger.Error("failed to configure aircraft", "aircraft_id", aircraftID, "error", err)
		return fmt.Errorf("configure seats: %w", err)
	}

	return nil
}

func (s *AircraftService) GetAircrafts(ctx context.Context) ([]domain.Aircraft, error) {
	list, err := s.repo.GetAircrafts(ctx)
	if err != nil {
		s.logger.Error("failed to get aircraft list", "error", err)
		return nil, fmt.Errorf("get aircrafts: %w", err)
	}
	return list, nil
}

func (s *AircraftService) GetAircraftDetails(ctx context.Context, aircraftID int64) (*domain.Aircraft, []domain.AircraftSeat, error) {
	aircraft, err := s.repo.GetAircraftByID(ctx, aircraftID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get aircraft: %w", err)
	}

	seats, err := s.repo.GetAircraftSeats(ctx, aircraftID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get aircraft seats schema: %w", err)
	}

	return aircraft, seats, nil
}
