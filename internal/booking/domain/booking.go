package domain

import "time"

type Booking struct {
	ID                string        `db:"id"`
	UserID            int64         `db:"user_id"`
	FlightID          int64         `db:"flight_id"`
	SeatNumber        string        `db:"seat_number"`
	PassengerName     string        `db:"passenger_name"`
	PassengerPassport string        `db:"passenger_passport"`
	PriceCents        int64         `db:"price_cents"`
	Currency          string        `db:"currency"`
	Status            BookingStatus `db:"status"`
	CreatedAt         time.Time     `db:"created_at"`
	UpdatedAt         time.Time     `db:"updated_at"`
}

type BookingStatus string

const (
	StatusPending   BookingStatus = "PENDING"
	StatusPaid      BookingStatus = "PAID"
	StatusCancelled BookingStatus = "CANCELLED"
	StatusFailed    BookingStatus = "FAILED"
)
