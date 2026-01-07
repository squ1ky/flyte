package router

import (
	"github.com/gin-gonic/gin"
	"github.com/squ1ky/flyte/internal/gateway/handler"
)

func RegisterFlightRoutes(rg *gin.RouterGroup, h *handler.GatewayHandler) {
	flights := rg.Group("/flights")
	{
		flights.GET("/search", h.Flight.SearchFlights)
		flights.GET("/:id", h.Flight.GetFlightDetails)

		flights.GET("/:id/seats", h.Flight.GetFlightSeats)
	}

	rg.GET("/airports", h.Flight.ListAirports)
}
