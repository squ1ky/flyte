package router

import (
	"github.com/gin-gonic/gin"
	userv1 "github.com/squ1ky/flyte/gen/go/user"
	"github.com/squ1ky/flyte/internal/gateway/handler"
)

func NewRouter(h *handler.GatewayHandler, userClient userv1.UserServiceClient) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1")

	RegisterUserRoutes(v1, h, userClient)

	return r
}
