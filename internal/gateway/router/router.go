package router

import (
	"github.com/gin-gonic/gin"
	userv1 "github.com/squ1ky/flyte/gen/go/user"
	"github.com/squ1ky/flyte/internal/gateway/handler"
)

func InitRoutes(h *handler.GatewayHandler, userClient userv1.UserServiceClient) *gin.Engine {
	router := gin.Default()

	api := router.Group("/api/v1")
	{
		RegisterUserRoutes(api, h, userClient)
		RegisterFlightRoutes(api, h, userClient)
		RegisterBookingRoutes(api, h, userClient)
	}

	return router
}
