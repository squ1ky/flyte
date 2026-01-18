package router

import (
	"github.com/gin-gonic/gin"
	userv1 "github.com/squ1ky/flyte/gen/go/user"
	"github.com/squ1ky/flyte/internal/gateway/handler"
)

func RegisterFlightRoutes(rg *gin.RouterGroup, h *handler.GatewayHandler, userClient userv1.UserServiceClient) {
	flights := rg.Group("/flights")
	{
		flights.GET("", h.Flight.SearchFlights)
		flights.GET("/:id", h.Flight.GetFlightDetails)
		flights.GET("/:id/seats", h.Flight.GetFlightSeats)
	}

	rg.GET("/airports", h.Flight.ListAirports)
	rg.GET("/aircrafts", h.Flight.ListAircrafts)

	admin := rg.Group("", AuthMiddleware(userClient), AdminOnlyMiddleware())
	{
		admin.POST("/flights", h.Flight.CreateFlight)
		admin.POST("/aircrafts", h.Flight.CreateAircraft)
		admin.POST("/aircrafts/:id/seats", h.Flight.AddAircraftSeats)
	}
}
