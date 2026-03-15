package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/squ1ky/flyte/internal/user/application/formatter"
	"github.com/squ1ky/flyte/internal/user/domain"
	"log/slog"
)

type PassengerService struct {
	passRepo domain.PassengerRepository
	logger   *slog.Logger
}

func NewPassengerService(repo domain.PassengerRepository, logger *slog.Logger) *PassengerService {
	return &PassengerService{
		passRepo: repo,
		logger:   logger,
	}
}

func (s *PassengerService) AddPassenger(ctx context.Context, p *domain.Passenger) (int64, error) {
	if err := domain.ValidatePassenger(p); err != nil {
		return 0, err
	}

	p.DocumentNumber = formatter.FormatDocumentNumber(p.DocumentNumber)
	p.DocumentType = formatter.FormatDocumentType(p.DocumentType)
	p.FirstName = formatter.FormatName(p.FirstName)
	p.LastName = formatter.FormatName(p.LastName)
	p.MiddleName = formatter.FormatName(p.MiddleName)

	id, err := s.passRepo.Create(ctx, p)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create passenger",
			slog.Int64("user_id", p.UserID),
			slog.Any("error", err),
		)
		return 0, fmt.Errorf("failed to add passenger: %w", err)
	}

	s.logger.InfoContext(ctx, "passenger added",
		slog.Int64("passenger_id", id),
		slog.Int64("user_id", p.UserID),
	)

	return id, nil
}

func (s *PassengerService) GetPassengers(ctx context.Context, userID int64) ([]domain.Passenger, error) {
	passengers, err := s.passRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get passengers",
			slog.Int64("user_id", userID),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("failed to get passengers: %w", err)
	}

	return passengers, nil
}

func (s *PassengerService) DeletePassenger(ctx context.Context, userID, passengerID int64) error {
	if err := s.passRepo.Delete(ctx, userID, passengerID); err != nil {
		if errors.Is(err, domain.ErrPassengerNotFound) {
			return domain.ErrPassengerNotFound
		}
		s.logger.ErrorContext(ctx, "failed to delete passenger",
			slog.Int64("user_id", userID),
			slog.Int64("passenger_id", passengerID),
			slog.Any("error", err),
		)
		return fmt.Errorf("failed to delete passenger: %w", err)
	}

	s.logger.InfoContext(ctx, "passenger deleted",
		slog.Int64("passenger_id", passengerID),
		slog.Int64("user_id", userID),
	)

	return nil
}
