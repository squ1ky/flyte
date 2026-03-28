package domain

import "time"

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "PENDING"
	BookingStatusPaid      BookingStatus = "PAID"
	BookingStatusCancelled BookingStatus = "CANCELLED"
)

type Booking struct {
	ID              string        `db:"id"`
	UserID          int64         `db:"user_id"`
	FlightID        int64         `db:"flight_id"`
	RefCode         string        `db:"ref_code"`
	Status          BookingStatus `db:"status"`
	TotalPriceCents int64         `db:"total_price_cents"`
	Currency        string        `db:"currency"`
	ContactEmail    string        `db:"contact_email"`
	ExpiresAt       time.Time     `db:"expires_at"`
	CreatedAt       time.Time     `db:"created_at"`

	Tickets []Ticket `db:"-"`
}

func (s BookingStatus) IsTerminal() bool {
	return s == BookingStatusPaid || s == BookingStatusCancelled
}
