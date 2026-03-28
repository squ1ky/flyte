package service

import (
	"context"
	"fmt"
	"github.com/squ1ky/flyte/internal/booking/domain"
	"log/slog"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type BookingQueryService struct {
	bookings domain.BookingRepository
	tickets  domain.TicketRepository
	logger   *slog.Logger
}

func NewBookingQueryService(
	bookings domain.BookingRepository,
	tickets domain.TicketRepository,
	logger *slog.Logger,
) *BookingQueryService {
	return &BookingQueryService{
		bookings: bookings,
		tickets:  tickets,
		logger:   logger,
	}
}

func (s *BookingQueryService) GetBooking(ctx context.Context, bookingID, refCode string) (*domain.Booking, error) {
	var (
		booking *domain.Booking
		err     error
	)

	if bookingID != "" {
		booking, err = s.bookings.GetByID(ctx, bookingID)
	} else if refCode != "" {
		booking, err = s.bookings.GetByRefCode(ctx, refCode)
	} else {
		return nil, domain.ErrBookingNotFound
	}

	if err != nil {
		return nil, err
	}

	tickets, err := s.tickets.ListByBookingID(ctx, booking.ID)
	if err != nil {
		return nil, fmt.Errorf("list tickets: %w", err)
	}
	booking.Tickets = tickets

	return booking, nil
}

func (s *BookingQueryService) ListBookings(ctx context.Context, userID int64, limit, offset int) ([]domain.Booking, error) {
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > 100 {
		limit = maxLimit
	}
	return s.bookings.ListByUserID(ctx, userID, limit, offset)
}
