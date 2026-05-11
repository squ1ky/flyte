package pgrepo_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/squ1ky/flyte/internal/booking/domain"
	"github.com/squ1ky/flyte/internal/booking/infrastructure/repository/pgrepo"
)

func TestTicketRepo_ListByBookingID(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()

	bookingRepo := pgrepo.NewBookingRepo(testDB)
	ticketRepo := pgrepo.NewTicketRepo(testDB)

	b := newBooking(1, "TKT001", domain.BookingStatusPending, time.Now().Add(15*time.Minute))
	tickets := newTickets(b.ID, "1A", "1B", "1C")
	require.NoError(t, bookingRepo.Create(ctx, b, tickets))

	got, err := ticketRepo.ListByBookingID(ctx, b.ID)
	require.NoError(t, err)
	assert.Len(t, got, 3)

	seats := make(map[string]bool)
	for _, tk := range got {
		seats[tk.SeatNumber] = true
		assert.Equal(t, b.ID, tk.BookingID)
		assert.Equal(t, domain.TicketStatusReserved, tk.Status)
	}
	assert.True(t, seats["1A"] && seats["1B"] && seats["1C"])
}

func TestTicketRepo_ListByBookingID_Empty(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()

	bookingRepo := pgrepo.NewBookingRepo(testDB)
	ticketRepo := pgrepo.NewTicketRepo(testDB)

	b := newBooking(1, "TKT002", domain.BookingStatusPending, time.Now().Add(15*time.Minute))
	require.NoError(t, bookingRepo.Create(ctx, b, nil))

	got, err := ticketRepo.ListByBookingID(ctx, b.ID)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestTicketRepo_IssueTickets(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()

	bookingRepo := pgrepo.NewBookingRepo(testDB)
	ticketRepo := pgrepo.NewTicketRepo(testDB)

	b := newBooking(1, "TKT003", domain.BookingStatusPending, time.Now().Add(15*time.Minute))
	tickets := newTickets(b.ID, "5A", "5B")
	require.NoError(t, bookingRepo.Create(ctx, b, tickets))

	ticketNumbers := map[string]string{
		tickets[0].ID: "TN-000001",
		tickets[1].ID: "TN-000002",
	}
	require.NoError(t, ticketRepo.IssueTickets(ctx, b.ID, ticketNumbers))

	got, err := ticketRepo.ListByBookingID(ctx, b.ID)
	require.NoError(t, err)
	require.Len(t, got, 2)

	for _, tk := range got {
		assert.Equal(t, domain.TicketStatusIssued, tk.Status)
		assert.NotEmpty(t, tk.TicketNumber)
		assert.Equal(t, ticketNumbers[tk.ID], tk.TicketNumber)
	}
}

func TestTicketRepo_IssueTickets_EmptyMap(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()
	ticketRepo := pgrepo.NewTicketRepo(testDB)

	// should be a no-op, not an error
	err := ticketRepo.IssueTickets(ctx, "nonexistent-booking", map[string]string{})
	assert.NoError(t, err)
}

func TestTicketRepo_IssueTickets_WrongBookingID(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()

	bookingRepo := pgrepo.NewBookingRepo(testDB)
	ticketRepo := pgrepo.NewTicketRepo(testDB)

	b := newBooking(1, "TKT004", domain.BookingStatusPending, time.Now().Add(15*time.Minute))
	tickets := newTickets(b.ID, "7A")
	require.NoError(t, bookingRepo.Create(ctx, b, tickets))

	// wrong booking ID — rows affected = 0 → error
	err := ticketRepo.IssueTickets(ctx, "00000000-0000-0000-0000-000000000000", map[string]string{
		tickets[0].ID: "TN-WRONG",
	})
	assert.Error(t, err)
}
