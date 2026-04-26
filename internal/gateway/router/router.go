package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	bookingv1 "github.com/squ1ky/flyte/gen/proto/booking"
	flightv1 "github.com/squ1ky/flyte/gen/proto/flight"
	userv1 "github.com/squ1ky/flyte/gen/proto/user"
	"github.com/squ1ky/flyte/internal/gateway/booking"
	"github.com/squ1ky/flyte/internal/gateway/config"
	"github.com/squ1ky/flyte/internal/gateway/flight"
	"github.com/squ1ky/flyte/internal/gateway/user"
	"time"
)

func InitRoutes(
	cfg config.HTTPConfig,
	userClient userv1.UserServiceClient,
	flightClient flightv1.FlightServiceClient,
	bookingClient bookingv1.BookingServiceClient,
) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.AllowedOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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
