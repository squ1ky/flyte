package router

import (
	"github.com/gin-gonic/gin"
	userv1 "github.com/squ1ky/flyte/gen/go/user"
	"github.com/squ1ky/flyte/internal/gateway/handler"
)

func RegisterUserRoutes(
	rg *gin.RouterGroup,
	h *handler.GatewayHandler,
	userClient userv1.UserServiceClient,
) {
	auth := rg.Group("/auth")
	{
		auth.POST("/sign-up", h.User.SignUp)
		auth.POST("/sign-in", h.User.SignIn)
	}

	users := rg.Group("/users", AuthMiddleware(userClient))
	{
		users.POST("/:id/passengers", h.User.AddPassenger)
		users.GET("/:id/passengers", h.User.GetPassengers)
	}
}
