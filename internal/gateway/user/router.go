package user

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc) {
	auth := rg.Group("/auth")
	{
		auth.POST("/sign-up", h.SignUp)
		auth.POST("/sign-in", h.SignIn)
	}

	users := rg.Group("/users", authMiddleware)
	{
		users.POST("/:id/passengers", h.AddPassenger)
		users.GET("/:id/passengers", h.GetPassengers)
	}
}
