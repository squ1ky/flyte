package service

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/squ1ky/flyte/internal/booking/domain"
	"github.com/squ1ky/flyte/internal/booking/infrastructure/outbox"
)

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

// --- mock: BookingRepository ---

type mockBookingRepo struct {
	createFn        func(ctx context.Context, booking *domain.Booking, tickets []domain.Ticket) error
	getByIDFn       func(ctx context.Context, id string) (*domain.Booking, error)
	getByRefCodeFn  func(ctx context.Context, refCode string) (*domain.Booking, error)
	listByUserIDFn  func(ctx context.Context, userID int64, limit, offset int) ([]domain.Booking, error)
	updateStatusFn  func(ctx context.Context, id string, status domain.BookingStatus) error
	findExpiredFn   func(ctx context.Context, limit int) ([]domain.Booking, error)
}

func (m *mockBookingRepo) Create(ctx context.Context, booking *domain.Booking, tickets []domain.Ticket) error {
	return m.createFn(ctx, booking, tickets)
}
func (m *mockBookingRepo) GetByID(ctx context.Context, id string) (*domain.Booking, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockBookingRepo) GetByRefCode(ctx context.Context, refCode string) (*domain.Booking, error) {
	return m.getByRefCodeFn(ctx, refCode)
}
func (m *mockBookingRepo) ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]domain.Booking, error) {
	return m.listByUserIDFn(ctx, userID, limit, offset)
}
func (m *mockBookingRepo) UpdateStatus(ctx context.Context, id string, status domain.BookingStatus) error {
	return m.updateStatusFn(ctx, id, status)
}
func (m *mockBookingRepo) FindExpired(ctx context.Context, limit int) ([]domain.Booking, error) {
	return m.findExpiredFn(ctx, limit)
}

// --- mock: TicketRepository ---

type mockTicketRepo struct {
	listByBookingIDFn func(ctx context.Context, bookingID string) ([]domain.Ticket, error)
	issueTicketsFn    func(ctx context.Context, bookingID string, ticketNumbers map[string]string) error
}

func (m *mockTicketRepo) ListByBookingID(ctx context.Context, bookingID string) ([]domain.Ticket, error) {
	return m.listByBookingIDFn(ctx, bookingID)
}
func (m *mockTicketRepo) IssueTickets(ctx context.Context, bookingID string, ticketNumbers map[string]string) error {
	return m.issueTicketsFn(ctx, bookingID, ticketNumbers)
}

// --- mock: outbox.Repository ---

type mockOutboxRepo struct {
	insertFn       func(ctx context.Context, event outbox.Event) error
	fetchPendingFn func(ctx context.Context, limit int) ([]outbox.Event, error)
	markSentFn     func(ctx context.Context, id string) error
}

func (m *mockOutboxRepo) Insert(ctx context.Context, event outbox.Event) error {
	return m.insertFn(ctx, event)
}
func (m *mockOutboxRepo) FetchPending(ctx context.Context, limit int) ([]outbox.Event, error) {
	return m.fetchPendingFn(ctx, limit)
}
func (m *mockOutboxRepo) MarkSent(ctx context.Context, id string) error {
	return m.markSentFn(ctx, id)
}

// --- mock: FlightClient ---

type mockFlightClient struct {
	reserveSeatsFn      func(ctx context.Context, flightID int64, seatNumbers []string, bookingID string) ([]string, time.Time, error)
	confirmReservFn     func(ctx context.Context, flightID int64, seatNumbers []string, bookingID string) error
	cancelReservFn      func(ctx context.Context, bookingID string) error
	getFlightFaresFn    func(ctx context.Context, flightID int64) (map[domain.SeatClass]int64, error)
}

func (m *mockFlightClient) ReserveSeats(ctx context.Context, flightID int64, seatNumbers []string, bookingID string) ([]string, time.Time, error) {
	return m.reserveSeatsFn(ctx, flightID, seatNumbers, bookingID)
}
func (m *mockFlightClient) ConfirmReservation(ctx context.Context, flightID int64, seatNumbers []string, bookingID string) error {
	return m.confirmReservFn(ctx, flightID, seatNumbers, bookingID)
}
func (m *mockFlightClient) CancelReservation(ctx context.Context, bookingID string) error {
	return m.cancelReservFn(ctx, bookingID)
}
func (m *mockFlightClient) GetFlightFares(ctx context.Context, flightID int64) (map[domain.SeatClass]int64, error) {
	return m.getFlightFaresFn(ctx, flightID)
}
