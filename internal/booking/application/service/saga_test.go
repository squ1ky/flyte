package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/squ1ky/flyte/internal/booking/config"
	"github.com/squ1ky/flyte/internal/booking/domain"
	"github.com/squ1ky/flyte/internal/booking/infrastructure/outbox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validInput() CreateBookingInput {
	return CreateBookingInput{
		UserID:       1,
		FlightID:     10,
		ContactEmail: "test@example.com",
		Passengers: []PassengerInput{
			{
				FirstName:      "Ivan",
				LastName:       "Petrov",
				DocumentNumber: "1234 567890",
				SeatNumber:     "12A",
				SeatClass:      domain.SeatClassEconomy,
			},
		},
	}
}

// --- CreateBookingInput.Validate ---

func TestCreateBookingInput_Validate(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*CreateBookingInput)
		wantErr error
	}{
		{
			name:    "valid",
			mutate:  func(_ *CreateBookingInput) {},
			wantErr: nil,
		},
		{
			name:    "zero user_id",
			mutate:  func(i *CreateBookingInput) { i.UserID = 0 },
			wantErr: domain.ErrInvalidUserID,
		},
		{
			name:    "negative user_id",
			mutate:  func(i *CreateBookingInput) { i.UserID = -1 },
			wantErr: domain.ErrInvalidUserID,
		},
		{
			name:    "zero flight_id",
			mutate:  func(i *CreateBookingInput) { i.FlightID = 0 },
			wantErr: domain.ErrInvalidFlightID,
		},
		{
			name:    "empty email",
			mutate:  func(i *CreateBookingInput) { i.ContactEmail = "" },
			wantErr: domain.ErrInvalidEmail,
		},
		{
			name:    "no passengers",
			mutate:  func(i *CreateBookingInput) { i.Passengers = nil },
			wantErr: domain.ErrNoPassengers,
		},
		{
			name: "passenger empty first name",
			mutate: func(i *CreateBookingInput) {
				i.Passengers[0].FirstName = ""
			},
			wantErr: domain.ErrInvalidName,
		},
		{
			name: "passenger empty last name",
			mutate: func(i *CreateBookingInput) {
				i.Passengers[0].LastName = ""
			},
			wantErr: domain.ErrInvalidName,
		},
		{
			name: "passenger empty document",
			mutate: func(i *CreateBookingInput) {
				i.Passengers[0].DocumentNumber = ""
			},
			wantErr: domain.ErrInvalidDocument,
		},
		{
			name: "passenger empty seat number",
			mutate: func(i *CreateBookingInput) {
				i.Passengers[0].SeatNumber = ""
			},
			wantErr: domain.ErrInvalidSeatNumber,
		},
		{
			name: "passenger invalid seat class",
			mutate: func(i *CreateBookingInput) {
				i.Passengers[0].SeatClass = "INVALID"
			},
			wantErr: domain.ErrInvalidSeatClass,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := validInput()
			tc.mutate(&input)
			err := input.Validate()
			if tc.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tc.wantErr)
			}
		})
	}
}

// --- BookingSaga.CreateBooking ---

func newTestSaga(bookings domain.BookingRepository, tickets domain.TicketRepository, ob *mockOutboxRepo, flights FlightClient) *BookingSaga {
	return NewBookingSaga(bookings, tickets, ob, flights, config.CleanerConfig{BookingTTL: 15 * time.Minute}, discardLogger)
}

func happyFlightClient() *mockFlightClient {
	return &mockFlightClient{
		reserveSeatsFn: func(_ context.Context, _ int64, _ []string, _ string) ([]string, time.Time, error) {
			return nil, time.Now().Add(15 * time.Minute), nil
		},
		getFlightFaresFn: func(_ context.Context, _ int64) (map[domain.SeatClass]int64, error) {
			return map[domain.SeatClass]int64{domain.SeatClassEconomy: 5000}, nil
		},
		cancelReservFn: func(_ context.Context, _ string) error { return nil },
	}
}

func happyBookingRepo() *mockBookingRepo {
	return &mockBookingRepo{
		createFn: func(_ context.Context, _ *domain.Booking, _ []domain.Ticket) error { return nil },
	}
}

func happyOutbox() *mockOutboxRepo {
	return &mockOutboxRepo{
		insertFn: func(_ context.Context, _ outbox.Event) error { return nil },
	}
}

func TestBookingSaga_CreateBooking_HappyPath(t *testing.T) {
	saga := newTestSaga(happyBookingRepo(), &mockTicketRepo{}, happyOutbox(), happyFlightClient())

	booking, err := saga.CreateBooking(context.Background(), validInput())

	require.NoError(t, err)
	require.NotNil(t, booking)
	assert.Equal(t, int64(1), booking.UserID)
	assert.Equal(t, int64(10), booking.FlightID)
	assert.Equal(t, domain.BookingStatusPending, booking.Status)
	assert.Equal(t, int64(5000), booking.TotalPriceCents)
	assert.Len(t, booking.Tickets, 1)
}

func TestBookingSaga_CreateBooking_ValidationError(t *testing.T) {
	saga := newTestSaga(happyBookingRepo(), &mockTicketRepo{}, happyOutbox(), happyFlightClient())

	input := validInput()
	input.UserID = 0

	_, err := saga.CreateBooking(context.Background(), input)

	assert.ErrorIs(t, err, domain.ErrInvalidUserID)
}

func TestBookingSaga_CreateBooking_ReserveSeatsError(t *testing.T) {
	reserveErr := errors.New("flight service unavailable")
	fc := happyFlightClient()
	fc.reserveSeatsFn = func(_ context.Context, _ int64, _ []string, _ string) ([]string, time.Time, error) {
		return nil, time.Time{}, reserveErr
	}
	saga := newTestSaga(happyBookingRepo(), &mockTicketRepo{}, happyOutbox(), fc)

	_, err := saga.CreateBooking(context.Background(), validInput())

	assert.ErrorContains(t, err, reserveErr.Error())
}

func TestBookingSaga_CreateBooking_SeatsUnavailable(t *testing.T) {
	fc := happyFlightClient()
	fc.reserveSeatsFn = func(_ context.Context, _ int64, _ []string, _ string) ([]string, time.Time, error) {
		return []string{"12A"}, time.Time{}, nil
	}
	saga := newTestSaga(happyBookingRepo(), &mockTicketRepo{}, happyOutbox(), fc)

	_, err := saga.CreateBooking(context.Background(), validInput())

	assert.ErrorIs(t, err, domain.ErrSeatsUnavailable)
}

func TestBookingSaga_CreateBooking_GetFaresError(t *testing.T) {
	faresErr := errors.New("fares unavailable")
	fc := happyFlightClient()
	fc.getFlightFaresFn = func(_ context.Context, _ int64) (map[domain.SeatClass]int64, error) {
		return nil, faresErr
	}
	cancelCalled := false
	fc.cancelReservFn = func(_ context.Context, _ string) error {
		cancelCalled = true
		return nil
	}
	saga := newTestSaga(happyBookingRepo(), &mockTicketRepo{}, happyOutbox(), fc)

	_, err := saga.CreateBooking(context.Background(), validInput())

	assert.ErrorContains(t, err, faresErr.Error())
	assert.True(t, cancelCalled, "compensation must be called when fares fail")
}

func TestBookingSaga_CreateBooking_NoFareForClass(t *testing.T) {
	fc := happyFlightClient()
	fc.getFlightFaresFn = func(_ context.Context, _ int64) (map[domain.SeatClass]int64, error) {
		return map[domain.SeatClass]int64{domain.SeatClassBusiness: 10000}, nil
	}
	cancelCalled := false
	fc.cancelReservFn = func(_ context.Context, _ string) error {
		cancelCalled = true
		return nil
	}
	saga := newTestSaga(happyBookingRepo(), &mockTicketRepo{}, happyOutbox(), fc)

	_, err := saga.CreateBooking(context.Background(), validInput())

	assert.ErrorContains(t, err, "no fare for class")
	assert.True(t, cancelCalled)
}

func TestBookingSaga_CreateBooking_SaveBookingError(t *testing.T) {
	saveErr := errors.New("db error")
	br := happyBookingRepo()
	br.createFn = func(_ context.Context, _ *domain.Booking, _ []domain.Ticket) error {
		return saveErr
	}
	cancelCalled := false
	fc := happyFlightClient()
	fc.cancelReservFn = func(_ context.Context, _ string) error {
		cancelCalled = true
		return nil
	}
	saga := newTestSaga(br, &mockTicketRepo{}, happyOutbox(), fc)

	_, err := saga.CreateBooking(context.Background(), validInput())

	assert.ErrorContains(t, err, "save booking")
	assert.True(t, cancelCalled)
}

func TestBookingSaga_CreateBooking_EnqueuePaymentError(t *testing.T) {
	ob := &mockOutboxRepo{
		insertFn: func(_ context.Context, _ outbox.Event) error {
			return errors.New("outbox insert failed")
		},
	}
	saga := newTestSaga(happyBookingRepo(), &mockTicketRepo{}, ob, happyFlightClient())

	_, err := saga.CreateBooking(context.Background(), validInput())

	assert.ErrorContains(t, err, "enqueue payment")
}


