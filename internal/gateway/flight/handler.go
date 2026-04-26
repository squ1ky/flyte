package flight

import (
	"fmt"
	"github.com/gin-gonic/gin"
	flightv1 "github.com/squ1ky/flyte/gen/proto/flight"
	"github.com/squ1ky/flyte/internal/gateway/common"
	"net/http"
)

const (
	paramFlightID    = "id"
	paramAircraftID  = "id"
	paramAirportCode = "code"
	paramAirlineID   = "id"
)

const (
	msgInvalidDate        = "invalid date format, expected YYYY-MM-DD"
	msgInvalidDateRFC3339 = "invalid %s format, expected RFC 3339"
	msgMissingAirportCode = "missing airport code"
)

type Handler struct {
	client flightv1.FlightServiceClient
}

func NewHandler(client flightv1.FlightServiceClient) *Handler {
	return &Handler{client: client}
}

// SearchFlights godoc
//
//	@Summary		Search flights
//	@Description	Searches for available flights by route, date, and optional seat class.
//	@Tags			flights
//	@Produce		json
//	@Param			from		query	string	true	"Departure airport IATA code (e.g. DME)"
//	@Param			to			query	string	true	"Arrival airport IATA code (e.g. LED)"
//	@Param			date		query	string	true	"Departure date in YYYY-MM-DD format"
//	@Param			passengers	query	int		false	"Number of passengers"	default(1)
//	@Param			seat_class	query	string	false	"Seat class filter"		Enums(economy, business, first)
//	@Success		200			{array}		FlightResponse					"List of matching flights"
//	@Failure		400			{object}	common.ErrorResponse			"Invalid query parameters"
//	@Failure		500			{object}	common.ErrorResponse			"Internal server error"
//	@Router			/flights [get]
func (h *Handler) SearchFlights(c *gin.Context) {
	var q SearchFlightsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		common.AbortInvalidInput(c, err)
		return
	}

	depDate, err := timestampFromDateStr(q.Date)
	if err != nil {
		common.AbortWithError(c, http.StatusBadRequest, msgInvalidDate)
	}

	resp, err := h.client.SearchFlights(c.Request.Context(), &flightv1.SearchFlightsRequest{
		FromAirportCode: q.From,
		ToAirportCode:   q.To,
		DepartureDate:   depDate,
		PassengerCount:  q.Passengers,
		SeatClass:       seatClassToProto(q.SeatClass),
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	flights := make([]FlightResponse, len(resp.Flights))
	for i, flight := range resp.Flights {
		flights[i] = mapProtoToFlight(flight)
	}

	c.JSON(http.StatusOK, flights)
}

// SearchAirports godoc
//
//	@Summary		Search airports
//	@Description	Searches airports by name, city, or IATA code.
//	@Tags			airports
//	@Produce		json
//	@Param			query	query	string	true	"Search query (name, city, or IATA code)"
//	@Success		200		{array}		AirportResponse				"List of matching airports"
//	@Failure		400		{object}	common.ErrorResponse		"Invalid query"
//	@Failure		500		{object}	common.ErrorResponse		"Internal server error"
//	@Router			/airports/search [get]
func (h *Handler) SearchAirports(c *gin.Context) {
	var q SearchAirportsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		common.AbortInvalidInput(c, err)
		return
	}

	resp, err := h.client.SearchAirports(c.Request.Context(), &flightv1.SearchAirportsRequest{
		Query: q.Query,
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	airports := make([]AirportResponse, len(resp.Airports))
	for i, airport := range resp.Airports {
		airports[i] = mapProtoToAirport(airport)
	}

	c.JSON(http.StatusOK, airports)
}

// GetFlightDetails godoc
//
//	@Summary		Get flight details
//	@Description	Returns full details for a specific flight.
//	@Tags			flights
//	@Produce		json
//	@Param			id	path		int	true	"Flight ID"
//	@Success		200	{object}	FlightResponse				"Flight details"
//	@Failure		404	{object}	common.ErrorResponse		"Flight not found"
//	@Failure		500	{object}	common.ErrorResponse		"Internal server error"
//	@Router			/flights/{id} [get]
func (h *Handler) GetFlightDetails(c *gin.Context) {
	id, ok := common.ParseID(c, paramFlightID)
	if !ok {
		return
	}

	resp, err := h.client.GetFlightDetails(c.Request.Context(), &flightv1.GetFlightDetailsRequest{
		FlightId: id,
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapProtoToFlight(resp.Flight))
}

// GetFlightFares godoc
//
//	@Summary		Get flight fares
//	@Description	Returns fare rules (price per seat class) for a specific flight.
//	@Tags			flights
//	@Produce		json
//	@Param			id	path		int	true	"Flight ID"
//	@Success		200	{array}		FareRuleResponse			"List of fare rules"
//	@Failure		404	{object}	common.ErrorResponse		"Flight not found"
//	@Failure		500	{object}	common.ErrorResponse		"Internal server error"
//	@Router			/flights/{id}/fares [get]
func (h *Handler) GetFlightFares(c *gin.Context) {
	id, ok := common.ParseID(c, paramFlightID)
	if !ok {
		return
	}

	resp, err := h.client.GetFlightFares(c.Request.Context(), &flightv1.GetFlightFaresRequest{
		FlightId: id,
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	fares := make([]FareRuleResponse, len(resp.Fares))
	for i, fare := range resp.Fares {
		fares[i] = mapProtoToFareRule(fare)
	}

	c.JSON(http.StatusOK, fares)
}

// UpdateFlightStatus godoc
//
//	@Summary		Update flight status
//	@Description	Updates the status of a flight. Admin only.
//	@Tags			flights
//	@Accept			json
//	@Param			id		path	int							true	"Flight ID"
//	@Param			input	body	UpdateFlightStatusRequest	true	"New status"
//	@Success		204		"Status updated"
//	@Failure		400		{object}	common.ErrorResponse	"Invalid input"
//	@Failure		404		{object}	common.ErrorResponse	"Flight not found"
//	@Failure		500		{object}	common.ErrorResponse	"Internal server error"
//	@Security		BearerAuth
//	@Router			/flights/{id}/status [patch]
func (h *Handler) UpdateFlightStatus(c *gin.Context) {
	id, ok := common.ParseID(c, paramFlightID)
	if !ok {
		return
	}

	var input UpdateFlightStatusRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		common.AbortInvalidInput(c, err)
		return
	}

	_, err := h.client.UpdateFlightStatus(c.Request.Context(), &flightv1.UpdateFlightStatusRequest{
		FlightId: id,
		Status:   flightStatusToProto(input.Status),
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetReservedSeats godoc
//
//	@Summary		Get reserved seats
//	@Description	Returns all seat reservations for a specific flight. Admin only.
//	@Tags			reservations
//	@Produce		json
//	@Param			id	path		int	true	"Flight ID"
//	@Success		200	{array}		SeatReservationResponse		"List of reservations"
//	@Failure		404	{object}	common.ErrorResponse		"Flight not found"
//	@Failure		500	{object}	common.ErrorResponse		"Internal server error"
//	@Security		BearerAuth
//	@Router			/flights/{id}/seats/reserved [get]
func (h *Handler) GetReservedSeats(c *gin.Context) {
	flightID, ok := common.ParseID(c, paramFlightID)
	if !ok {
		return
	}

	resp, err := h.client.GetReservedSeats(c.Request.Context(), &flightv1.GetReservedSeatsRequest{
		FlightId: flightID,
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	seats := make([]SeatReservationResponse, len(resp.Seats))
	for i, seat := range resp.Seats {
		seats[i] = mapProtoToSeatReservation(seat)
	}

	c.JSON(http.StatusOK, seats)
}

// GetAirport godoc
//
//	@Summary		Get airport by code
//	@Description	Returns airport information by IATA code.
//	@Tags			airports
//	@Produce		json
//	@Param			code	path		string	true	"IATA airport code"
//	@Success		200		{object}	AirportResponse				"Airport info"
//	@Failure		404		{object}	common.ErrorResponse		"Airport not found"
//	@Failure		500		{object}	common.ErrorResponse		"Internal server error"
//	@Router			/airports/{code} [get]
func (h *Handler) GetAirport(c *gin.Context) {
	code := c.Param(paramAirportCode)
	if code == "" {
		common.AbortWithError(c, http.StatusBadRequest, msgMissingAirportCode)
		return
	}

	resp, err := h.client.GetAirport(c.Request.Context(), &flightv1.GetAirportRequest{
		Code: code,
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapProtoToAirport(resp.Airport))
}

// GetAirline godoc
//
//	@Summary		Get airline by ID
//	@Description	Returns airline information by ID.
//	@Tags			airlines
//	@Produce		json
//	@Param			id	path		int	true	"Airline ID"
//	@Success		200	{object}	AirlineResponse				"Airline info"
//	@Failure		404	{object}	common.ErrorResponse		"Airline not found"
//	@Failure		500	{object}	common.ErrorResponse		"Internal server error"
//	@Router			/airlines/{id} [get]
func (h *Handler) GetAirline(c *gin.Context) {
	id, ok := common.ParseID(c, paramAirlineID)
	if !ok {
		return
	}

	resp, err := h.client.GetAirline(c.Request.Context(), &flightv1.GetAirlineRequest{
		AirlineId: id,
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapProtoToAirline(resp.Airline))
}

// ListAirlines godoc
//
//	@Summary		List airlines
//	@Description	Returns a paginated list of airlines.
//	@Tags			airlines
//	@Produce		json
//	@Param			limit	query		int	false	"Max results"	default(20)
//	@Param			offset	query		int	false	"Offset"		default(0)
//	@Success		200		{array}		AirlineResponse				"List of airlines"
//	@Failure		500		{object}	common.ErrorResponse		"Internal server error"
//	@Router			/airlines [get]
func (h *Handler) ListAirlines(c *gin.Context) {
	const defaultLimit = 20
	const defaultOffset = 0

	limit := int32(common.QueryInt(c, "limit", defaultLimit))
	offset := int32(common.QueryInt(c, "offset", defaultOffset))

	resp, err := h.client.ListAirlines(c.Request.Context(), &flightv1.ListAirlinesRequest{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	airlines := make([]AirlineResponse, len(resp.Airlines))
	for i, airline := range resp.Airlines {
		airlines[i] = mapProtoToAirline(airline)
	}

	c.JSON(http.StatusOK, airlines)
}

// CreateAircraft godoc
//
//	@Summary		Create aircraft
//	@Description	Creates a new aircraft model. Admin only.
//	@Tags			aircrafts
//	@Accept			json
//	@Produce		json
//	@Param			input	body		CreateAircraftRequest	true	"Aircraft data"
//	@Success		201		{object}	IDResponse				"Aircraft created"
//	@Failure		400		{object}	common.ErrorResponse	"Invalid input"
//	@Failure		409		{object}	common.ErrorResponse	"Model already exists"
//	@Failure		500		{object}	common.ErrorResponse	"Internal server error"
//	@Security		BearerAuth
//	@Router			/aircrafts [post]
func (h *Handler) CreateAircraft(c *gin.Context) {
	var input CreateAircraftRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		common.AbortInvalidInput(c, err)
		return
	}

	resp, err := h.client.CreateAircraft(c.Request.Context(), &flightv1.CreateAircraftRequest{
		Model:      input.Model,
		TotalSeats: input.TotalSeats,
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusCreated, IDResponse{
		ID: int64(resp.AircraftId),
	})
}

// AddAircraftSeats godoc
//
//	@Summary		Add seats to aircraft
//	@Description	Adds seat definitions to an existing aircraft. Admin only.
//	@Tags			aircrafts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"Aircraft ID"
//	@Param			input	body		AddAircraftSeatsRequest	true	"Seat definitions"
//	@Success		200		{object}	AddAircraftSeatsResponse	"Seats added"
//	@Failure		400		{object}	common.ErrorResponse		"Invalid input"
//	@Failure		404		{object}	common.ErrorResponse		"Aircraft not found"
//	@Failure		500		{object}	common.ErrorResponse		"Internal server error"
//	@Security		BearerAuth
//	@Router			/aircrafts/{id}/seats [post]
func (h *Handler) AddAircraftSeats(c *gin.Context) {
	aircraftID, ok := common.ParseID(c, paramAircraftID)
	if !ok {
		return
	}

	var input AddAircraftSeatsRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		common.AbortInvalidInput(c, err)
		return
	}

	pbSeats := make([]*flightv1.SeatTemplate, len(input.Seats))
	for i, seat := range input.Seats {
		pbSeats[i] = &flightv1.SeatTemplate{
			SeatNumber: seat.SeatNumber,
			SeatClass:  seatClassToProto(seat.SeatClass),
			Row:        seat.Row,
			Col:        seat.Col,
		}
	}

	resp, err := h.client.AddAircraftSeats(c.Request.Context(), &flightv1.AddAircraftSeatsRequest{
		AircraftId: aircraftID,
		Seats:      pbSeats,
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, AddAircraftSeatsResponse{
		SeatsAdded: resp.SeatsAdded,
	})
}

// GetAircraftSeats godoc
//
//	@Summary		Get aircraft seat map
//	@Description	Returns the seat template for an aircraft.
//	@Tags			aircrafts
//	@Produce		json
//	@Param			id	path		int	true	"Aircraft ID"
//	@Success		200	{array}		SeatTemplateResponse		"Seat map"
//	@Failure		404	{object}	common.ErrorResponse		"Aircraft not found"
//	@Failure		500	{object}	common.ErrorResponse		"Internal server error"
//	@Router			/aircrafts/{id}/seats [get]
func (h *Handler) GetAircraftSeats(c *gin.Context) {
	aircraftID, ok := common.ParseID(c, paramAircraftID)
	if !ok {
		return
	}

	resp, err := h.client.GetAircraftSeats(c.Request.Context(), &flightv1.GetAircraftSeatsRequest{
		AircraftId: aircraftID,
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	seats := make([]SeatTemplateResponse, len(resp.Seats))
	for i, seat := range resp.Seats {
		seats[i] = mapProtoToSeatTemplate(seat)
	}

	c.JSON(http.StatusOK, seats)
}

// CreateAirline godoc
//
//	@Summary		Create airline
//	@Description	Creates a new airline. Admin only.
//	@Tags			airlines
//	@Accept			json
//	@Produce		json
//	@Param			input	body		CreateAirlineRequest	true	"Airline data"
//	@Success		201		{object}	IDResponse				"Airline created"
//	@Failure		400		{object}	common.ErrorResponse	"Invalid input"
//	@Failure		409		{object}	common.ErrorResponse	"IATA code already exists"
//	@Failure		500		{object}	common.ErrorResponse	"Internal server error"
//	@Security		BearerAuth
//	@Router			/airlines [post]
func (h *Handler) CreateAirline(c *gin.Context) {
	var input CreateAirlineRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		common.AbortInvalidInput(c, err)
		return
	}

	resp, err := h.client.CreateAirline(c.Request.Context(), &flightv1.CreateAirlineRequest{
		IataCode: input.IATACode,
		Name:     input.Name,
		Country:  input.Country,
		LogoUrl:  input.LogoURL,
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusCreated, IDResponse{
		ID: resp.AirlineId,
	})
}

// CreateFlight godoc
//
//	@Summary		Create flight
//	@Description	Creates a new flight with fare rules. Admin only.
//	@Tags			flights
//	@Accept			json
//	@Produce		json
//	@Param			input	body		CreateFlightRequest		true	"Flight data"
//	@Success		201		{object}	IDResponse				"Flight created"
//	@Failure		400		{object}	common.ErrorResponse	"Invalid input"
//	@Failure		500		{object}	common.ErrorResponse	"Internal server error"
//	@Security		BearerAuth
//	@Router			/flights [post]
func (h *Handler) CreateFlight(c *gin.Context) {
	var input CreateFlightRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		common.AbortInvalidInput(c, err)
		return
	}

	depTime, err := timestampFromRFC3339(input.DepartureTime)
	if err != nil {
		common.AbortWithError(c, http.StatusBadRequest, fmt.Sprintf(msgInvalidDateRFC3339, "departure_time"))
		return
	}

	arrTime, err := timestampFromRFC3339(input.ArrivalTime)
	if err != nil {
		common.AbortWithError(c, http.StatusBadRequest, fmt.Sprintf(msgInvalidDateRFC3339, "arrival_time"))
		return
	}

	fareRules := make([]*flightv1.FareRule, len(input.FareRules))
	for i, fare := range input.FareRules {
		fareRules[i] = &flightv1.FareRule{
			SeatClass:  seatClassToProto(fare.SeatClass),
			PriceCents: fare.PriceCents,
		}
	}

	resp, err := h.client.CreateFlight(c.Request.Context(), &flightv1.CreateFlightRequest{
		FlightNumber:         input.FlightNumber,
		AircraftId:           input.AircraftID,
		AirlineId:            input.AirlineID,
		DepartureAirportCode: input.DepartureAirportCode,
		ArrivalAirportCode:   input.ArrivalAirportCode,
		DepartureTime:        depTime,
		ArrivalTime:          arrTime,
		FareRules:            fareRules,
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusCreated, IDResponse{
		ID: resp.FlightId,
	})
}
