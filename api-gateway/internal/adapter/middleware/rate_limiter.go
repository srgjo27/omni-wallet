package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimiter(redisClient *redis.Client, requestsPerWindow int, windowDuration time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientIP := ctx.ClientIP()

		windowBucket := time.Now().Truncate(windowDuration).Unix()
		key := fmt.Sprintf("rate_limit:%s:%d", clientIP, windowBucket)

		bg := context.Background()

		count, err := redisClient.Incr(bg, key).Result()
		if err != nil {
			ctx.Next()
			return
		}

		if count == 1 {
			redisClient.Expire(bg, key, windowDuration)
		}

		ctx.Header("X-RateLimit-Limit", fmt.Sprintf("%d", requestsPerWindow))
		ctx.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, int64(requestsPerWindow)-count)))

		if count > int64(requestsPerWindow) {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "rate limit exceeded — please slow down",
			})
			return
		}

		ctx.Next()
	}
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
