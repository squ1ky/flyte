package router

import (
	"github.com/gin-gonic/gin"
	bookingv1 "github.com/squ1ky/flyte/gen/go/booking"
	flightv1 "github.com/squ1ky/flyte/gen/go/flight"
	userv1 "github.com/squ1ky/flyte/gen/go/user"
	"github.com/squ1ky/flyte/internal/gateway/booking"
	"github.com/squ1ky/flyte/internal/gateway/flight"
	"github.com/squ1ky/flyte/internal/gateway/user"
)

func InitRoutes(
	userClient userv1.UserServiceClient,
	flightClient flightv1.FlightServiceClient,
	bookingClient bookingv1.BookingServiceClient,
) *gin.Engine {
	r := gin.Default()

	authMiddleware := AuthMiddleware(userClient)
	adminMiddleware := AdminOnlyMiddleware()

	userHandler := user.NewHandler(userClient)
	flightHandler := flight.NewHandler(flightClient)
	bookingHandler := booking.NewHandler(bookingClient)

	apiGroup := r.Group("/api/v1")

	user.RegisterRoutes(apiGroup, userHandler, authMiddleware)
	flight.RegisterRoutes(apiGroup, flightHandler, authMiddleware, adminMiddleware)
	booking.RegisterRoutes(apiGroup, bookingHandler, authMiddleware)

	return r
}
