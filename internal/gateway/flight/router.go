package flight

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, authMiddleware, adminMiddleware gin.HandlerFunc) {
	flights := rg.Group("/flights")
	{
		flights.GET("", h.SearchFlights)
		flights.GET("/:id", h.GetFlightDetails)
		flights.GET("/:id/fares", h.GetFlightFares)
		flights.GET("/:id/seats/reserved/public", h.GetReservedSeatNumbers)
	}

	airports := rg.Group("/airports")
	{
		airports.GET("/search", h.SearchAirports)
		airports.GET("/:code", h.GetAirport)
	}

	airlines := rg.Group("/airlines")
	{
		airlines.GET("", h.ListAirlines)
		airlines.GET("/:id", h.GetAirline)
	}

	aircrafts := rg.Group("/aircrafts")
	{
		aircrafts.GET("/:id/seats", h.GetAircraftSeats)
	}

	admin := rg.Group("", authMiddleware, adminMiddleware)
	{
		admin.POST("/flights", h.CreateFlight)
		admin.PATCH("/flights/:id/status", h.UpdateFlightStatus)
		admin.GET("/flights/:id/seats/reserved", h.GetReservedSeats)

		admin.POST("/airlines", h.CreateAirline)

		admin.POST("/aircrafts", h.CreateAircraft)
		admin.POST("/aircrafts/:id/seats", h.AddAircraftSeats)
	}
}
