package booking

import (
	"github.com/gin-gonic/gin"
	bookingv1 "github.com/squ1ky/flyte/gen/go/booking"
	"github.com/squ1ky/flyte/internal/gateway/common"
	"github.com/squ1ky/flyte/pkg/api"
	"net/http"
	"time"
)

const (
	ErrTextEmptyBookingID = "empty booking id"
)

type Handler struct {
	client bookingv1.BookingServiceClient
}

func NewHandler(client bookingv1.BookingServiceClient) *Handler {
	return &Handler{client: client}
}

func (h *Handler) Create(c *gin.Context) {
	userID, ok := common.GetUserID(c)
	if !ok {
		return
	}

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.AbortInvalidInput(c, err)
		return
	}

	resp, err := h.client.CreateBooking(c.Request.Context(), &bookingv1.CreateBookingRequest{
		UserId:            userID,
		FlightId:          req.FlightId,
		SeatNumber:        req.SeatNumber,
		PassengerName:     req.PassengerName,
		PassengerPassport: req.PassengerPassport,
		PriceCents:        req.PriceCents,
		Currency:          req.Currency,
	})
	if err != nil {
		api.HandleGRPCErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, CreateResponse{
		BookingID: resp.BookingId,
	})
}

func (h *Handler) Get(c *gin.Context) {
	userID, ok := common.GetUserID(c)
	if !ok {
		return
	}

	bookingID := c.Param("id")
	if bookingID == "" {
		api.NewErrorResponse(c, http.StatusBadRequest, ErrTextEmptyBookingID)
		return
	}

	resp, err := h.client.GetBooking(c.Request.Context(), &bookingv1.GetBookingRequest{
		BookingId: bookingID,
	})
	if err != nil {
		api.HandleGRPCErr(c, err)
		return
	}

	if resp.Booking.UserId != userID {
		api.AbortForbidden(c)
		return
	}

	c.JSON(http.StatusOK, mapProtoToResponse(resp.Booking))
}

func (h *Handler) List(c *gin.Context) {
	userID, ok := common.GetUserID(c)
	if !ok {
		return
	}

	resp, err := h.client.ListBookings(c.Request.Context(), &bookingv1.ListBookingsRequest{
		UserId: userID,
	})
	if err != nil {
		api.HandleGRPCErr(c, err)
		return
	}

	bookings := make([]Response, len(resp.Bookings))
	for i, b := range resp.Bookings {
		bookings[i] = mapProtoToResponse(b)
	}

	c.JSON(http.StatusOK, bookings)
}

func (h *Handler) Cancel(c *gin.Context) {
	userID, ok := common.GetUserID(c)
	if !ok {
		return
	}

	bookingID := c.Param("id")
	if bookingID == "" {
		api.NewErrorResponse(c, http.StatusBadRequest, ErrTextEmptyBookingID)
		return
	}

	getResp, err := h.client.GetBooking(c.Request.Context(), &bookingv1.GetBookingRequest{
		BookingId: bookingID,
	})
	if err != nil {
		api.HandleGRPCErr(c, err)
		return
	}

	if getResp.Booking.UserId != userID {
		api.AbortForbidden(c)
		return
	}

	_, err = h.client.CancelBooking(c.Request.Context(), &bookingv1.CancelBookingRequest{
		BookingId: bookingID,
	})
	if err != nil {
		api.HandleGRPCErr(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func mapProtoToResponse(b *bookingv1.Booking) Response {
	return Response{
		ID:                b.Id,
		FlightID:          b.FlightId,
		SeatNumber:        b.SeatNumber,
		PassengerName:     b.PassengerName,
		PassengerPassport: b.PassengerPassport,
		Status:            b.Status,
		PriceCents:        b.PriceCents,
		Currency:          b.Currency,
		CreatedAt:         b.CreatedAt.AsTime().Format(time.RFC3339),
	}
}
