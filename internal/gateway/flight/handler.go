package flight

import (
	"github.com/gin-gonic/gin"
	flightv1 "github.com/squ1ky/flyte/gen/go/flight"
	"github.com/squ1ky/flyte/pkg/api"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net/http"
	"strconv"
	"time"
)

type Handler struct {
	client flightv1.FlightServiceClient
}

func NewHandler(client flightv1.FlightServiceClient) *Handler {
	return &Handler{client: client}
}

func (h *Handler) SearchFlights(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	dateStr := c.Query("date")
	passengersStr := c.Query("passengers")

	if from == "" || to == "" || dateStr == "" {
		api.NewErrorResponse(c, http.StatusBadRequest, "from, to and date are required")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		api.NewErrorResponse(c, http.StatusBadRequest, "invalid date format (expected YYYY-MM-DD)")
		return
	}

	passengers, err := strconv.Atoi(passengersStr)
	if err != nil || passengers < 1 {
		passengers = 1
	}

	req := &flightv1.SearchFlightsRequest{
		FromAirport:    from,
		ToAirport:      to,
		Date:           timestamppb.New(date),
		PassengerCount: int32(passengers),
	}

	resp, err := h.client.SearchFlights(c.Request.Context(), req)
	if err != nil {
		api.HandleGRPCErr(c, err)
		return
	}

	flights := make([]FlightResponse, len(resp.Flights))
	for i, f := range resp.Flights {
		flights[i] = mapProtoToFlight(f)
	}

	c.JSON(http.StatusOK, SearchResponse{Flights: flights})
}

func (h *Handler) CreateFlight(c *gin.Context) {
	var input CreateFlightRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		api.AbortInvalidInput(c, err)
		return
	}

	depTime, err := time.Parse(time.RFC3339, input.DepartureTime)
	if err != nil {
		api.NewErrorResponse(c, http.StatusBadRequest, "invalid departure_time format")
		return
	}

	arrTime, err := time.Parse(time.RFC3339, input.ArrivalTime)
	if err != nil {
		api.NewErrorResponse(c, http.StatusBadRequest, "invalid arrival_time format")
		return
	}

	req := &flightv1.CreateFlightRequest{
		FlightNumber:     input.FlightNumber,
		AircraftId:       input.AircraftID,
		DepartureAirport: input.DepartureAirport,
		ArrivalAirport:   input.ArrivalAirport,
		DepartureTime:    timestamppb.New(depTime),
		ArrivalTime:      timestamppb.New(arrTime),
		BasePriceCents:   input.BasePriceCents,
	}

	resp, err := h.client.CreateFlight(c.Request.Context(), req)
	if err != nil {
		api.HandleGRPCErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, IDResponse{
		ID: resp.FlightId,
	})
}

func (h *Handler) GetFlightDetails(c *gin.Context) {
	id, ok := api.ParseID(c, "id")
	if !ok {
		return
	}

	resp, err := h.client.GetFlightDetails(c.Request.Context(), &flightv1.GetFlightDetailsRequest{
		FlightId: id,
	})
	if err != nil {
		api.HandleGRPCErr(c, err)
		return
	}

	c.JSON(http.StatusOK, mapProtoToFlight(resp.Flight))
}

func (h *Handler) GetFlightSeats(c *gin.Context) {
	id, ok := api.ParseID(c, "id")
	if !ok {
		return
	}

	resp, err := h.client.GetFlightSeats(c.Request.Context(), &flightv1.GetFlightSeatsRequest{
		FlightId: id,
	})
	if err != nil {
		api.HandleGRPCErr(c, err)
		return
	}

	seats := make([]SeatResponse, len(resp.Seats))
	for i, s := range resp.Seats {
		seats[i] = SeatResponse{
			ID:              s.Id,
			SeatNumber:      s.SeatNumber,
			IsBooked:        s.IsBooked,
			PriceMultiplier: s.PriceMultiplier,
		}
	}

	c.JSON(http.StatusOK, seats)
}

func (h *Handler) ListAirports(c *gin.Context) {
	resp, err := h.client.ListAirports(c.Request.Context(), &flightv1.ListAirportsRequest{})
	if err != nil {
		api.HandleGRPCErr(c, err)
		return
	}

	airports := make([]AirportResponse, len(resp.Airports))
	for i, a := range resp.Airports {
		airports[i] = AirportResponse{
			Code:    a.Code,
			Name:    a.Name,
			City:    a.City,
			Country: a.Country,
		}
	}

	c.JSON(http.StatusOK, airports)
}

func (h *Handler) CreateAircraft(c *gin.Context) {
	var input CreateAircraftRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		api.AbortInvalidInput(c, err)
		return
	}

	resp, err := h.client.CreateAircraft(c.Request.Context(), &flightv1.CreateAircraftRequest{
		Model:      input.Model,
		TotalSeats: input.TotalSeats,
	})
	if err != nil {
		api.HandleGRPCErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, IDResponse{
		ID: resp.AircraftId,
	})
}

func (h *Handler) AddAircraftSeats(c *gin.Context) {
	id, ok := api.ParseID(c, "id")
	if !ok {
		return
	}

	var input AddSeatsRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		api.AbortInvalidInput(c, err)
		return
	}

	pbSeats := make([]*flightv1.AircraftSeatTemplate, 0, len(input.Seats))
	for _, s := range input.Seats {
		mult := s.PriceMultiplier
		if mult <= 0 {
			mult = 1.0
		}
		pbSeats = append(pbSeats, &flightv1.AircraftSeatTemplate{
			SeatNumber:      s.SeatNumber,
			SeatClass:       s.SeatClass,
			PriceMultiplier: mult,
		})
	}

	_, err := h.client.AddAircraftSeats(c.Request.Context(), &flightv1.AddAircraftSeatsRequest{
		AircraftId: id,
		Seats:      pbSeats,
	})
	if err != nil {
		api.HandleGRPCErr(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) ListAircrafts(c *gin.Context) {
	resp, err := h.client.ListAircrafts(c.Request.Context(), &flightv1.ListAircraftsRequest{})
	if err != nil {
		api.HandleGRPCErr(c, err)
		return
	}

	result := make([]AircraftResponse, len(resp.Aircrafts))
	for i, a := range resp.Aircrafts {
		result[i] = AircraftResponse{
			ID:         a.Id,
			Model:      a.Model,
			TotalSeats: a.TotalSeats,
		}
	}

	c.JSON(http.StatusOK, result)
}

func mapProtoToFlight(f *flightv1.Flight) FlightResponse {
	return FlightResponse{
		ID:               f.Id,
		FlightNumber:     f.FlightNumber,
		DepartureAirport: f.DepartureAirport,
		ArrivalAirport:   f.ArrivalAirport,
		DepartureTime:    f.DepartureTime.AsTime().Format(time.RFC3339),
		ArrivalTime:      f.ArrivalTime.AsTime().Format(time.RFC3339),
		BasePriceCents:   f.BasePriceCents,
		Status:           f.Status,
		TotalSeats:       f.TotalSeats,
		AvailableSeats:   f.AvailableSeats,
	}
}
