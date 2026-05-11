package worker

import (
	"context"
	"io"
	"log/slog"

	"github.com/squ1ky/flyte/internal/booking/domain"
	"github.com/squ1ky/flyte/internal/booking/infrastructure/outbox"
)

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

// --- mock: BookingRepository (only methods used by workers) ---

type mockBookingRepo struct {
	updateStatusFn func(ctx context.Context, id string, status domain.BookingStatus) error
	findExpiredFn  func(ctx context.Context, limit int) ([]domain.Booking, error)
}

func (m *mockBookingRepo) Create(_ context.Context, _ *domain.Booking, _ []domain.Ticket) error {
	panic("not implemented")
}
func (m *mockBookingRepo) GetByID(_ context.Context, _ string) (*domain.Booking, error) {
	panic("not implemented")
}
func (m *mockBookingRepo) GetByRefCode(_ context.Context, _ string) (*domain.Booking, error) {
	panic("not implemented")
}
func (m *mockBookingRepo) ListByUserID(_ context.Context, _ int64, _, _ int) ([]domain.Booking, error) {
	panic("not implemented")
}
func (m *mockBookingRepo) UpdateStatus(ctx context.Context, id string, status domain.BookingStatus) error {
	return m.updateStatusFn(ctx, id, status)
}
func (m *mockBookingRepo) FindExpired(ctx context.Context, limit int) ([]domain.Booking, error) {
	return m.findExpiredFn(ctx, limit)
}

// --- mock: FlightClient ---

type mockFlightClient struct {
	cancelReservFn func(ctx context.Context, bookingID string) error
}

func (m *mockFlightClient) CancelReservation(ctx context.Context, bookingID string) error {
	return m.cancelReservFn(ctx, bookingID)
}

// --- mock: PaymentProducer ---

type mockPaymentProducer struct {
	sendFn func(ctx context.Context, event domain.PaymentRequestEvent) error
}

func (m *mockPaymentProducer) SendPaymentRequest(ctx context.Context, event domain.PaymentRequestEvent) error {
	return m.sendFn(ctx, event)
}

// --- mock: outbox.Repository ---

type mockOutboxRepo struct {
	fetchPendingFn func(ctx context.Context, limit int) ([]outbox.Event, error)
	markSentFn     func(ctx context.Context, id string) error
}

func (m *mockOutboxRepo) Insert(_ context.Context, _ outbox.Event) error {
	panic("not implemented")
}
func (m *mockOutboxRepo) FetchPending(ctx context.Context, limit int) ([]outbox.Event, error) {
	return m.fetchPendingFn(ctx, limit)
}
func (m *mockOutboxRepo) MarkSent(ctx context.Context, id string) error {
	return m.markSentFn(ctx, id)
}
