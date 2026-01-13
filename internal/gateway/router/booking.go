package router

import (
	"github.com/gin-gonic/gin"
	userv1 "github.com/squ1ky/flyte/gen/go/user"
	"github.com/squ1ky/flyte/internal/gateway/handler"
)

func RegisterBookingRoutes(
	rg *gin.RouterGroup,
	h *handler.GatewayHandler,
	userClient userv1.UserServiceClient,
) {
	bookings := rg.Group("/bookings", AuthMiddleware(userClient))
	{
		bookings.POST("/", h.Booking.CreateBooking)
		bookings.GET("/", h.Booking.ListBookings)
		bookings.GET("/:id", h.Booking.GetBooking)
		bookings.POST("/:id/cancel", h.Booking.CancelBooking)
	}
}
