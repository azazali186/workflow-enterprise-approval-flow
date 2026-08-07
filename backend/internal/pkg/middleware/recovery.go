package middleware

import (
	"context"
	"runtime/debug"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"go.uber.org/zap"
)

// Recovery creates a new recovery middleware
func Recovery(cfg *config.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if r := recover(); r != nil {
				// Get stack trace
				stack := debug.Stack()

				// Log the panic
				cfg.Error("panic recovered",
					zap.Any("error", r),
					zap.String("method", string(c.Request.Method())),
					zap.String("path", string(c.Request.URI().Path())),
					zap.String("query", string(c.Request.URI().QueryString())),
					zap.String("client_ip", c.ClientIP()),
					zap.String("stack", string(stack)),
				)

				// Return 500 error
				c.JSON(consts.StatusInternalServerError, map[string]string{
					"error": "internal server error",
				})
				c.Abort()
			}
		}()

		// Continue to next handler
		c.Next(ctx)
	}
}

// RecoveryWithLogger creates a recovery middleware with custom logger
func RecoveryWithLogger(serviceName, env string, cfg *config.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if r := recover(); r != nil {
				// Get stack trace
				stack := debug.Stack()

				// Log the panic with context
				cfg.Error("panic recovered",
					zap.String("service", serviceName),
					zap.String("env", env),
					zap.Any("error", r),
					zap.String("method", string(c.Request.Method())),
					zap.String("path", string(c.Request.URI().Path())),
					zap.String("query", string(c.Request.URI().QueryString())),
					zap.String("client_ip", c.ClientIP()),
					zap.String("request_id", GetRequestIDFromContext(c)),
					zap.String("stack", string(stack)),
				)

				// Return 500 error
				c.JSON(consts.StatusInternalServerError, map[string]string{
					"error": "internal server error",
				})
				c.Abort()
			}
		}()

		// Continue to next handler
		c.Next(ctx)
	}
}

// RecoveryWithMiddleware creates a recovery middleware that logs request details
func RecoveryWithMiddleware(cfg *config.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if r := recover(); r != nil {
				// Get stack trace
				stack := debug.Stack()

				// Log the panic with full context
				cfg.Error("panic recovered",
					zap.Any("error", r),
					zap.String("method", string(c.Request.Method())),
					zap.String("path", string(c.Request.URI().Path())),
					zap.String("query", string(c.Request.URI().QueryString())),
					zap.String("client_ip", c.ClientIP()),
					zap.String("request_id", GetRequestIDFromContext(c)),
					zap.String("user_agent", string(c.Request.Header.Peek("User-Agent"))),
					zap.String("stack", string(stack)),
				)

				// Return 500 error
				c.JSON(consts.StatusInternalServerError, map[string]string{
					"error": "internal server error",
				})
				c.Abort()
			}
		}()

		// Continue to next handler
		c.Next(ctx)
	}
}

// Logger creates a request logging middleware
func Logger(cfg *config.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// Continue to next handler
		c.Next(ctx)

		// Log after handler completes
		cfg.Info("request completed",
			zap.String("method", string(c.Request.Method())),
			zap.String("path", string(c.Request.URI().Path())),
			zap.Int("status", c.Response.StatusCode()),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", GetRequestIDFromContext(c)),
		)
	}
}

// SecurityHeaders adds security headers to responses
func SecurityHeaders() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// Set security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Continue to next handler
		c.Next(ctx)
	}
}
