package domain

type EventType string

const (
	EventTypeFlightCreated  EventType = "flight.created"
	EventTypeSeatsReserved  EventType = "flight.seats_reserved"
	EventTypeSeatsConfirmed EventType = "flight.seats_confirmed"
	EventTypeSeatsReleased  EventType = "flight.seats_released"
)

type FlightCreatedPayload struct {
	FlightID int64 `json:"flight_id"`
}

type SeatsStatusPayload struct {
	FlightID    int64    `json:"flight_id"`
	BookingID   string   `json:"booking_id"`
	SeatNumbers []string `json:"seat_numbers"`
}
