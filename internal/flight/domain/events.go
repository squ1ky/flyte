package domain

type EventType string

const (
	EventFlightCreated EventType = "FLIGHT_CREATED"
	EventSeatsChanged  EventType = "SEATS_CHANGED"
)
