package middleware

import (
	"context"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"go.uber.org/zap"
)

const (
	// DefaultRequestTimeout is the default timeout for requests
	DefaultRequestTimeout = 30 * time.Second

	// MaxRequestTimeout is the maximum allowed timeout
	MaxRequestTimeout = 5 * time.Minute

	// MinRequestTimeout is the minimum timeout
	MinRequestTimeout = 5 * time.Second
)

// TimeoutConfig holds timeout middleware configuration
type TimeoutConfig struct {
	Timeout time.Duration
	// Message is the message returned when a timeout occurs
	Message string
	// Handler is called when a timeout occurs (optional)
	Handler func(ctx context.Context, c *app.RequestContext)
}

// DefaultTimeoutConfig returns the default timeout configuration
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		Timeout: DefaultRequestTimeout,
		Message: "request timeout",
	}
}

// TimeoutMiddleware creates a middleware that enforces request timeouts
func TimeoutMiddleware(cfg *config.Config) app.HandlerFunc {
	config := DefaultTimeoutConfig()

	return func(ctx context.Context, c *app.RequestContext) {
		// Create a context with timeout
		timeoutCtx, cancel := context.WithTimeout(ctx, config.Timeout)
		defer cancel()

		// Channel to signal completion
		done := make(chan struct{})

		go func() {
			// Execute next handler
			c.Next(ctx)
			close(done)
		}()

		select {
		case <-done:
			// Request completed within timeout
			return
		case <-timeoutCtx.Done():
			// Timeout occurred
			cfg.Warn("request timeout",
				zap.String("method", string(c.Request.Method())),
				zap.String("path", string(c.Request.URI().Path())),
				zap.Duration("timeout", config.Timeout),
			)

			// Only send response if not already written
			if !c.Response.IsBodyStream() {
				c.JSON(consts.StatusRequestTimeout, map[string]string{
					"error":   "timeout",
					"message": config.Message,
				})
			}
		}
	}
}

// TimeoutWithConfig creates a timeout middleware with custom configuration
func TimeoutWithConfig(cfg *config.Config, timeoutConfig TimeoutConfig) app.HandlerFunc {
	// Validate and normalize timeout
	if timeoutConfig.Timeout <= 0 {
		timeoutConfig.Timeout = DefaultRequestTimeout
	}
	if timeoutConfig.Timeout > MaxRequestTimeout {
		timeoutConfig.Timeout = MaxRequestTimeout
	}
	if timeoutConfig.Timeout < MinRequestTimeout {
		timeoutConfig.Timeout = MinRequestTimeout
	}
	if timeoutConfig.Message == "" {
		timeoutConfig.Message = "request timeout"
	}

	return func(ctx context.Context, c *app.RequestContext) {
		// Create a context with timeout
		timeoutCtx, cancel := context.WithTimeout(ctx, timeoutConfig.Timeout)
		defer cancel()

		// Channel to signal completion
		done := make(chan struct{})

		go func() {
			// Execute next handler
			c.Next(ctx)
			close(done)
		}()

		select {
		case <-done:
			// Request completed within timeout
			return
		case <-timeoutCtx.Done():
			// Timeout occurred
			cfg.Warn("request timeout",
				zap.String("method", string(c.Request.Method())),
				zap.String("path", string(c.Request.URI().Path())),
				zap.Duration("timeout", timeoutConfig.Timeout),
			)

			// Only send response if not already written
			if !c.Response.IsBodyStream() {
				c.JSON(consts.StatusRequestTimeout, map[string]string{
					"error":   "timeout",
					"message": timeoutConfig.Message,
				})
			}

			// Call custom handler if provided
			if timeoutConfig.Handler != nil {
				timeoutConfig.Handler(timeoutCtx, c)
			}
		}
	}
}
