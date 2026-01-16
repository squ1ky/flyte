package domain

import "time"

type Airport struct {
	Code     string `db:"code" json:"code"`
	Name     string `db:"name" json:"name"`
	City     string `db:"city" json:"city"`
	Country  string `db:"country" json:"country"`
	Timezone string `db:"timezone" json:"timezone"`
}

type FlightStatus string

const (
	FlightStatusScheduled FlightStatus = "scheduled"
	FlightStatusCancelled FlightStatus = "cancelled"
	FlightStatusDelayed   FlightStatus = "delayed"
	FlightStatusArrived   FlightStatus = "arrived"
)

type Flight struct {
	ID               int64        `db:"id" json:"id"`
	FlightNumber     string       `db:"flight_number" json:"flight_number"`
	DepartureAirport string       `db:"departure_airport" json:"departure_airport"`
	ArrivalAirport   string       `db:"arrival_airport" json:"arrival_airport"`
	DepartureTime    time.Time    `db:"departure_time" json:"departure_time"`
	ArrivalTime      time.Time    `db:"arrival_time" json:"arrival_time"`
	PriceCents       int64        `db:"price_cents" json:"price_cents"`
	TotalSeats       int          `db:"total_seats" json:"total_seats"`
	Status           FlightStatus `db:"status" json:"status"`
	CreatedAt        time.Time    `db:"created_at" json:"created_at"`
	AvailableSeats   int          `db:"available_seats" json:"available_seats"`
}

type Seat struct {
	ID              int64      `db:"id" json:"id"`
	FlightID        int64      `db:"flight_id" json:"flight_id"`
	SeatNumber      string     `db:"seat_number" json:"seat_number"`
	IsBooked        bool       `db:"is_booked" json:"is_booked"`
	PriceMultiplier float64    `db:"price_multiplier" json:"price_multiplier"`
	ReservedAt      *time.Time `db:"reserved_at" json:"reserved_at,omitempty"`
}
