package common

import (
	"github.com/gin-gonic/gin"
	"github.com/squ1ky/flyte/pkg/api"
)

const (
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
		api.AbortUnauthorized(c)
		return 0, false
	}

	id, ok := idVal.(int64)
	if !ok {
		api.AbortInternal(c, nil)
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
