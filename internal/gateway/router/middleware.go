package router

import (
	"github.com/gin-gonic/gin"
	userv1 "github.com/squ1ky/flyte/gen/proto/user"
	"github.com/squ1ky/flyte/internal/gateway/common"
	"strings"
)

const (
	headerAuthorization = "Authorization"
	prefixBearer        = "Bearer"
)

// AuthMiddleware validates the Bearer token via the UserService
// and stores the authenticated user's ID and role in the request context.
func AuthMiddleware(client userv1.UserServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(headerAuthorization)
		if authHeader == "" {
			common.AbortUnauthorized(c)
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != prefixBearer {
			common.AbortUnauthorized(c)
			return
		}

		resp, err := client.ValidateToken(c.Request.Context(), &userv1.ValidateTokenRequest{
			Token: headerParts[1],
		})

		if err != nil || !resp.Valid {
			common.AbortUnauthorized(c)
			return
		}

		common.SetUser(c, resp.UserId, common.RoleFromProto(resp.Role))

		c.Next()
	}
}

// AdminOnlyMiddleware restricts access with the admin role.
// Must be applied after AuthMiddleware.
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := common.GetUserRole(c)
		if role != common.RoleAdmin {
			common.AbortForbidden(c)
			return
		}
		c.Next()
	}
}
