package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"go.uber.org/zap"
)

// RateLimiter creates a new rate limiting middleware
func RateLimiter(redis *cache.Redis, maxRequests int, window time.Duration, cfg *config.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// Get client IP
		clientIP := c.ClientIP()

		// Build rate limit key
		key := fmt.Sprintf("ratelimit:%s", clientIP)

		// Get current count
		countStr, err := redis.Get(ctx, key)
		if err != nil && err.Error() != "redis: nil" {
			cfg.Error("failed to get rate limit count", zap.Error(err))
			// Allow request if Redis fails
			c.Next(ctx)
			return
		}

		// Parse count
		count := 0
		if countStr != "" {
			fmt.Sscanf(countStr, "%d", &count)
		}

		// Check if rate limit exceeded
		if count >= maxRequests {
			cfg.Warn("rate limit exceeded",
				zap.String("client_ip", clientIP),
				zap.Int("count", count),
				zap.Int("max", maxRequests),
			)
			c.JSON(consts.StatusTooManyRequests, map[string]interface{}{
				"error":       "rate limit exceeded",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}

		// Increment count
		if err := redis.Incr(ctx, key); err != nil {
			cfg.Error("failed to increment rate limit count", zap.Error(err))
			// Allow request if Redis fails
			c.Next(ctx)
			return
		}

		// Set expiration if this is a new key
		if count == 0 {
			redis.Expire(ctx, key, window)
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", maxRequests-count-1))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))

		// Continue to next handler
		c.Next(ctx)
	}
}

// PerIPRateLimiter creates a rate limiter that limits per IP address
func PerIPRateLimiter(redis *cache.Redis, maxRequests int, window time.Duration, cfg *config.Config) app.HandlerFunc {
	return RateLimiter(redis, maxRequests, window, cfg)
}

// PerUserRateLimiter creates a rate limiter that limits per user ID
func PerUserRateLimiter(redis *cache.Redis, maxRequests int, window time.Duration, cfg *config.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// Get user ID
		userID := GetUserIDFromContext(c)
		if userID == "" {
			// Fall back to IP rate limiting
			PerIPRateLimiter(redis, maxRequests, window, cfg)(ctx, c)
			return
		}

		// Build rate limit key
		key := fmt.Sprintf("ratelimit:user:%s", userID)

		// Get current count
		countStr, err := redis.Get(ctx, key)
		if err != nil && err.Error() != "redis: nil" {
			cfg.Error("failed to get rate limit count", zap.Error(err))
			c.Next(ctx)
			return
		}

		// Parse count
		count := 0
		if countStr != "" {
			fmt.Sscanf(countStr, "%d", &count)
		}

		// Check if rate limit exceeded
		if count >= maxRequests {
			cfg.Warn("rate limit exceeded",
				zap.String("user_id", userID),
				zap.Int("count", count),
				zap.Int("max", maxRequests),
			)
			c.JSON(consts.StatusTooManyRequests, map[string]interface{}{
				"error":       "rate limit exceeded",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}

		// Increment count
		if err := redis.Incr(ctx, key); err != nil {
			cfg.Error("failed to increment rate limit count", zap.Error(err))
			c.Next(ctx)
			return
		}

		// Set expiration if this is a new key
		if count == 0 {
			redis.Expire(ctx, key, window)
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", maxRequests-count-1))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))

		c.Next(ctx)
	}
}
