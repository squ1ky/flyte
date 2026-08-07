package domain

import "time"

type BookingCreatedEvent struct {
	BookingID       string    `json:"booking_id"`
	UserID          int64     `json:"user_id"`
	FlightID        int64     `json:"flight_id"`
	TotalPriceCents int64     `json:"total_price_cents"`
	Currency        string    `json:"currency"`
	CreatedAt       time.Time `json:"created_at"`
}

type BookingPaidEvent struct {
	BookingID string    `json:"booking_id"`
	PaidAt    time.Time `json:"paid_at"`
}

type BookingCancelledEvent struct {
	BookingID   string       `json:"booking_id"`
	Reason      CancelReason `json:"reason"`
	CancelledAt time.Time    `json:"cancelled_at"`
}
