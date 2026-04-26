package booking

// PassengerInput represents a single passenger in a booking request.
type PassengerInput struct {
	// Legal first name as in travel document
	FirstName string `json:"first_name" binding:"required" example:"Ivan"`
	// Legal last name as in travel document
	LastName string `json:"last_name" binding:"required" example:"Ivanov"`
	// Travel document number
	DocumentNumber string `json:"document_number" binding:"required" example:"7712345678"`
	// Selected seat number
	SeatNumber string `json:"seat_number" binding:"required" example:"12A"`
	// Seat class: economy, business, first
	SeatClass string `json:"seat_class" binding:"required,oneof=economy business first" example:"economy"`
}

// CreateBookingRequest is the JSON body for creating a booking.
type CreateBookingRequest struct {
	// Flight ID
	FlightID int64 `json:"flight_id" binding:"required" example:"1"`
	// Contact email for booking notifications
	ContactEmail string `json:"contact_email" binding:"required,email" example:"user@example.com"`
	// List of passengers
	Passengers []PassengerInput `json:"passengers" binding:"required,min=1,dive"`
}

// CreateBookingResponse is returned after a booking is successfully created.
type CreateBookingResponse struct {
	// Booking UUID
	BookingID string `json:"booking_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Short reference code for the passenger
	RefCode string `json:"ref_code" example:"A1B2C3"`
	// Reservation expiry time in RFC 3339
	ExpiresAt string `json:"expires_at" example:"2025-07-15T10:15:00Z"`
}

// TicketResponse represents a single ticket within a booking.
type TicketResponse struct {
	// Ticket UUID
	ID string `json:"id" example:"660e8400-e29b-41d4-a716-446655440000"`
	// Passenger first name
	PassengerFirstName string `json:"passenger_first_name" example:"Ivan"`
	// Passenger last name
	PassengerLastName string `json:"passenger_last_name" example:"Ivanov"`
	// Passenger document number
	PassengerDocNum string `json:"passenger_doc_num" example:"7712345678"`
	// Seat number
	SeatNumber string `json:"seat_number" example:"12A"`
	// Seat class
	SeatClass string `json:"seat_class" example:"economy" enums:"economy,business,first"`
	// Ticket price in cents
	PriceCents int64 `json:"price_cents" example:"500000"`
	// Ticket number (assigned after issuance)
	TicketNumber string `json:"ticket_number,omitempty" example:"SU1234567890"`
	// Ticket status
	Status string `json:"status" example:"reserved" enums:"reserved,issued"`
}

// BookingResponse represents a full booking in API responses.
type BookingResponse struct {
	// Booking UUID
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Owner user ID
	UserID int64 `json:"user_id" example:"1"`
	// Flight ID
	FlightID int64 `json:"flight_id" example:"1"`
	// Short reference code
	RefCode string `json:"ref_code" example:"A1B2C3"`
	// Booking status
	Status string `json:"status" example:"pending" enums:"pending,paid,cancelled"`
	// Total price in cents
	TotalPriceCents int64 `json:"total_price_cents" example:"1000000"`
	// Currency code
	Currency string `json:"currency" example:"RUB"`
	// Contact email
	ContactEmail string `json:"contact_email" example:"user@example.com"`
	// Tickets
	Tickets []TicketResponse `json:"tickets"`
	// Reservation expiry time in RFC 3339 (for pending bookings)
	ExpiresAt string `json:"expires_at,omitempty" example:"2025-07-15T10:15:00Z"`
	// Creation timestamp in RFC 3339
	CreatedAt string `json:"created_at" example:"2025-07-15T10:00:00Z"`
}
