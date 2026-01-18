package handler

import (
	"github.com/gin-gonic/gin"
	flightv1 "github.com/squ1ky/flyte/gen/go/flight"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net/http"
	"strconv"
	"time"
)

type FlightHandler struct {
	client flightv1.FlightServiceClient
}

func NewFlightHandler(client flightv1.FlightServiceClient) *FlightHandler {
	return &FlightHandler{client: client}
}

func (h *FlightHandler) SearchFlights(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	dateStr := c.Query("date")
	passengersStr := c.Query("passengers")

	if from == "" || to == "" || dateStr == "" {
		newErrorResponse(c, http.StatusBadRequest, "from, to and date are required")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid date format (expected YYYY-MM-DD)")
		return
	}

	passengers, err := strconv.Atoi(passengersStr)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid passengers")
	}

	req := &flightv1.SearchFlightsRequest{
		FromAirport:    from,
		ToAirport:      to,
		Date:           timestamppb.New(date),
		PassengerCount: int32(passengers),
	}

	resp, err := h.client.SearchFlights(c.Request.Context(), req)
	if err != nil {
		mapGRPCErr(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

type createFlightInput struct {
	FlightNumber     string  `json:"flight_number" binding:"required"`
	AircraftID       int64   `json:"aircraft_id" binding:"required"`
	DepartureAirport string  `json:"departure_airport" binding:"required,len=3"`
	ArrivalAirport   string  `json:"arrival_airport" binding:"required,len=3"`
	DepartureTime    string  `json:"departure_time" binding:"required"`
	ArrivalTime      string  `json:"arrival_time" binding:"required"`
	BasePrice        float64 `json:"price" binding:"required,gt=0"`
}

func (h *FlightHandler) CreateFlight(c *gin.Context) {
	var input createFlightInput
	if err := c.ShouldBindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	depTime, err := time.Parse(time.RFC3339, input.DepartureTime)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid departure_time format")
		return
	}

	arrTime, err := time.Parse(time.RFC3339, input.ArrivalTime)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid arrival_time format")
		return
	}

	req := &flightv1.CreateFlightRequest{
		FlightNumber:     input.FlightNumber,
		AircraftId:       input.AircraftID,
		DepartureAirport: input.DepartureAirport,
		ArrivalAirport:   input.ArrivalAirport,
		DepartureTime:    timestamppb.New(depTime),
		ArrivalTime:      timestamppb.New(arrTime),
		BasePriceCents:   int64(input.BasePrice * 100),
	}

	resp, err := h.client.CreateFlight(c.Request.Context(), req)
	if err != nil {
		mapGRPCErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"flight_id": resp.FlightId,
	})
}

func (h *FlightHandler) GetFlightDetails(c *gin.Context) {
	flightID, err := parseIDParam(c, "id")
	if err != nil {
		return
	}

	resp, err := h.client.GetFlightDetails(c.Request.Context(), &flightv1.GetFlightDetailsRequest{
		FlightId: flightID,
	})
	if err != nil {
		mapGRPCErr(c, err)
		return
	}

	c.JSON(http.StatusOK, resp.Flight)
}

func (h *FlightHandler) GetFlightSeats(c *gin.Context) {
	flightID, err := parseIDParam(c, "id")
	if err != nil {
		return
	}

	resp, err := h.client.GetFlightSeats(c.Request.Context(), &flightv1.GetFlightSeatsRequest{
		FlightId: flightID,
	})
	if err != nil {
		mapGRPCErr(c, err)
		return
	}

	c.JSON(http.StatusOK, resp.Seats)
}

func (h *FlightHandler) ListAirports(c *gin.Context) {
	resp, err := h.client.ListAirports(c.Request.Context(), &flightv1.ListAirportsRequest{})
	if err != nil {
		mapGRPCErr(c, err)
		return
	}

	c.JSON(http.StatusOK, resp.Airports)
}

type createAircraftInput struct {
	Model      string `json:"model" binding:"required"`
	TotalSeats int32  `json:"total_seats" binding:"required,gt=0"`
}

func (h *FlightHandler) CreateAircraft(c *gin.Context) {
	var input createAircraftInput
	if err := c.ShouldBindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.client.CreateAircraft(c.Request.Context(), &flightv1.CreateAircraftRequest{
		Model:      input.Model,
		TotalSeats: input.TotalSeats,
	})
	if err != nil {
		mapGRPCErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"aircraft_id": resp.AircraftId,
	})
}

type seatTemplateInput struct {
	SeatNumber      string  `json:"seat_number" binding:"required"`
	SeatClass       string  `json:"seat_class" binding:"required"`
	PriceMultiplier float64 `json:"price_multiplier"`
}

type addAircraftSeatsInput struct {
	Seats []seatTemplateInput `json:"seats" binding:"required,min=1"`
}

func (h *FlightHandler) AddAircraftSeats(c *gin.Context) {
	aircraftID, err := parseIDParam(c, "id")
	if err != nil {
		return
	}

	var input addAircraftSeatsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	pbSeats := make([]*flightv1.AircraftSeatTemplate, 0, len(input.Seats))
	for _, s := range input.Seats {
		mult := s.PriceMultiplier
		if mult <= 0 {
			mult = 1
		}

		pbSeats = append(pbSeats, &flightv1.AircraftSeatTemplate{
			SeatNumber:      s.SeatNumber,
			SeatClass:       s.SeatClass,
			PriceMultiplier: mult,
		})
	}

	_, err = h.client.AddAircraftSeats(c.Request.Context(), &flightv1.AddAircraftSeatsRequest{
		AircraftId: aircraftID,
		Seats:      pbSeats,
	})
	if err != nil {
		mapGRPCErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *FlightHandler) ListAircrafts(c *gin.Context) {
	resp, err := h.client.ListAircrafts(c.Request.Context(), &flightv1.ListAircraftsRequest{})
	if err != nil {
		mapGRPCErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp.Aircrafts)
}
