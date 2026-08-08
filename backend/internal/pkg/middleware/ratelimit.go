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

// RateLimiter creates a fixed-window rate limiter keyed by client IP.
// It uses a single atomic INCR + EXPIRE so concurrent requests cannot
// overshoot the limit. If Redis is unreachable the request is allowed through
// (fail-open) with a warning, so a cache outage cannot take down the API.
func RateLimiter(redis *cache.Redis, maxRequests int, window time.Duration, cfg *config.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if maxRequests <= 0 {
			c.Next(ctx)
			return
		}

		// Get client IP
		clientIP := c.ClientIP()

		// Build rate limit key
		key := fmt.Sprintf("ratelimit:%s", clientIP)

		// Atomically seed the window (SET NX with TTL), then increment. This
		// guarantees the TTL is always set exactly once per window — an EXPIRE
		// that fails can never leave a permanent counter behind.
		count, ok := redis.IncrWindow(ctx, key, window)
		if !ok {
			cfg.Warn("rate limiter unavailable; allowing request (fail-open)",
				zap.String("client_ip", clientIP),
			)
			c.Next(ctx)
			return
		}

		// Check if rate limit exceeded
		if count > int64(maxRequests) {
			cfg.Warn("rate limit exceeded",
				zap.String("client_ip", clientIP),
				zap.Int64("count", count),
				zap.Int("max", maxRequests),
			)
			c.JSON(consts.StatusTooManyRequests, map[string]interface{}{
				"error":       "rate limit exceeded",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", maxRequests-int(count)))
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

		// Atomically seed the window (SET NX with TTL), then increment.
		count, ok := redis.IncrWindow(ctx, key, window)
		if !ok {
			cfg.Warn("rate limiter unavailable; allowing request (fail-open)",
				zap.String("user_id", userID),
			)
			c.Next(ctx)
			return
		}

		// Check if rate limit exceeded
		if count > int64(maxRequests) {
			cfg.Warn("rate limit exceeded",
				zap.String("user_id", userID),
				zap.Int64("count", count),
				zap.Int("max", maxRequests),
			)
			c.JSON(consts.StatusTooManyRequests, map[string]interface{}{
				"error":       "rate limit exceeded",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", maxRequests-int(count)))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))

		c.Next(ctx)
	}
}
