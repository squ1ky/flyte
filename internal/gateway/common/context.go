package common

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

const (
	RoleUser  = "user"
	RoleAdmin = "admin"

	ctxKeyUserID = "userId"
	ctxKeyRole   = "role"
)

func SetUser(c *gin.Context, id int64, role string) {
	c.Set(ctxKeyUserID, id)
	c.Set(ctxKeyRole, role)
}

func GetUserID(c *gin.Context) (int64, bool) {
	idVal, exists := c.Get(ctxKeyUserID)
	if !exists {
		AbortUnauthorized(c)
		return 0, false
	}

	id, ok := idVal.(int64)
	if !ok {
		AbortInternal(c, nil)
		return 0, false
	}

	return id, true
}

func GetUserRole(c *gin.Context) string {
	roleVal, exists := c.Get(ctxKeyRole)
	if !exists {
		return ""
	}

	role, ok := roleVal.(string)
	if !ok {
		return ""
	}
	return role
}

func ParseID(c *gin.Context, paramName string) (int64, bool) {
	idStr := c.Param(paramName)
	if idStr == "" {
		AbortWithError(c, http.StatusBadRequest, fmt.Sprintf("missing %s parameter", paramName))
		return 0, false
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		AbortWithError(c, http.StatusBadRequest, fmt.Sprintf("invalid %s parameter", paramName))
		return 0, false
	}

	return id, true
}

func QueryInt(c *gin.Context, name string, defaultVal int) int {
	s := c.Query(name)
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}
