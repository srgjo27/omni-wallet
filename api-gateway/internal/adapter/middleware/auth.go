package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// jwtClaims defines the expected claims structure inside the JWT.
type jwtClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// AuthMiddleware returns a Gin middleware that validates a Bearer JWT token.
// On success it injects auth_user_id and auth_email into the request context
// and forwards the original Authorization header to the upstream service.
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "missing or invalid Authorization header",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims := &jwtClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "invalid or expired token",
			})
			return
		}

		// Inject verified identity into the context so upstream services
		// can rely on these headers without re-verifying the token.
		ctx.Set("auth_user_id", claims.UserID)
		ctx.Set("auth_email", claims.Email)
		ctx.Request.Header.Set("X-Auth-User-ID", claims.UserID)
		ctx.Request.Header.Set("X-Auth-Email", claims.Email)

		ctx.Next()
	}
}
