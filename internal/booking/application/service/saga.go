package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/squ1ky/flyte/internal/booking/config"
	"github.com/squ1ky/flyte/internal/booking/domain"
	"github.com/squ1ky/flyte/internal/booking/infrastructure/outbox"
	"log/slog"
	"math/big"
	"time"
)

const (
	refCodeLength   = 6
	refCodeAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type CreateBookingInput struct {
	UserID       int64
	FlightID     int64
	ContactEmail string
	Passengers   []PassengerInput
}

func (i *CreateBookingInput) Validate() error {
	if i.UserID <= 0 {
		return domain.ErrInvalidUserID
	}
	if i.FlightID <= 0 {
		return domain.ErrInvalidFlightID
	}
	if i.ContactEmail == "" {
		return domain.ErrInvalidEmail
	}
	if len(i.Passengers) == 0 {
		return domain.ErrNoPassengers
	}

	for _, passenger := range i.Passengers {
		if passenger.FirstName == "" || passenger.LastName == "" {
			return domain.ErrInvalidName
		}
		if passenger.DocumentNumber == "" {
			return domain.ErrInvalidDocument
		}
		if passenger.SeatNumber == "" {
			return domain.ErrInvalidSeatNumber
		}
		if !passenger.SeatClass.IsValid() {
			return domain.ErrInvalidSeatClass
		}
	}
	return nil
}

type PassengerInput struct {
	FirstName      string
	LastName       string
	DocumentNumber string
	SeatNumber     string
	SeatClass      domain.SeatClass
}

type BookingSaga struct {
	bookings   domain.BookingRepository
	tickets    domain.TicketRepository
	outbox     outbox.Repository
	flights    FlightClient
	bookingTTL time.Duration
	logger     *slog.Logger
}

func NewBookingSaga(
	bookings domain.BookingRepository,
	tickets domain.TicketRepository,
	outbox outbox.Repository,
	flights FlightClient,
	cfg config.CleanerConfig,
	logger *slog.Logger,
) *BookingSaga {
	return &BookingSaga{
		bookings:   bookings,
		tickets:    tickets,
		outbox:     outbox,
		flights:    flights,
		bookingTTL: cfg.BookingTTL,
		logger:     logger,
	}
}

func (s *BookingSaga) CreateBooking(ctx context.Context, input CreateBookingInput) (*domain.Booking, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	log := s.logger.With(
		"user_id", input.UserID,
		"flight_id", input.FlightID,
	)

	bookingID := uuid.New().String()

	refCode, err := generateRefCode()
	if err != nil {
		return nil, fmt.Errorf("generate ref code: %w", err)
	}

	seatNumbers := extractSeatNumbers(input.Passengers)

	expiresAt, err := s.reserveSeats(ctx, input.FlightID, seatNumbers, bookingID, log)
	if err != nil {
		return nil, err
	}

	tickets, totalPrice, err := s.buildTickets(ctx, bookingID, input, log)
	if err != nil {
		s.compensateSeatReservation(ctx, bookingID, log)
		return nil, err
	}

	booking := &domain.Booking{
		ID:              bookingID,
		UserID:          input.UserID,
		FlightID:        input.FlightID,
		RefCode:         refCode,
		Status:          domain.BookingStatusPending,
		TotalPriceCents: totalPrice,
		Currency:        "RUB", // TODO: remove hardcode
		ContactEmail:    input.ContactEmail,
		ExpiresAt:       expiresAt,
		CreatedAt:       time.Now(),
		Tickets:         tickets,
	}

	if err := s.bookings.Create(ctx, booking, tickets); err != nil {
		s.compensateSeatReservation(ctx, bookingID, log)
		return nil, fmt.Errorf("save booking: %w", err)
	}

	if err := s.enqueuePayment(ctx, booking); err != nil {
		return nil, fmt.Errorf("enqueue payment: %w", err)
	}

	log.InfoContext(ctx, "booking created",
		slog.String("booking_id", bookingID),
		slog.String("ref_code", refCode),
		slog.Int64("total_price_cents", totalPrice),
	)

	return booking, nil
}

func (s *BookingSaga) reserveSeats(ctx context.Context, flightID int64, seatNumbers []string, bookingID string, log *slog.Logger) (time.Time, error) {
	log.InfoContext(ctx, "reserving seats",
		slog.Any("seats", seatNumbers),
	)

	unavailableSeats, expiresAt, err := s.flights.ReserveSeats(ctx, flightID, seatNumbers, bookingID)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to reserve seats: %w", err)
	}

	if len(unavailableSeats) > 0 {
		log.WarnContext(ctx, "seats unavailable",
			slog.Any("unavailable_seats", unavailableSeats),
		)
		return time.Time{}, fmt.Errorf("%w: %v", domain.ErrSeatsUnavailable, unavailableSeats)
	}

	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(s.bookingTTL)
	}

	return expiresAt, nil
}

func (s *BookingSaga) buildTickets(ctx context.Context, bookingID string, input CreateBookingInput, log *slog.Logger) (tickets []domain.Ticket, totalPriceCents int64, err error) {
	fareMap, err := s.flights.GetFlightFares(ctx, input.FlightID)
	if err != nil {
		return nil, 0, fmt.Errorf("get fares: %w", err)
	}

	var totalPrice int64
	tickets = make([]domain.Ticket, len(input.Passengers))

	for i, passenger := range input.Passengers {
		price, ok := fareMap[passenger.SeatClass]
		if !ok {
			return nil, 0, fmt.Errorf("no fare for class %s (seat %s)", passenger.SeatClass, passenger.SeatNumber)
		}

		tickets[i] = domain.Ticket{
			ID:                 uuid.New().String(),
			BookingID:          bookingID,
			PassengerFirstName: passenger.FirstName,
			PassengerLastName:  passenger.LastName,
			PassengerDocNum:    passenger.DocumentNumber,
			SeatNumber:         passenger.SeatNumber,
			PriceCents:         price,
			Status:             domain.TicketStatusReserved,
		}
		totalPrice += price
	}

	log.InfoContext(ctx, "tickets built",
		slog.Int("count", len(tickets)),
		slog.Int64("total_price", totalPrice),
	)
	return tickets, totalPrice, nil
}

func (s *BookingSaga) enqueuePayment(ctx context.Context, booking *domain.Booking) error {
	event := domain.PaymentRequestEvent{
		BookingID:   booking.ID,
		UserID:      booking.UserID,
		AmountCents: booking.TotalPriceCents,
		Currency:    booking.Currency,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal payment event: %w", err)
	}

	return s.outbox.Insert(ctx, outbox.Event{
		ID:        uuid.New().String(),
		EventType: outbox.EventPaymentRequest,
		Payload:   payload,
		Status:    outbox.StatusPending,
		CreatedAt: time.Now(),
	})
}

func (s *BookingSaga) compensateSeatReservation(ctx context.Context, bookingID string, log *slog.Logger) {
	if err := s.flights.CancelReservation(ctx, bookingID); err != nil {
		log.ErrorContext(ctx, "failed to cancel seat reservation (compensation)",
			slog.String("booking_id", bookingID),
			slog.Any("error", err),
		)
	}
}

func extractSeatNumbers(passengers []PassengerInput) []string {
	seats := make([]string, len(passengers))
	for i, passenger := range passengers {
		seats[i] = passenger.SeatNumber
	}
	return seats
}

func generateRefCode() (string, error) {
	code := make([]byte, refCodeLength)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(refCodeAlphabet))))
		if err != nil {
			return "", err
		}
		code[i] = refCodeAlphabet[n.Int64()]
	}
	return string(code), nil
}
