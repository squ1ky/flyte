package router

import (
	"github.com/gin-gonic/gin"
	flightv1 "github.com/squ1ky/flyte/gen/go/flight"
	userv1 "github.com/squ1ky/flyte/gen/go/user"
	"github.com/squ1ky/flyte/internal/gateway/handler"
)

type Router struct {
	handler      *handler.GatewayHandler
	userClient   userv1.UserServiceClient
	flightClient flightv1.FlightServiceClient
}

func NewRouter(h *handler.GatewayHandler, userClient userv1.UserServiceClient) *Router {
	return &Router{
		handler:    h,
		userClient: userClient,
	}
}

func (r *Router) Run(addr string) error {
	return r.InitRoutes().Run(addr)
}

func (r *Router) InitRoutes() *gin.Engine {
	router := gin.Default()

	api := router.Group("/api/v1")
	{
		RegisterUserRoutes(api, r.handler, r.userClient)
		RegisterFlightRoutes(api, r.handler)
	}

	return router
}
