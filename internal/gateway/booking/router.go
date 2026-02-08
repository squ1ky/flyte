package booking

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc) {
	bookings := rg.Group("/bookings", authMiddleware)
	{
		bookings.POST("/", h.Create)
		bookings.GET("/", h.List)
		bookings.GET("/:id", h.Get)
		bookings.POST("/:id/cancel", h.Cancel)
	}
}
