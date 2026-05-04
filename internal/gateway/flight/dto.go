package flight

// SearchFlightsQuery represents query parameters for the flight search endpoint.
// When from, to, and date are all provided the endpoint performs an ES search.
// Otherwise it returns a paginated list of upcoming flights from the DB.
type SearchFlightsQuery struct {
	// Departure airport IATA code
	From string `form:"from" binding:"omitempty,len=3" example:"DME"`
	// Arrival airport IATA code
	To string `form:"to" binding:"omitempty,len=3" example:"LED"`
	// Departure date in YYYY-MM-DD format
	Date string `form:"date" binding:"omitempty" example:"2025-07-15"`
	// Number of passengers (default: 1)
	Passengers int32 `form:"passengers,default=1" binding:"min=1" example:"1"`
	// Seat class filter: economy, business, first (optional)
	SeatClass string `form:"seat_class" binding:"omitempty,oneof=economy business first" example:"economy"`
	// Max results for list mode (default: 20)
	Limit int32 `form:"limit,default=20" binding:"min=1,max=100" example:"20"`
	// Offset for list mode (default: 0)
	Offset int32 `form:"offset,default=0" binding:"min=0" example:"0"`
}

// SearchAirportsQuery represents query parameters for airport search.
type SearchAirportsQuery struct {
	// Search query (airport name, city, or IATA code)
	Query string `form:"query" binding:"required,min=1" example:"Moscow"`
}

// UpdateFlightStatusRequest is the JSON body for updating a flight's status.
type UpdateFlightStatusRequest struct {
	// New status: scheduled, departed, arrived, cancelled
	Status string `json:"status" binding:"required,oneof=scheduled departed arrived cancelled" example:"departed"`
}

// CreateAircraftRequest is the JSON body for creating an aircraft.
type CreateAircraftRequest struct {
	// Aircraft model name
	Model string `json:"model" binding:"required" example:"Boeing 737-800"`
	// Total number of seats
	TotalSeats int32 `json:"total_seats" binding:"required,min=1" example:"189"`
}

// SeatTemplateInput represents a single seat definition when adding seats to an aircraft.
type SeatTemplateInput struct {
	// Seat number (e.g. "12A")
	SeatNumber string `json:"seat_number" binding:"required" example:"12A"`
	// Seat class: economy, business, first
	SeatClass string `json:"seat_class" binding:"required,oneof=economy business first" example:"economy"`
	// Row number
	Row int32 `json:"row" binding:"required,min=1" example:"12"`
	// Column number
	Col int32 `json:"col" binding:"required,min=1" example:"1"`
}

// AddAircraftSeatsRequest is the JSON body for adding seats to an aircraft.
type AddAircraftSeatsRequest struct {
	// List of seat definitions
	Seats []SeatTemplateInput `json:"seats" binding:"required,min=1,dive"`
}

// CreateAirlineRequest is the JSON body for creating an airline.
type CreateAirlineRequest struct {
	// IATA two-letter airline code
	IATACode string `json:"iata_code" binding:"required,len=2" example:"SU"`
	// Airline name
	Name string `json:"name" binding:"required" example:"Aeroflot"`
	// Country of registration
	Country string `json:"country" binding:"required" example:"Russia"`
	// Logo URL (optional)
	LogoURL string `json:"logo_url" example:"https://example.com/logo.png"`
}

// FareRuleInput represents a fare rule per seat class for flight creation.
type FareRuleInput struct {
	// Seat class: economy, business, first
	SeatClass string `json:"seat_class" binding:"required,oneof=economy business first" example:"economy"`
	// Price in cents
	PriceCents int64 `json:"price_cents" binding:"required,min=0" example:"500000"`
}

// CreateFlightRequest is the JSON body for creating a flight.
type CreateFlightRequest struct {
	// Flight number (e.g. "SU1234")
	FlightNumber string `json:"flight_number" binding:"required" example:"SU1234"`
	// Aircraft ID
	AircraftID int64 `json:"aircraft_id" binding:"required" example:"1"`
	// Airline ID
	AirlineID int64 `json:"airline_id" binding:"required" example:"1"`
	// Departure airport IATA code
	DepartureAirportCode string `json:"departure_airport_code" binding:"required,len=3" example:"DME"`
	// Arrival airport IATA code
	ArrivalAirportCode string `json:"arrival_airport_code" binding:"required,len=3" example:"LED"`
	// Departure time in RFC 3339 format
	DepartureTime string `json:"departure_time" binding:"required" example:"2025-07-15T10:00:00Z"`
	// Arrival time in RFC 3339 format
	ArrivalTime string `json:"arrival_time" binding:"required" example:"2025-07-15T11:30:00Z"`
	// Fare rules per seat class
	FareRules []FareRuleInput `json:"fare_rules" binding:"required,min=1,dive"`
}

// ReserveSeatsRequest is the JSON body for reserving seats.
type ReserveSeatsRequest struct {
	// List of seat numbers to reserve
	SeatNumbers []string `json:"seat_numbers" binding:"required,min=1" example:"12A,12B"`
	// Booking ID (UUID)
	BookingID string `json:"booking_id" binding:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// ConfirmReservationRequest is the JSON body for confirming a reservation.
type ConfirmReservationRequest struct {
	// List of seat numbers to confirm
	SeatNumbers []string `json:"seat_numbers" binding:"required,min=1" example:"12A,12B"`
	// Booking ID (UUID)
	BookingID string `json:"booking_id" binding:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// CancelReservationRequest is the JSON body for cancelling a reservation.
type CancelReservationRequest struct {
	// Booking ID (UUID)
	BookingID string `json:"booking_id" binding:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// AirportResponse represents an airport in API responses.
type AirportResponse struct {
	// IATA airport code
	Code string `json:"code" example:"DME"`
	// Airport name
	Name string `json:"name" example:"Moscow Domodedovo Airport"`
	// City
	City string `json:"city" example:"Moscow"`
	// Country
	Country string `json:"country" example:"Russia"`
	// IANA timezone
	Timezone string `json:"timezone" example:"Europe/Moscow"`
}

// AirlineResponse represents an airline in API responses.
type AirlineResponse struct {
	// Airline ID
	ID int64 `json:"id" example:"1"`
	// IATA two-letter code
	IATACode string `json:"iata_code" example:"SU"`
	// Airline name
	Name string `json:"name" example:"Aeroflot"`
	// Country
	Country string `json:"country" example:"Russia"`
	// Logo URL
	LogoURL string `json:"logo_url" example:"https://example.com/logo.png"`
}

// AircraftResponse represents an aircraft in API responses.
type AircraftResponse struct {
	// Aircraft ID
	ID int64 `json:"id" example:"1"`
	// Aircraft model
	Model string `json:"model" example:"Boeing 737-800"`
	// Total number of seats
	TotalSeats int32 `json:"total_seats" example:"189"`
}

// FlightResponse represents a flight in API responses.
type FlightResponse struct {
	// Flight ID
	ID int64 `json:"id" example:"1"`
	// Flight number
	FlightNumber string `json:"flight_number" example:"SU1234"`
	// Airline info
	Airline AirlineResponse `json:"airline"`
	// Aircraft info
	Aircraft AircraftResponse `json:"aircraft"`
	// Departure airport
	DepartureAirport AirportResponse `json:"departure_airport"`
	// Arrival airport
	ArrivalAirport AirportResponse `json:"arrival_airport"`
	// Departure time in RFC 3339
	DepartureTime string `json:"departure_time" example:"2025-07-15T10:00:00Z"`
	// Arrival time in RFC 3339
	ArrivalTime string `json:"arrival_time" example:"2025-07-15T11:30:00Z"`
	// Flight status
	Status string `json:"status" example:"scheduled" enums:"scheduled,departed,arrived,cancelled"`
	// Minimum price in cents across all fare classes
	MinPriceCents int64 `json:"min_price_cents" example:"500000"`
	// Number of available seats
	AvailableSeats int32 `json:"available_seats" example:"45"`
	// Currency code
	Currency string `json:"currency" example:"RUB"`
}

// SeatResponse represents a flight seat in API responses.
type SeatResponse struct {
	// Seat ID
	ID int64 `json:"id" example:"1"`
	// Seat number
	SeatNumber string `json:"seat_number" example:"12A"`
	// Seat class
	SeatClass string `json:"seat_class" example:"economy" enums:"economy,business,first"`
	// Row number
	Row int32 `json:"row" example:"12"`
	// Column number
	Col int32 `json:"col" example:"1"`
	// Whether the seat is available for booking
	IsAvailable bool `json:"is_available" example:"true"`
	// Price in cents
	PriceCents int64 `json:"price_cents" example:"500000"`
}

// SeatTemplateResponse represents an aircraft seat template in API responses.
type SeatTemplateResponse struct {
	// Seat number
	SeatNumber string `json:"seat_number" example:"12A"`
	// Seat class
	SeatClass string `json:"seat_class" example:"economy" enums:"economy,business,first"`
	// Row number
	Row int32 `json:"row" example:"12"`
	// Column number
	Col int32 `json:"col" example:"1"`
}

// FareRuleResponse represents a fare rule in API responses.
type FareRuleResponse struct {
	// Seat class
	SeatClass string `json:"seat_class" example:"economy" enums:"economy,business,first"`
	// Price in cents
	PriceCents int64 `json:"price_cents" example:"500000"`
}

// ReserveSeatsResponse is returned after a seat reservation attempt.
type ReserveSeatsResponse struct {
	// Seats that could not be reserved (already taken)
	UnavailableSeats []string `json:"unavailable_seats"`
	// Reservation expiry time in RFC 3339
	ExpiresAt string `json:"expires_at" example:"2025-07-15T10:15:00Z"`
}

// SeatReservationResponse represents a single seat reservation.
type SeatReservationResponse struct {
	// Flight ID
	FlightID int64 `json:"flight_id" example:"1"`
	// Seat number
	SeatNumber string `json:"seat_number" example:"12A"`
	// Booking ID
	BookingID string `json:"booking_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Reservation expiry time
	ExpiresAt string `json:"expires_at" example:"2025-07-15T10:15:00Z"`
}

// IDResponse is a generic response containing a single created entity ID.
type IDResponse struct {
	// Created entity ID
	ID int64 `json:"id" example:"1"`
}

// AddAircraftSeatsResponse is returned after adding seats to an aircraft.
type AddAircraftSeatsResponse struct {
	// Number of seats added
	SeatsAdded int32 `json:"seats_added" example:"189"`
}
