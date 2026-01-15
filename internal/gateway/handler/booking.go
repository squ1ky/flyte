package handler

import (
	"github.com/gin-gonic/gin"
	bookingv1 "github.com/squ1ky/flyte/gen/go/booking"
	"net/http"
)

type BookingHandler struct {
	client bookingv1.BookingServiceClient
}

func NewBookingHandler(client bookingv1.BookingServiceClient) *BookingHandler {
	return &BookingHandler{
		client: client,
	}
}

type createBookingInput struct {
	FlightId          int64   `json:"flightId" binding:"required,gt=0"`
	SeatNumber        string  `json:"seat_number" binding:"required"`
	PassengerName     string  `json:"passenger_name" binding:"required"`
	PassengerPassport string  `json:"passenger_passport" binding:"required"`
	Price             float64 `json:"price" binding:"required,gt=0"`
	Currency          string  `json:"currency"`
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		newErrorResponse(c, http.StatusUnauthorized, ErrUserUnauthorized)
		return
	}

	var inp createBookingInput
	if err := c.ShouldBindJSON(&inp); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.client.CreateBooking(c.Request.Context(), &bookingv1.CreateBookingRequest{
		UserId:            userID.(int64),
		FlightId:          inp.FlightId,
		SeatNumber:        inp.SeatNumber,
		PassengerName:     inp.PassengerName,
		PassengerPassport: inp.PassengerPassport,
		PriceCents:        int64(inp.Price * 100),
		Currency:          inp.Currency,
	})
	if err != nil {
		mapGRPCErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"booking_id": resp.BookingId,
	})
}

func (h *BookingHandler) GetBooking(c *gin.Context) {
	bookingID := c.Param("id")
	if bookingID == "" {
		newErrorResponse(c, http.StatusBadRequest, "empty booking id")
		return
	}

	// TODO: check owner

	resp, err := h.client.GetBooking(c.Request.Context(), &bookingv1.GetBookingRequest{
		BookingId: bookingID,
	})
	if err != nil {
		mapGRPCErr(c, err)
		return
	}

	c.JSON(http.StatusOK, resp.Booking)
}

func (h *BookingHandler) ListBookings(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		newErrorResponse(c, http.StatusUnauthorized, ErrUserUnauthorized)
		return
	}

	resp, err := h.client.ListBookings(c.Request.Context(), &bookingv1.ListBookingsRequest{
		UserId: userID.(int64),
	})
	if err != nil {
		mapGRPCErr(c, err)
		return
	}

	c.JSON(http.StatusOK, resp.Bookings)
}

func (h *BookingHandler) CancelBooking(c *gin.Context) {
	bookingID := c.Param("id")
	if bookingID == "" {
		newErrorResponse(c, http.StatusBadRequest, "empty booking id")
		return
	}

	// TODO: check owner

	_, err := h.client.CancelBooking(c.Request.Context(), &bookingv1.CancelBookingRequest{
		BookingId: bookingID,
	})
	if err != nil {
		mapGRPCErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "booking cancelled",
	})
}
