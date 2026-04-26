package common

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	"net/http"
)

const (
	msgInvalidInput  = "invalid input body"
	msgUnauthorized  = "user not authenticated"
	msgForbidden     = "forbidden"
	msgNotFound      = "resource not found"
	msgInternalError = "internal server error"
)

// ErrorResponse is the unified JSON error body for all gateway endpoints.
//
//	@Description	Standard error response
type ErrorResponse struct {
	// Human-readable error message
	Message string `json:"message" example:"invalid input body: field email is required"`
}

// AbortWithError is the single point for writing error responses.
// All other Abort* helpers delegate to this function.
func AbortWithError(c *gin.Context, httpStatus int, message string) {
	c.AbortWithStatusJSON(httpStatus, ErrorResponse{
		Message: message,
	})
}

// AbortInvalidInput responds with 400 Bad Request.
// If err is non-nil, its text is appended to the message.
func AbortInvalidInput(c *gin.Context, err error) {
	msg := msgInvalidInput
	if err != nil {
		msg += ": " + err.Error()
	}
	AbortWithError(c, http.StatusBadRequest, msg)
}

// AbortUnauthorized responds with 401 Unauthorized.
func AbortUnauthorized(c *gin.Context) {
	AbortWithError(c, http.StatusUnauthorized, msgUnauthorized)
}

// AbortForbidden responds with 403 Forbidden.
func AbortForbidden(c *gin.Context) {
	AbortWithError(c, http.StatusForbidden, msgForbidden)
}

// AbortNotFound responds with 404 Not Found.
// If message is empty, a default text is used.
func AbortNotFound(c *gin.Context, message string) {
	if message == "" {
		message = msgNotFound
	}
	AbortWithError(c, http.StatusNotFound, message)
}

// AbortInternal responds with 500 Internal Server Error
func AbortInternal(c *gin.Context, err error) {
	if err != nil {
		slog.ErrorContext(c.Request.Context(), msgInternalError,
			slog.Any("error", err),
			slog.String("path", c.FullPath()),
			slog.String("method", c.Request.Method),
		)
	}
	AbortWithError(c, http.StatusInternalServerError, msgInternalError)
}

// HandleGRPCError maps a gRPC status to the unified error format.
func HandleGRPCError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		AbortInternal(c, err)
		return
	}

	switch st.Code() {
	case codes.InvalidArgument:
		AbortWithError(c, http.StatusBadRequest, st.Message())
	case codes.Unauthenticated:
		AbortWithError(c, http.StatusUnauthorized, st.Message())
	case codes.PermissionDenied:
		AbortWithError(c, http.StatusForbidden, st.Message())
	case codes.NotFound:
		AbortWithError(c, http.StatusNotFound, st.Message())
	case codes.AlreadyExists:
		AbortWithError(c, http.StatusConflict, st.Message())
	case codes.DeadlineExceeded:
		AbortWithError(c, http.StatusGatewayTimeout, st.Message())
	case codes.Unavailable:
		AbortWithError(c, http.StatusServiceUnavailable, st.Message())
	default:
		AbortInternal(c, err)
	}
}
