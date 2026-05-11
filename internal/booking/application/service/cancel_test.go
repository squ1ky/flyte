package service

import (
	"context"
	"errors"
	"testing"

	"github.com/squ1ky/flyte/internal/booking/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCancelService_CancelBooking(t *testing.T) {
	pendingBooking := &domain.Booking{ID: "bk-1", Status: domain.BookingStatusPending}
	paidBooking := &domain.Booking{ID: "bk-2", Status: domain.BookingStatusPaid}
	cancelledBooking := &domain.Booking{ID: "bk-3", Status: domain.BookingStatusCancelled}

	cases := []struct {
		name           string
		bookingID      string
		getByIDResult  *domain.Booking
		getByIDErr     error
		updateStatusErr error
		cancelReservErr error
		wantErr        error
		cancelCalled   bool
	}{
		{
			name:          "happy path",
			bookingID:     "bk-1",
			getByIDResult: pendingBooking,
			cancelCalled:  true,
		},
		{
			name:       "booking not found",
			bookingID:  "missing",
			getByIDErr: domain.ErrBookingNotFound,
			wantErr:    domain.ErrBookingNotFound,
		},
		{
			name:          "already cancelled",
			bookingID:     "bk-3",
			getByIDResult: cancelledBooking,
			wantErr:       domain.ErrBookingAlreadyCancelled,
		},
		{
			name:          "already paid",
			bookingID:     "bk-2",
			getByIDResult: paidBooking,
			wantErr:       domain.ErrBookingAlreadyPaid,
		},
		{
			name:            "update status error",
			bookingID:       "bk-1",
			getByIDResult:   pendingBooking,
			updateStatusErr: errors.New("db error"),
			wantErr:         errors.New("update status"),
		},
		{
			name:            "cancel reservation error is non-fatal",
			bookingID:       "bk-1",
			getByIDResult:   pendingBooking,
			cancelReservErr: errors.New("flight service down"),
			cancelCalled:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cancelCalled := false

			br := &mockBookingRepo{
				getByIDFn: func(_ context.Context, id string) (*domain.Booking, error) {
					return tc.getByIDResult, tc.getByIDErr
				},
				updateStatusFn: func(_ context.Context, _ string, _ domain.BookingStatus) error {
					return tc.updateStatusErr
				},
			}
			fc := &mockFlightClient{
				cancelReservFn: func(_ context.Context, _ string) error {
					cancelCalled = true
					return tc.cancelReservErr
				},
			}

			svc := NewCancelService(br, fc, discardLogger)
			err := svc.CancelBooking(context.Background(), tc.bookingID)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.wantErr.Error())
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.cancelCalled, cancelCalled)
		})
	}
}
