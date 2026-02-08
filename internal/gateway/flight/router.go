package flight

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, authMiddleware, adminMiddleware gin.HandlerFunc) {
	flights := rg.Group("/flights")
	{
		flights.GET("", h.SearchFlights)
		flights.GET("/:id", h.GetFlightDetails)
		flights.GET("/:id/seats", h.GetFlightSeats)
	}

	rg.GET("/airports", h.ListAirports)
	rg.GET("/aircrafts", h.ListAircrafts)

	admin := rg.Group("", authMiddleware, adminMiddleware)
	{
		admin.POST("/flights", h.CreateFlight)
		admin.POST("/aircrafts", h.CreateAircraft)
		admin.POST("/aircrafts/:id/seats", h.AddAircraftSeats)
	}
}
