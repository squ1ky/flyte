package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/squ1ky/flyte/internal/booking/domain"
	"github.com/stretchr/testify/assert"
)

func newCleaner(br *mockBookingRepo, fc *mockFlightClient) *ExpiredBookingCleaner {
	return &ExpiredBookingCleaner{
		bookings: br,
		flights:  fc,
		logger:   discardLogger,
	}
}

func TestExpiredBookingCleaner_ExpireBatch(t *testing.T) {
	expired := []domain.Booking{
		{ID: "bk-1"},
		{ID: "bk-2"},
		{ID: "bk-3"},
	}

	cases := []struct {
		name            string
		findExpiredResult []domain.Booking
		findExpiredErr  error
		updateStatusErr error
		cancelReservErr error
		wantCancelled   int
	}{
		{
			name:              "empty batch does nothing",
			findExpiredResult: nil,
		},
		{
			name:              "happy path: all cancelled",
			findExpiredResult: expired,
			wantCancelled:     3,
		},
		{
			name:           "find expired error: no panic",
			findExpiredErr: errors.New("db error"),
			wantCancelled:  0,
		},
		{
			name:              "update status error: skips booking, continues",
			findExpiredResult: expired,
			updateStatusErr:   errors.New("db error"),
			wantCancelled:     0,
		},
		{
			name:              "cancel reservation error is non-fatal",
			findExpiredResult: expired,
			cancelReservErr:   errors.New("flight service down"),
			wantCancelled:     3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cancelled := 0

			br := &mockBookingRepo{
				findExpiredFn: func(_ context.Context, _ int) ([]domain.Booking, error) {
					return tc.findExpiredResult, tc.findExpiredErr
				},
				updateStatusFn: func(_ context.Context, _ string, _ domain.BookingStatus) error {
					return tc.updateStatusErr
				},
			}
			fc := &mockFlightClient{
				cancelReservFn: func(_ context.Context, _ string) error {
					cancelled++
					return tc.cancelReservErr
				},
			}

			c := newCleaner(br, fc)
			c.expireBatch(context.Background())

			assert.Equal(t, tc.wantCancelled, cancelled)
		})
	}
}
