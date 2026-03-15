package domain

import (
	"time"
)

type FlightStatus string

const (
	FlightStatusScheduled FlightStatus = "SCHEDULED"
	FlightStatusDeparted  FlightStatus = "DEPARTED"
	FlightStatusArrived   FlightStatus = "ARRIVED"
	FlightStatusCancelled FlightStatus = "CANCELLED"
)

type SeatClass string

const (
	SeatClassEconomy  SeatClass = "ECONOMY"
	SeatClassBusiness SeatClass = "BUSINESS"
	SeatClassFirst    SeatClass = "FIRST"
)

type ReservationStatus string

const (
	ReservationStatusLocked    ReservationStatus = "LOCKED"
	ReservationStatusConfirmed ReservationStatus = "CONFIRMED"
)

type Airport struct {
	Code     string `db:"code" json:"code"`
	Name     string `db:"name" json:"name"`
	City     string `db:"city" json:"city"`
	Country  string `db:"country" json:"country"`
	Timezone string `db:"timezone" json:"timezone"`
}

type Airline struct {
	ID       int64  `db:"id" json:"id"`
	IATACode string `db:"iata_code" json:"iata_code"`
	Name     string `db:"name" json:"name"`
	Country  string `db:"country" json:"country"`
	LogoURL  string `db:"logo_url" json:"logo_url"`
}

func (a Airline) Validate() error {
	if a.IATACode == "" {
		return ErrInvalidIATACode
	}
	if a.Name == "" {
		return ErrInvalidAirlineName
	}
	return nil
}

type Aircraft struct {
	ID         int64  `db:"id" json:"id"`
	Model      string `db:"model" json:"model"`
	TotalSeats int    `db:"total_seats" json:"total_seats"`
}

func (a Aircraft) Validate() error {
	if a.Model == "" {
		return ErrInvalidAircraftModel
	}
	if a.TotalSeats <= 0 {
		return ErrInvalidTotalSeats
	}
	return nil
}

type AircraftSeat struct {
	ID         int64     `db:"id" json:"id"`
	AircraftID int64     `db:"aircraft_id" json:"aircraft_id"`
	SeatNumber string    `db:"seat_number" json:"seat_number"`
	SeatClass  SeatClass `db:"seat_class" json:"seat_class"`
	Row        int       `db:"row_number" json:"row"`
	Col        int       `db:"col_number" json:"col"`
}

type Flight struct {
	ID                   int64        `db:"id" json:"id"`
	FlightNumber         string       `db:"flight_number" json:"flight_number"`
	AirlineID            int64        `db:"airline_id" json:"airline_id"`
	AircraftID           int64        `db:"aircraft_id" json:"aircraft_id"`
	DepartureAirportCode string       `db:"departure_airport" json:"departure_airport_code"`
	ArrivalAirportCode   string       `db:"arrival_airport" json:"arrival_airport_code"`
	DepartureTime        time.Time    `db:"departure_time" json:"departure_time"`
	ArrivalTime          time.Time    `db:"arrival_time" json:"arrival_time"`
	Status               FlightStatus `db:"status" json:"status"`
	CreatedAt            time.Time    `db:"created_at" json:"created_at"`
}

func (f Flight) Validate() error {
	if f.FlightNumber == "" {
		return ErrEmptyFlightNumber
	}
	if f.AirlineID <= 0 {
		return ErrInvalidAirlineID
	}
	if f.AircraftID <= 0 {
		return ErrInvalidAircraftID
	}
	if f.DepartureAirportCode == "" || f.ArrivalAirportCode == "" {
		return ErrInvalidAirportCode
	}
	if f.DepartureAirportCode == f.ArrivalAirportCode {
		return ErrSameAirports
	}
	if !f.ArrivalTime.After(f.DepartureTime) {
		return ErrInvalidFlightTime
	}
	return nil
}

type FlightFare struct {
	ID          int64     `db:"id" json:"id"`
	FlightID    int64     `db:"flight_id" json:"flight_id"`
	SeatClass   SeatClass `db:"seat_class" json:"seat_class"`
	PriceCents  int64     `db:"price_cents" json:"price_cents"`
	SeatsTotal  int       `db:"seats_total" json:"seats_total"`
	SeatsBooked int       `db:"seats_booked" json:"seats_booked"`
}

type SeatReservation struct {
	ID         int64             `db:"id" json:"id"`
	FlightID   int64             `db:"flight_id" json:"flight_id"`
	SeatNumber string            `db:"seat_number" json:"seat_number"`
	BookingID  string            `db:"booking_id" json:"booking_id"`
	Status     ReservationStatus `db:"status" json:"status"`
	ExpiresAt  *time.Time        `db:"expires_at" json:"expires_at"`
	CreatedAt  time.Time         `db:"created_at" json:"created_at"`
}
