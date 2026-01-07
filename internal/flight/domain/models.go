package domain

import "time"

type Airport struct {
	Code     string `db:"code"`
	Name     string `db:"name"`
	City     string `db:"city"`
	Country  string `db:"country"`
	Timezone string `db:"timezone"`
}

type Flight struct {
	ID               int64     `db:"id"`
	FlightNumber     string    `db:"flight_number"`
	DepartureAirport string    `db:"departure_airport"`
	ArrivalAirport   string    `db:"arrival_airport"`
	DepartureTime    time.Time `db:"departure_time"`
	ArrivalTime      time.Time `db:"arrival_time"`
	Price            float64   `db:"price"`
	TotalSeats       int       `db:"total_seats"`
	Status           string    `db:"status"`
	CreatedAt        time.Time `db:"created_at"`

	AvailableSeats int `db:"available_seats"`
}

type Seat struct {
	ID              int64   `db:"id"`
	FlightID        int64   `db:"flight_id"`
	SeatNumber      string  `db:"seat_number"`
	IsBooked        bool    `db:"is_booked"`
	PriceMultiplier float64 `db:"price_multiplier"`
}
