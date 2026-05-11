package pgrepo_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/squ1ky/flyte/internal/booking/domain"
	"github.com/squ1ky/flyte/internal/booking/infrastructure/repository/pgrepo"
)

func newBooking(userID int64, refCode string, status domain.BookingStatus, expiresAt time.Time) *domain.Booking {
	return &domain.Booking{
		ID:              uuid.NewString(),
		UserID:          userID,
		FlightID:        42,
		RefCode:         refCode,
		Status:          status,
		TotalPriceCents: 10000,
		Currency:        "RUB",
		ContactEmail:    "test@example.com",
		ExpiresAt:       expiresAt.UTC().Truncate(time.Microsecond),
		CreatedAt:       time.Now().UTC().Truncate(time.Microsecond),
	}
}

func newTickets(bookingID string, seats ...string) []domain.Ticket {
	tickets := make([]domain.Ticket, len(seats))
	for i, seat := range seats {
		tickets[i] = domain.Ticket{
			ID:                 uuid.NewString(),
			BookingID:          bookingID,
			PassengerFirstName: "Ivan",
			PassengerLastName:  "Petrov",
			PassengerDocNum:    "1234567890",
			SeatNumber:         seat,
			PriceCents:         5000,
			Status:             domain.TicketStatusReserved,
		}
	}
	return tickets
}

func TestBookingRepo_Create(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()
	repo := pgrepo.NewBookingRepo(testDB)

	b := newBooking(1, "ABC123", domain.BookingStatusPending, time.Now().Add(15*time.Minute))
	tickets := newTickets(b.ID, "12A", "12B")

	err := repo.Create(ctx, b, tickets)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, b.ID)
	require.NoError(t, err)
	assert.Equal(t, b.ID, got.ID)
	assert.Equal(t, b.RefCode, got.RefCode)
	assert.Equal(t, domain.BookingStatusPending, got.Status)
}

func TestBookingRepo_Create_RollsBackOnDuplicateRef(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()
	repo := pgrepo.NewBookingRepo(testDB)

	b1 := newBooking(1, "SAME00", domain.BookingStatusPending, time.Now().Add(15*time.Minute))
	require.NoError(t, repo.Create(ctx, b1, nil))

	b2 := newBooking(2, "SAME00", domain.BookingStatusPending, time.Now().Add(15*time.Minute))
	err := repo.Create(ctx, b2, nil)
	require.Error(t, err)

	_, err = repo.GetByID(ctx, b2.ID)
	assert.ErrorIs(t, err, domain.ErrBookingNotFound)
}

func TestBookingRepo_GetByID_NotFound(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()
	repo := pgrepo.NewBookingRepo(testDB)

	_, err := repo.GetByID(ctx, uuid.NewString())
	assert.ErrorIs(t, err, domain.ErrBookingNotFound)
}

func TestBookingRepo_GetByRefCode(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()
	repo := pgrepo.NewBookingRepo(testDB)

	b := newBooking(1, "REF001", domain.BookingStatusPending, time.Now().Add(15*time.Minute))
	require.NoError(t, repo.Create(ctx, b, nil))

	got, err := repo.GetByRefCode(ctx, "REF001")
	require.NoError(t, err)
	assert.Equal(t, b.ID, got.ID)
}

func TestBookingRepo_GetByRefCode_NotFound(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()
	repo := pgrepo.NewBookingRepo(testDB)

	_, err := repo.GetByRefCode(ctx, "NOEXST")
	assert.ErrorIs(t, err, domain.ErrBookingNotFound)
}

func TestBookingRepo_ListByUserID(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()
	repo := pgrepo.NewBookingRepo(testDB)

	const userID int64 = 7
	for i, ref := range []string{"AA0001", "AA0002", "AA0003"} {
		b := newBooking(userID, ref, domain.BookingStatusPending, time.Now().Add(time.Duration(i+1)*time.Hour))
		require.NoError(t, repo.Create(ctx, b, nil))
	}
	// booking for a different user — should not appear
	other := newBooking(99, "ZZ0001", domain.BookingStatusPending, time.Now().Add(time.Hour))
	require.NoError(t, repo.Create(ctx, other, nil))

	list, err := repo.ListByUserID(ctx, userID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, list, 3)
	for _, b := range list {
		assert.Equal(t, userID, b.UserID)
	}
}

func TestBookingRepo_ListByUserID_Pagination(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()
	repo := pgrepo.NewBookingRepo(testDB)

	const userID int64 = 5
	for _, ref := range []string{"PG0001", "PG0002", "PG0003"} {
		b := newBooking(userID, ref, domain.BookingStatusPending, time.Now().Add(time.Hour))
		require.NoError(t, repo.Create(ctx, b, nil))
	}

	page1, err := repo.ListByUserID(ctx, userID, 2, 0)
	require.NoError(t, err)
	assert.Len(t, page1, 2)

	page2, err := repo.ListByUserID(ctx, userID, 2, 2)
	require.NoError(t, err)
	assert.Len(t, page2, 1)
}

func TestBookingRepo_UpdateStatus(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()
	repo := pgrepo.NewBookingRepo(testDB)

	b := newBooking(1, "UPD001", domain.BookingStatusPending, time.Now().Add(15*time.Minute))
	require.NoError(t, repo.Create(ctx, b, nil))

	require.NoError(t, repo.UpdateStatus(ctx, b.ID, domain.BookingStatusPaid))

	got, err := repo.GetByID(ctx, b.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.BookingStatusPaid, got.Status)
}

func TestBookingRepo_UpdateStatus_NotFound(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()
	repo := pgrepo.NewBookingRepo(testDB)

	err := repo.UpdateStatus(ctx, uuid.NewString(), domain.BookingStatusPaid)
	assert.ErrorIs(t, err, domain.ErrBookingNotFound)
}

func TestBookingRepo_FindExpired(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()
	repo := pgrepo.NewBookingRepo(testDB)

	expired := newBooking(1, "EXP001", domain.BookingStatusPending, time.Now().Add(-1*time.Minute))
	notExpired := newBooking(1, "EXP002", domain.BookingStatusPending, time.Now().Add(1*time.Hour))
	paid := newBooking(1, "EXP003", domain.BookingStatusPaid, time.Now().Add(-1*time.Minute))

	require.NoError(t, repo.Create(ctx, expired, nil))
	require.NoError(t, repo.Create(ctx, notExpired, nil))
	require.NoError(t, repo.Create(ctx, paid, nil))

	found, err := repo.FindExpired(ctx, 10)
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, expired.ID, found[0].ID)
}

func TestBookingRepo_FindExpired_RespectsLimit(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()
	repo := pgrepo.NewBookingRepo(testDB)

	for _, ref := range []string{"LM0001", "LM0002", "LM0003"} {
		b := newBooking(1, ref, domain.BookingStatusPending, time.Now().Add(-1*time.Minute))
		require.NoError(t, repo.Create(ctx, b, nil))
	}

	found, err := repo.FindExpired(ctx, 2)
	require.NoError(t, err)
	assert.Len(t, found, 2)
}
