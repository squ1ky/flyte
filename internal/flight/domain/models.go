package domain

import (
	"time"
)

type FlightStatus string

const (
	FlightStatusScheduled FlightStatus = "scheduled"
	FlightStatusCancelled FlightStatus = "cancelled"
	FlightStatusDelayed   FlightStatus = "delayed"
	FlightStatusArrived   FlightStatus = "arrived"
)

type SeatClass string

const (
	SeatClassEconomy  SeatClass = "economy"
	SeatClassComfort  SeatClass = "comfort"
	SeatClassBusiness SeatClass = "business"
)

type Airport struct {
	Code     string `db:"code" json:"code"`
	Name     string `db:"name" json:"name"`
	City     string `db:"city" json:"city"`
	Country  string `db:"country" json:"country"`
	Timezone string `db:"timezone" json:"timezone"`
}

type Aircraft struct {
	ID         int64  `db:"id" json:"id"`
	Model      string `db:"model" json:"model"`
	TotalSeats int    `db:"total_seats" json:"total_seats"`
}

type AircraftSeat struct {
	ID              int64     `db:"id" json:"id"`
	AircraftID      int64     `db:"aircraft_id" json:"aircraft_id"`
	SeatNumber      string    `db:"seat_number" json:"seat_number"`
	SeatClass       SeatClass `db:"seat_class" json:"seat_class"`
	PriceMultiplier float64   `db:"price_multiplier" json:"price_multiplier"`
}

type Flight struct {
	ID               int64        `db:"id" json:"id"`
	FlightNumber     string       `db:"flight_number" json:"flight_number"`
	AircraftID       int64        `db:"aircraft_id" json:"aircraft_id"`
	DepartureAirport string       `db:"departure_airport" json:"departure_airport"`
	ArrivalAirport   string       `db:"arrival_airport" json:"arrival_airport"`
	DepartureTime    time.Time    `db:"departure_time" json:"departure_time"`
	ArrivalTime      time.Time    `db:"arrival_time" json:"arrival_time"`
	BasePriceCents   int64        `db:"base_price_cents" json:"base_price_cents"`
	Status           FlightStatus `db:"status" json:"status"`
	CreatedAt        time.Time    `db:"created_at" json:"created_at"`

	AvailableSeats int    `db:"-" json:"available_seats"`
	Seats          []Seat `db:"-" json:"seats,omitempty"`
}

type Seat struct {
	ID              int64      `db:"id" json:"id"`
	FlightID        int64      `db:"flight_id" json:"flight_id"`
	SeatNumber      string     `db:"seat_number" json:"seat_number"`
	SeatClass       SeatClass  `db:"seat_class" json:"seat_class"`
	IsBooked        bool       `db:"is_booked" json:"is_booked"`
	PriceMultiplier float64    `db:"price_multiplier" json:"price_multiplier"`
	ReservedAt      *time.Time `db:"reserved_at" json:"reserved_at,omitempty"`
}
