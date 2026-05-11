package service

import (
	"context"
	"errors"
	"testing"

	"github.com/squ1ky/flyte/internal/booking/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBookingQueryService_GetBooking(t *testing.T) {
	booking := &domain.Booking{ID: "bk-1", RefCode: "ABC123"}
	tickets := []domain.Ticket{{ID: "t-1", BookingID: "bk-1"}}

	cases := []struct {
		name          string
		bookingID     string
		refCode       string
		getByIDResult *domain.Booking
		getByIDErr    error
		getByRefResult *domain.Booking
		getByRefErr   error
		ticketsResult []domain.Ticket
		ticketsErr    error
		wantErr       error
		wantNil       bool
	}{
		{
			name:          "by booking_id",
			bookingID:     "bk-1",
			getByIDResult: booking,
			ticketsResult: tickets,
		},
		{
			name:           "by ref_code",
			refCode:        "ABC123",
			getByRefResult: booking,
			ticketsResult:  tickets,
		},
		{
			name:    "both empty returns not found",
			wantErr: domain.ErrBookingNotFound,
			wantNil: true,
		},
		{
			name:       "booking not found by id",
			bookingID:  "missing",
			getByIDErr: domain.ErrBookingNotFound,
			wantErr:    domain.ErrBookingNotFound,
			wantNil:    true,
		},
		{
			name:          "list tickets error",
			bookingID:     "bk-1",
			getByIDResult: booking,
			ticketsErr:    errors.New("db error"),
			wantErr:       errors.New("list tickets"),
			wantNil:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			br := &mockBookingRepo{
				getByIDFn: func(_ context.Context, _ string) (*domain.Booking, error) {
					return tc.getByIDResult, tc.getByIDErr
				},
				getByRefCodeFn: func(_ context.Context, _ string) (*domain.Booking, error) {
					return tc.getByRefResult, tc.getByRefErr
				},
			}
			tr := &mockTicketRepo{
				listByBookingIDFn: func(_ context.Context, _ string) ([]domain.Ticket, error) {
					return tc.ticketsResult, tc.ticketsErr
				},
			}

			svc := NewBookingQueryService(br, tr, discardLogger)
			result, err := svc.GetBooking(context.Background(), tc.bookingID, tc.refCode)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.wantErr.Error())
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tickets, result.Tickets)
			}
		})
	}
}

func TestBookingQueryService_ListBookings(t *testing.T) {
	twoBookings := []domain.Booking{{ID: "1"}, {ID: "2"}}

	cases := []struct {
		name       string
		limit      int
		offset     int
		wantLimit  int
		wantOffset int
	}{
		{
			name:      "normal",
			limit:     10,
			wantLimit: 10,
		},
		{
			name:      "zero limit uses default",
			limit:     0,
			wantLimit: defaultLimit,
		},
		{
			name:      "negative limit uses default",
			limit:     -1,
			wantLimit: defaultLimit,
		},
		{
			name:      "over max limit capped",
			limit:     200,
			wantLimit: maxLimit,
		},
		{
			name:       "offset passed through",
			limit:      5,
			offset:     20,
			wantLimit:  5,
			wantOffset: 20,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var gotLimit, gotOffset int
			br := &mockBookingRepo{
				listByUserIDFn: func(_ context.Context, _ int64, limit, offset int) ([]domain.Booking, error) {
					gotLimit = limit
					gotOffset = offset
					return twoBookings, nil
				},
			}
			svc := NewBookingQueryService(br, &mockTicketRepo{}, discardLogger)

			result, err := svc.ListBookings(context.Background(), 1, tc.limit, tc.offset)

			require.NoError(t, err)
			assert.Equal(t, twoBookings, result)
			assert.Equal(t, tc.wantLimit, gotLimit)
			assert.Equal(t, tc.wantOffset, gotOffset)
		})
	}
}
