package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func ParseID(c *gin.Context, paramName string) (int64, bool) {
	idStr := c.Param(paramName)
	if idStr == "" {
		NewErrorResponse(c, http.StatusBadRequest, "missing id parameter")
		return 0, false
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		NewErrorResponse(c, http.StatusBadRequest, "invalid id parameter")
		return 0, false
	}

	return id, true
}
