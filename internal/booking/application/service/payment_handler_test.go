package service

import (
	"context"
	"errors"
	"testing"

	"github.com/squ1ky/flyte/internal/booking/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaymentHandler_HandlePaymentResult(t *testing.T) {
	pendingBooking := &domain.Booking{ID: "bk-1", FlightID: 10, Status: domain.BookingStatusPending}
	paidBooking := &domain.Booking{ID: "bk-2", Status: domain.BookingStatusPaid}
	cancelledBooking := &domain.Booking{ID: "bk-3", Status: domain.BookingStatusCancelled}
	tickets := []domain.Ticket{{ID: "t-1", SeatNumber: "12A"}}

	cases := []struct {
		name            string
		event           domain.PaymentResultEvent
		booking         *domain.Booking
		getByIDErr      error
		ticketsResult   []domain.Ticket
		ticketsErr      error
		confirmErr      error
		issueErr        error
		updateStatusErr error
		cancelReservErr error
		wantErr         bool
		wantErrContains string
	}{
		{
			name:          "success: booking confirmed",
			event:         domain.PaymentResultEvent{BookingID: "bk-1", Status: domain.PaymentStatusSuccess},
			booking:       pendingBooking,
			ticketsResult: tickets,
		},
		{
			name:          "failure: booking cancelled",
			event:         domain.PaymentResultEvent{BookingID: "bk-1", Status: domain.PaymentStatusFailed},
			booking:       pendingBooking,
			ticketsResult: tickets,
		},
		{
			name:            "booking not found",
			event:           domain.PaymentResultEvent{BookingID: "missing", Status: domain.PaymentStatusSuccess},
			getByIDErr:      domain.ErrBookingNotFound,
			wantErr:         true,
			wantErrContains: "get booking",
		},
		{
			name:    "terminal status: already paid → skip idempotent",
			event:   domain.PaymentResultEvent{BookingID: "bk-2", Status: domain.PaymentStatusSuccess},
			booking: paidBooking,
		},
		{
			name:    "terminal status: already cancelled → skip idempotent",
			event:   domain.PaymentResultEvent{BookingID: "bk-3", Status: domain.PaymentStatusFailed},
			booking: cancelledBooking,
		},
		{
			name:            "unknown payment status",
			event:           domain.PaymentResultEvent{BookingID: "bk-1", Status: "UNKNOWN"},
			booking:         pendingBooking,
			wantErr:         true,
			wantErrContains: "unknown payment status",
		},
		{
			name:            "success: list tickets error",
			event:           domain.PaymentResultEvent{BookingID: "bk-1", Status: domain.PaymentStatusSuccess},
			booking:         pendingBooking,
			ticketsErr:      errors.New("db error"),
			wantErr:         true,
			wantErrContains: "list tickets",
		},
		{
			name:            "success: confirm reservation error",
			event:           domain.PaymentResultEvent{BookingID: "bk-1", Status: domain.PaymentStatusSuccess},
			booking:         pendingBooking,
			ticketsResult:   tickets,
			confirmErr:      errors.New("grpc error"),
			wantErr:         true,
			wantErrContains: "confirm reservation",
		},
		{
			name:            "success: issue tickets error",
			event:           domain.PaymentResultEvent{BookingID: "bk-1", Status: domain.PaymentStatusSuccess},
			booking:         pendingBooking,
			ticketsResult:   tickets,
			issueErr:        errors.New("db error"),
			wantErr:         true,
			wantErrContains: "issue tickets",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			br := &mockBookingRepo{
				getByIDFn: func(_ context.Context, _ string) (*domain.Booking, error) {
					return tc.booking, tc.getByIDErr
				},
				updateStatusFn: func(_ context.Context, _ string, _ domain.BookingStatus) error {
					return tc.updateStatusErr
				},
			}
			tr := &mockTicketRepo{
				listByBookingIDFn: func(_ context.Context, _ string) ([]domain.Ticket, error) {
					return tc.ticketsResult, tc.ticketsErr
				},
				issueTicketsFn: func(_ context.Context, _ string, _ map[string]string) error {
					return tc.issueErr
				},
			}
			fc := &mockFlightClient{
				confirmReservFn: func(_ context.Context, _ int64, _ []string, _ string) error {
					return tc.confirmErr
				},
				cancelReservFn: func(_ context.Context, _ string) error {
					return tc.cancelReservErr
				},
			}

			h := NewPaymentHandler(br, tr, fc, discardLogger)
			err := h.HandlePaymentResult(context.Background(), tc.event)

			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrContains != "" {
					assert.ErrorContains(t, err, tc.wantErrContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
