package router

import (
	"github.com/gin-gonic/gin"
	userv1 "github.com/squ1ky/flyte/gen/go/user"
	"github.com/squ1ky/flyte/internal/gateway/handler"
)

func RegisterFlightRoutes(rg *gin.RouterGroup, h *handler.GatewayHandler, userClient userv1.UserServiceClient) {
	flights := rg.Group("/flights")
	{
		flights.GET("/search", h.Flight.SearchFlights)
		flights.GET("/:id", h.Flight.GetFlightDetails)

		flights.GET("/:id/seats", h.Flight.GetFlightSeats)
	}

	rg.GET("/airports", h.Flight.ListAirports)

	admin := rg.Group("/flights", AuthMiddleware(userClient), AdminOnlyMiddleware())
	{
		admin.POST("/", h.Flight.CreateFlight)
	}
}
