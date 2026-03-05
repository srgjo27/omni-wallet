package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter returns a Gin middleware that enforces a sliding-window rate
// limit per client IP address using Redis INCR + EXPIRE.
//
// Algorithm:
//  1. Build a Redis key from the client IP and the current time window bucket.
//  2. Atomically increment the request counter for that key.
//  3. On the very first request (count == 1), set the TTL equal to the window.
//  4. If the count exceeds the allowed limit, reject with HTTP 429.
func RateLimiter(redisClient *redis.Client, requestsPerWindow int, windowDuration time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientIP := ctx.ClientIP()
		// Create a time-bucketed key so each window starts fresh.
		windowBucket := time.Now().Truncate(windowDuration).Unix()
		key := fmt.Sprintf("rate_limit:%s:%d", clientIP, windowBucket)

		bg := context.Background()

		// Increment counter and get new value in one round-trip.
		count, err := redisClient.Incr(bg, key).Result()
		if err != nil {
			// Fail open: if Redis is unavailable, allow the request through
			// and log the error rather than blocking all traffic.
			ctx.Next()
			return
		}

		// Set TTL only on the first request to ensure window expiry.
		if count == 1 {
			redisClient.Expire(bg, key, windowDuration)
		}

		// Attach rate-limit metadata headers for observability.
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

// max returns the larger of two int64 values.
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
