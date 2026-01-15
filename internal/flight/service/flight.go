package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/squ1ky/flyte/internal/flight/domain"
	"github.com/squ1ky/flyte/internal/flight/repository"
	"log/slog"
	"time"
)

type FlightService struct {
	flightStorage  repository.FlightStorage
	flightSearcher repository.FlightSearcher
	logger         *slog.Logger
}

func NewFlightService(
	flightStorage repository.FlightStorage,
	flightSearcher repository.FlightSearcher,
	logger *slog.Logger,
) *FlightService {
	return &FlightService{
		flightStorage:  flightStorage,
		flightSearcher: flightSearcher,
		logger:         logger,
	}
}

func (s *FlightService) CreateFlight(ctx context.Context, f *domain.Flight) (int64, error) {
	id, err := s.flightStorage.CreateFlight(ctx, f)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *FlightService) SearchFlights(ctx context.Context, from, to string, date time.Time, passengerCount int) ([]domain.Flight, error) {
	flights, err := s.flightSearcher.Search(ctx, from, to, date, passengerCount)
	if err != nil {
		s.logger.Error("failed to search flights in elastic", "error", err)
		return nil, fmt.Errorf("search failed: %w", err)
	}
	return flights, nil
}

func (s *FlightService) GetFlightDetails(ctx context.Context, flightID int64) (*domain.Flight, error) {
	flight, err := s.flightStorage.GetByID(ctx, flightID)
	if err != nil {
		if errors.Is(err, domain.ErrFlightNotFound) {
			return nil, err
		}
		s.logger.Error("failed to get flight details", "flight_id", flightID, "error", err)
		return nil, fmt.Errorf("get details failed: %w", err)
	}
	return flight, nil
}

func (s *FlightService) GetFlightSeats(ctx context.Context, flightID int64) ([]domain.Seat, error) {
	seats, err := s.flightStorage.GetSeatsByFlightID(ctx, flightID)
	if err != nil {
		if errors.Is(err, domain.ErrFlightNotFound) {
			return nil, err
		}
		s.logger.Error("failed to get flight seats", "flight_id", flightID, "error", err)
		return nil, fmt.Errorf("get seats failed: %w", err)
	}
	return seats, nil
}

func (s *FlightService) GetAirports(ctx context.Context) ([]domain.Airport, error) {
	return s.flightStorage.GetAirports(ctx)
}

func (s *FlightService) ReserveSeat(ctx context.Context, flightID int64, seatNumber string) (int64, error) {
	seatID, err := s.flightStorage.BookSeat(ctx, flightID, seatNumber)
	if err != nil {
		return 0, err
	}

	s.runAsyncSyncElastic(flightID)
	return seatID, nil
}

func (s *FlightService) ReleaseSeat(ctx context.Context, flightID int64, seatNumber string) error {
	if err := s.flightStorage.ReleaseSeat(ctx, flightID, seatNumber); err != nil {
		return err
	}

	s.runAsyncSyncElastic(flightID)
	return nil
}

func (s *FlightService) runAsyncSyncElastic(flightID int64) {
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.syncSeatsInElastic(bgCtx, flightID); err != nil {
			s.logger.Error("async elastic sync failed", "flight_id", flightID, "error", err)
		}
	}()
}

func (s *FlightService) syncSeatsInElastic(ctx context.Context, flightID int64) error {
	flight, err := s.flightStorage.GetByID(ctx, flightID)
	if err != nil {
		return err
	}
	if flight == nil {
		return errors.New("flight disappeared from db")
	}

	return s.flightSearcher.UpdateAvailableSeats(ctx, flightID, flight.AvailableSeats)
}
