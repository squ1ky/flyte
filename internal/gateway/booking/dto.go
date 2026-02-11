package booking

type CreateRequest struct {
	FlightId          int64  `json:"flight_id"`
	SeatNumber        string `json:"seat_number"`
	PassengerName     string `json:"passenger_name"`
	PassengerPassport string `json:"passenger_passport"`
	PriceCents        int64  `json:"price_cents"`
	Currency          string `json:"currency"`
}

type CreateResponse struct {
	BookingID string `json:"booking_id"`
}

type Response struct {
	ID                string `json:"id"`
	FlightID          int64  `json:"flight_id"`
	SeatNumber        string `json:"seat_number"`
	PassengerName     string `json:"passenger_name"`
	PassengerPassport string `json:"passenger_passport"`
	Status            string `json:"status"`
	PriceCents        int64  `json:"price_cents"`
	Currency          string `json:"currency"`
	CreatedAt         string `json:"created_at"`
}

type CancelResponse struct {
	Message string `json:"message"`
}
