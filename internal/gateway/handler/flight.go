package handler

import (
	"github.com/gin-gonic/gin"
	flightv1 "github.com/squ1ky/flyte/gen/go/flight"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net/http"
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

	if from == "" || to == "" || dateStr == "" {
		newErrorResponse(c, http.StatusBadRequest, "from, to and date are required")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid date format (expected YYYY-MM-DD)")
		return
	}

	passengers := 1

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
	DepartureAirport string  `json:"departure_airport" binding:"required,len=3"`
	ArrivalAirport   string  `json:"arrival_airport" binding:"required,len=3"`
	DepartureTime    string  `json:"departure_time" binding:"required"`
	ArrivalTime      string  `json:"arrival_time" binding:"required"`
	Price            float64 `json:"price" binding:"required,gt=0"`
	TotalSeats       int32   `json:"total_seats" binding:"required,gt=0"`
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
		DepartureAirport: input.DepartureAirport,
		ArrivalAirport:   input.ArrivalAirport,
		DepartureTime:    timestamppb.New(depTime),
		ArrivalTime:      timestamppb.New(arrTime),
		Price:            input.Price,
		TotalSeats:       input.TotalSeats,
	}

	resp, err := h.client.CreateFlight(c.Request.Context(), req)
	if err != nil {
		mapGRPCErr(c, err)
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"flight_id": resp.FlightId,
	})
}

func (h *FlightHandler) GetFlightDetails(c *gin.Context) {
	var uri struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid flight id")
		return
	}

	resp, err := h.client.GetFlightDetails(c.Request.Context(), &flightv1.GetFlightDetailsRequest{
		FlightId: uri.ID,
	})
	if err != nil {
		mapGRPCErr(c, err)
		return
	}

	c.JSON(http.StatusOK, resp.Flight)
}

func (h *FlightHandler) GetFlightSeats(c *gin.Context) {
	var uri struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid flight id")
		return
	}

	resp, err := h.client.GetFlightSeats(c.Request.Context(), &flightv1.GetFlightSeatsRequest{
		FlightId: uri.ID,
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
