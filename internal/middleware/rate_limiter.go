package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiterConfig struct {
	RedisClient *redis.Client
	MaxRequests int           // Maximum number of requests allowed
	Window      time.Duration // Time window for rate limiting
}

func NewRateLimiterMiddleware(config RateLimiterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP address as key
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		ctx := context.Background()
		pipe := config.RedisClient.Pipeline()

		// Add current timestamp to sorted set
		now := time.Now().UnixNano()
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})

		// Remove old entries outside the window
		windowStart := time.Now().Add(-config.Window).UnixNano()
		pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

		// Count requests in current window
		pipe.ZCard(ctx, key)

		// Set key expiration
		pipe.Expire(ctx, key, config.Window)

		// Execute pipeline
		cmds, err := pipe.Exec(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "rate limit check failed"})
			c.Abort()
			return
		}

		// Get request count from pipeline result
		requestCount := cmds[2].(*redis.IntCmd).Val()

		if requestCount > int64(config.MaxRequests) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": fmt.Sprintf("rate limit exceeded. maximum %d requests allowed per %v",
					config.MaxRequests, config.Window),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Example usage:
// config := RateLimiterConfig{
//     RedisClient: redisClient,
//     MaxRequests: 100,
//     Window:      time.Minute,
// }
// router.Use(NewRateLimiterMiddleware(config))
