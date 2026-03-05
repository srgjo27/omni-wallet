package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/omni-wallet/user-service/internal/adapter/handler/http/response"
	"github.com/omni-wallet/user-service/internal/core/services"
)

const (
	AuthUserIDKey = "auth_user_id"
	AuthEmailKey = "auth_email"
)

// AuthMiddleware validates the JWT Bearer token on protected routes.
// It uses the UserService to verify the token, which keeps JWT logic in the service layer.
func AuthMiddleware(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "authorization header is required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Unauthorized(c, "authorization header format must be: Bearer <token>")
			c.Abort()
			return
		}

		claims, err := userService.VerifyToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set(AuthUserIDKey, claims.UserID)
		c.Set(AuthEmailKey, claims.Email)
		c.Next()
	}
}
