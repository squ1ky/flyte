package service

import (
	"context"
	"github.com/squ1ky/flyte/internal/user/domain"
	"github.com/squ1ky/flyte/internal/user/formatter"
	"github.com/squ1ky/flyte/internal/user/repository"
)

type PassengerService struct {
	passRepo repository.PassengerRepository
}

func NewPassengerService(repo repository.PassengerRepository) *PassengerService {
	return &PassengerService{passRepo: repo}
}

func (s *PassengerService) AddPassenger(ctx context.Context, p *domain.Passenger) (int64, error) {
	p.DocumentNumber = formatter.FormatDocumentNumber(p.DocumentNumber)
	p.DocumentType = formatter.FormatDocumentType(p.DocumentType)
	p.FirstName = formatter.FormatName(p.FirstName)
	p.LastName = formatter.FormatName(p.LastName)
	p.MiddleName = formatter.FormatName(p.MiddleName)

	return s.passRepo.Create(ctx, p)
}

func (s *PassengerService) GetPassengers(ctx context.Context, userID int64) ([]domain.Passenger, error) {
	return s.passRepo.GetByUserID(ctx, userID)
}

func (s *PassengerService) DeletePassenger(ctx context.Context, userID, passengerID int64) error {
	return s.passRepo.Delete(ctx, userID, passengerID)
}
