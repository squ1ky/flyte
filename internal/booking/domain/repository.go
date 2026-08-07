package domain

import "context"

type BookingRepository interface {
	// Create saves the booking and tickets in a single transaction.
	Create(ctx context.Context, booking *Booking, tickets []Ticket) error

	// GetByID returns the booking without tickets.
	GetByID(ctx context.Context, id string) (*Booking, error)

	// GetByRefCode returns a booking without tickets.
	GetByRefCode(ctx context.Context, refCode string) (*Booking, error)

	// ListByUserID returns a user's bookings (without tickets).
	ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]Booking, error)

	// UpdateStatus changes the booking status. reason is optional and only
	// meaningful when status is BookingStatusCancelled.
	UpdateStatus(ctx context.Context, id string, status BookingStatus, reason *CancelReason) error

	// FindExpired returns Pending bookings with an expired expires_at
	FindExpired(ctx context.Context, limit int) ([]Booking, error)
}

type TicketRepository interface {
	// ListByBookingID returns all tickets for a booking.
	ListByBookingID(ctx context.Context, bookingID string) ([]Ticket, error)

	// IssueTickets sets the ticket_number and status to ISSUED.
	// ticketNumbers – map[ticket_id] -> ticket_number.
	IssueTickets(ctx context.Context, bookingID string, ticketNumbers map[string]string) error
}
