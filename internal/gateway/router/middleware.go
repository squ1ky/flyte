package router

import (
	"github.com/gin-gonic/gin"
	userv1 "github.com/squ1ky/flyte/gen/go/user"
	"github.com/squ1ky/flyte/internal/gateway/common"
	"github.com/squ1ky/flyte/pkg/api"
	"net/http"
	"strings"
)

func AuthMiddleware(client userv1.UserServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "empty auth header",
			})
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid auth header",
			})
			return
		}

		resp, err := client.ValidateToken(c.Request.Context(), &userv1.ValidateTokenRequest{
			Token: headerParts[1],
		})

		if err != nil || !resp.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			return
		}

		common.SetUser(c, resp.UserId, resp.Role)

		c.Next()
	}
}

func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := common.GetUserRole(c)
		if role != "admin" {
			api.AbortForbidden(c)
			return
		}
		c.Next()
	}
}
