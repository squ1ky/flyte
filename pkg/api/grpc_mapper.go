package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HandleGRPCErr(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	switch st.Code() {
	case codes.NotFound:
		NewErrorResponse(c, http.StatusNotFound, st.Message())
	case codes.InvalidArgument:
		NewErrorResponse(c, http.StatusBadRequest, st.Message())
	case codes.AlreadyExists:
		NewErrorResponse(c, http.StatusConflict, st.Message())
	case codes.Unauthenticated:
		NewErrorResponse(c, http.StatusUnauthorized, st.Message())
	case codes.PermissionDenied:
		NewErrorResponse(c, http.StatusForbidden, st.Message())
	case codes.DeadlineExceeded:
		NewErrorResponse(c, http.StatusGatewayTimeout, st.Message())
	case codes.Unavailable:
		NewErrorResponse(c, http.StatusServiceUnavailable, st.Message())
	default:
		NewErrorResponse(c, http.StatusInternalServerError, st.Message())
	}
}
