package flight

type CreateFlightRequest struct {
	FlightNumber     string `json:"flight_number" binding:"required"`
	AircraftID       int64  `json:"aircraft_id" binding:"required"`
	DepartureAirport string `json:"departure_airport" binding:"required,len=3"`
	ArrivalAirport   string `json:"arrival_airport" binding:"required,len=3"`
	DepartureTime    string `json:"departure_time" binding:"required"`
	ArrivalTime      string `json:"arrival_time" binding:"required"`
	BasePriceCents   int64  `json:"base_price_cents" binding:"required,gt=0"`
}

type CreateAircraftRequest struct {
	Model      string `json:"model" binding:"required"`
	TotalSeats int32  `json:"total_seats" binding:"required,gt=0"`
}

type SeatTemplateInput struct {
	SeatNumber      string  `json:"seat_number" binding:"required"`
	SeatClass       string  `json:"seat_class" binding:"required"`
	PriceMultiplier float64 `json:"price_multiplier"`
}

type AddSeatsRequest struct {
	Seats []SeatTemplateInput `json:"seats" binding:"required,min=1"`
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
