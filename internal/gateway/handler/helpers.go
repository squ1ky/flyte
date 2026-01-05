package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

const (
	ErrInvalidInputBody = "invalid input body"
	ErrUserUnauthorized = "user not authenticated"
	ErrAccessDenied     = "access denied"
	ErrInternalServer   = "internal server error"
)

func newErrorResponse(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, gin.H{"error": message})
}

func parseIDParam(c *gin.Context, paramName string) (int64, error) {
	idStr := c.Param(paramName)
	if idStr == "" {
		newErrorResponse(c, http.StatusBadRequest, "missing id param")
		return 0, errors.New("empty param")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id param")
		return 0, err
	}

	return id, nil
}

func verifyUserOwnership(c *gin.Context, targetUserID int64) bool {
	tokenID, exists := c.Get("userId")
	if !exists {
		newErrorResponse(c, http.StatusUnauthorized, ErrUserUnauthorized)
		return false
	}

	currentUserID, ok := tokenID.(int64)
	if !ok {
		newErrorResponse(c, http.StatusInternalServerError, ErrInternalServer)
		return false
	}

	if targetUserID != currentUserID {
		newErrorResponse(c, http.StatusForbidden, ErrAccessDenied)
		return false
	}

	return true
}
