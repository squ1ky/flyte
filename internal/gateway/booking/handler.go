package booking

import (
	"github.com/gin-gonic/gin"
	bookingv1 "github.com/squ1ky/flyte/gen/proto/booking"
	"github.com/squ1ky/flyte/internal/gateway/common"
	"net/http"
	"time"
)

const (
	paramBookingID = "id"
	paramRefCode   = "ref"
)

const (
	msgMissingBookingID = "missing booking id"
	msgMissingRefCode   = "missing ref code"
)

type Handler struct {
	client bookingv1.BookingServiceClient
}

func NewHandler(client bookingv1.BookingServiceClient) *Handler {
	return &Handler{client: client}
}

// Create godoc
//
//	@Summary		Create a booking
//	@Description	Creates a new booking with one or more passengers and reserves their seats.
//	@Tags			bookings
//	@Accept			json
//	@Produce		json
//	@Param			input	body		CreateBookingRequest	true	"Booking data"
//	@Success		201		{object}	CreateBookingResponse	"Booking created"
//	@Failure		400		{object}	common.ErrorResponse	"Invalid input"
//	@Failure		409		{object}	common.ErrorResponse	"Seat already taken"
//	@Failure		500		{object}	common.ErrorResponse	"Internal server error"
//	@Security		BearerAuth
//	@Router			/bookings [post]
func (h *Handler) Create(c *gin.Context) {
	userID, ok := common.GetUserID(c)
	if !ok {
		return
	}

	var input CreateBookingRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		common.AbortInvalidInput(c, err)
		return
	}

	pbPassengers := make([]*bookingv1.Passenger, len(input.Passengers))
	for i, passenger := range input.Passengers {
		pbPassengers[i] = &bookingv1.Passenger{
			FirstName:      passenger.FirstName,
			LastName:       passenger.LastName,
			DocumentNumber: passenger.DocumentNumber,
			SeatNumber:     passenger.SeatNumber,
			SeatClass:      seatClassToProto(passenger.SeatClass),
		}
	}

	resp, err := h.client.CreateBooking(c.Request.Context(), &bookingv1.CreateBookingRequest{
		UserId:       userID,
		FlightId:     input.FlightID,
		ContactEmail: input.ContactEmail,
		Passengers:   pbPassengers,
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	result := CreateBookingResponse{
		BookingID: resp.BookingId,
		RefCode:   resp.RefCode,
	}
	if resp.ExpiresAt != nil {
		result.ExpiresAt = resp.ExpiresAt.AsTime().Format(time.RFC3339)
	}

	c.JSON(http.StatusCreated, result)
}

// Get godoc
//
//	@Summary		Get booking details
//	@Description	Returns full details for a booking. Only the booking owner can access it.
//	@Tags			bookings
//	@Produce		json
//	@Param			id	path		string	true	"Booking ID (UUID)"
//	@Success		200	{object}	BookingResponse			"Booking details"
//	@Failure		400	{object}	common.ErrorResponse	"Invalid booking ID"
//	@Failure		403	{object}	common.ErrorResponse	"Forbidden — not the booking owner"
//	@Failure		404	{object}	common.ErrorResponse	"Booking not found"
//	@Failure		500	{object}	common.ErrorResponse	"Internal server error"
//	@Security		BearerAuth
//	@Router			/bookings/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	userID, ok := common.GetUserID(c)
	if !ok {
		return
	}

	bookingID := c.Param(paramBookingID)
	if bookingID == "" {
		common.AbortWithError(c, http.StatusBadRequest, msgMissingBookingID)
		return
	}

	resp, err := h.client.GetBooking(c.Request.Context(), &bookingv1.GetBookingRequest{
		BookingId: bookingID,
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	if resp.Booking.UserId != userID {
		common.AbortForbidden(c)
		return
	}

	c.JSON(http.StatusOK, mapProtoToBooking(resp.Booking))
}

// GetByRefCode godoc
//
//	@Summary		Get booking by reference code
//	@Description	Returns full details for a booking by its short reference code. Only the booking owner can access it.
//	@Tags			bookings
//	@Produce		json
//	@Param			ref	path		string	true	"Reference code"
//	@Success		200	{object}	BookingResponse			"Booking details"
//	@Failure		403	{object}	common.ErrorResponse	"Forbidden — not the booking owner"
//	@Failure		404	{object}	common.ErrorResponse	"Booking not found"
//	@Failure		500	{object}	common.ErrorResponse	"Internal server error"
//	@Security		BearerAuth
//	@Router			/bookings/ref/{ref} [get]
func (h *Handler) GetByRefCode(c *gin.Context) {
	userID, ok := common.GetUserID(c)
	if !ok {
		return
	}

	refCode := c.Param(paramRefCode)
	if refCode == "" {
		common.AbortWithError(c, http.StatusBadRequest, msgMissingRefCode)
		return
	}

	resp, err := h.client.GetBooking(c.Request.Context(), &bookingv1.GetBookingRequest{
		RefCode: refCode,
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	if resp.Booking.UserId != userID {
		common.AbortForbidden(c)
		return
	}

	c.JSON(http.StatusOK, mapProtoToBooking(resp.Booking))
}

// List godoc
//
//	@Summary		List user bookings
//	@Description	Returns all bookings for the authenticated user with pagination.
//	@Tags			bookings
//	@Produce		json
//	@Param			limit	query		int	false	"Max results"	default(20)
//	@Param			offset	query		int	false	"Offset"		default(0)
//	@Success		200		{array}		BookingResponse			"List of bookings"
//	@Failure		500		{object}	common.ErrorResponse	"Internal server error"
//	@Security		BearerAuth
//	@Router			/bookings [get]
func (h *Handler) List(c *gin.Context) {
	userID, ok := common.GetUserID(c)
	if !ok {
		return
	}

	const defaultLimit = 20
	const defaultOffset = 0

	limit := int32(common.QueryInt(c, "limit", defaultLimit))
	offset := int32(common.QueryInt(c, "offset", defaultOffset))

	resp, err := h.client.ListBookings(c.Request.Context(), &bookingv1.ListBookingsRequest{
		UserId: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	bookings := make([]BookingResponse, len(resp.Bookings))
	for i, booking := range resp.Bookings {
		bookings[i] = mapProtoToBooking(booking)
	}

	c.JSON(http.StatusOK, bookings)
}

// Cancel godoc
//
//	@Summary		Cancel a booking
//	@Description	Cancels a booking and releases reserved seats. Only the booking owner can cancel.
//	@Tags			bookings
//	@Param			id	path	string	true	"Booking ID (UUID)"
//	@Success		204	"Booking cancelled"
//	@Failure		400	{object}	common.ErrorResponse	"Invalid booking ID"
//	@Failure		403	{object}	common.ErrorResponse	"Forbidden — not the booking owner"
//	@Failure		404	{object}	common.ErrorResponse	"Booking not found"
//	@Failure		500	{object}	common.ErrorResponse	"Internal server error"
//	@Security		BearerAuth
//	@Router			/bookings/{id}/cancel [post]
func (h *Handler) Cancel(c *gin.Context) {
	userID, ok := common.GetUserID(c)
	if !ok {
		return
	}

	bookingID := c.Param(paramBookingID)
	if bookingID == "" {
		common.AbortWithError(c, http.StatusBadRequest, msgMissingBookingID)
		return
	}

	// Verify ownership before cancelling
	getResp, err := h.client.GetBooking(c.Request.Context(), &bookingv1.GetBookingRequest{
		BookingId: bookingID,
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	if getResp.Booking.UserId != userID {
		common.AbortForbidden(c)
		return
	}

	_, err = h.client.CancelBooking(c.Request.Context(), &bookingv1.CancelBookingRequest{
		BookingId: bookingID,
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
