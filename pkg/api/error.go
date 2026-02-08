package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	ErrTextUnauthorized = "user not authenticated"
	ErrTextInvalidInput = "invalid input body"
	ErrTextInternal     = "internal server error"
	ErrTextAccessDenied = "access denied"
	ErrTextNotFound     = "resource not found"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewErrorResponse(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, ErrorResponse{Error: message})
}

func AbortUnauthorized(c *gin.Context) {
	NewErrorResponse(c, http.StatusUnauthorized, ErrTextUnauthorized)
}

func AbortForbidden(c *gin.Context) {
	NewErrorResponse(c, http.StatusForbidden, ErrTextAccessDenied)
}

func AbortInvalidInput(c *gin.Context, err error) {
	msg := ErrTextInvalidInput
	if err != nil {
		msg += ": " + err.Error()
	}
	NewErrorResponse(c, http.StatusBadRequest, msg)
}

func AbortNotFound(c *gin.Context, message string) {
	if message == "" {
		message = ErrTextNotFound
	}
	NewErrorResponse(c, http.StatusNotFound, message)
}

func AbortInternal(c *gin.Context, err error) {
	NewErrorResponse(c, http.StatusInternalServerError, ErrTextInternal)
}
