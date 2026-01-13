package handler

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

type GatewayHandler struct {
	User    *UserHandler
	Flight  *FlightHandler
	Booking *BookingHandler
}

func NewGatewayHandler(
	user *UserHandler,
	flight *FlightHandler,
	booking *BookingHandler,
) *GatewayHandler {
	return &GatewayHandler{
		User:    user,
		Flight:  flight,
		Booking: booking,
	}
}

func mapGRPCErr(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	switch st.Code() {
	case codes.NotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
	case codes.InvalidArgument:
		c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
	case codes.AlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": st.Message()})
	case codes.Unauthenticated:
		c.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
	}
}
