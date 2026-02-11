package flight

type CreateFlightRequest struct {
	FlightNumber     string `json:"flight_number"`
	AircraftID       int64  `json:"aircraft_id"`
	DepartureAirport string `json:"departure_airport"`
	ArrivalAirport   string `json:"arrival_airport"`
	DepartureTime    string `json:"departure_time"`
	ArrivalTime      string `json:"arrival_time"`
	BasePriceCents   int64  `json:"base_price_cents"`
}

type CreateAircraftRequest struct {
	Model      string `json:"model"`
	TotalSeats int32  `json:"total_seats"`
}

type SeatTemplateInput struct {
	SeatNumber      string  `json:"seat_number"`
	SeatClass       string  `json:"seat_class"`
	PriceMultiplier float64 `json:"price_multiplier"`
}

type AddSeatsRequest struct {
	Seats []SeatTemplateInput `json:"seats"`
}

type FlightResponse struct {
	ID               int64  `json:"id"`
	FlightNumber     string `json:"flight_number"`
	DepartureAirport string `json:"departure_airport"`
	ArrivalAirport   string `json:"arrival_airport"`
	DepartureTime    string `json:"departure_time"`
	ArrivalTime      string `json:"arrival_time"`
	BasePriceCents   int64  `json:"base_price_cents"`
	Status           string `json:"status"`
	TotalSeats       int32  `json:"total_seats"`
	AvailableSeats   int32  `json:"available_seats"`
}

type AirportResponse struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	City    string `json:"city"`
	Country string `json:"country"`
}

type SeatResponse struct {
	ID              int64   `json:"id"`
	SeatNumber      string  `json:"seat_number"`
	IsBooked        bool    `json:"is_booked"`
	PriceMultiplier float64 `json:"price_multiplier"`
}

type AircraftResponse struct {
	ID         int64  `json:"id"`
	Model      string `json:"model"`
	TotalSeats int32  `json:"total_seats"`
}

type IDResponse struct {
	ID int64 `json:"id"`
}

type SearchResponse struct {
	Flights []FlightResponse `json:"flights"`
}
