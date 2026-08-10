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

// runWithTimeout runs the remaining handler chain with a cancelable context
// and reports whether it timed out. On timeout the context is canceled, so
// downstream handlers that respect context cancellation (DB/Redis calls with
// ctx) abort their work instead of running to completion after the client
// already received a 408 — the previous implementation passed the parent
// context to c.Next, leaking goroutines under slow upstreams.
//
// Panics raised by handlers in the spawned goroutine are forwarded back to
// the middleware goroutine and re-raised there, so the Recovery middleware
// (wrapping this chain in the outer goroutine) still catches them. Without
// this, a handler panic would crash the entire process.
func runWithTimeout(ctx context.Context, c *app.RequestContext, cfg *config.Config, timeout time.Duration, message string) bool {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan struct{})
	panicCh := make(chan any, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Log immediately so a crash-inducing bug is visible even when
				// the timeout branch wins the select below and drops the panic.
				cfg.Error("panic recovered in timed-out handler",
					zap.Any("error", r),
					zap.String("method", string(c.Request.Method())),
					zap.String("path", string(c.Request.URI().Path())),
				)
				panicCh <- r
			}
			close(done)
		}()
		// Execute the chain with the timeout context so cancellation propagates.
		c.Next(timeoutCtx)
	}()

	select {
	case <-done:
		// Request completed within timeout. Re-raise any handler panic on this
		// goroutine so the surrounding Recovery middleware handles it.
		select {
		case r := <-panicCh:
			panic(r)
		default:
		}
		return false
	case <-timeoutCtx.Done():
		// Timeout occurred: cancel the handler goroutine's work and respond 408.
		cancel()
		cfg.Warn("request timeout",
			zap.String("method", string(c.Request.Method())),
			zap.String("path", string(c.Request.URI().Path())),
			zap.Duration("timeout", timeout),
		)

		// Only send response if not already written.
		if !c.Response.IsBodyStream() {
			c.JSON(consts.StatusRequestTimeout, map[string]string{
				"error":   "timeout",
				"message": message,
			})
		}
		return true
	}
}

// TimeoutMiddleware creates a middleware that enforces request timeouts
func TimeoutMiddleware(cfg *config.Config) app.HandlerFunc {
	config := DefaultTimeoutConfig()
	return func(ctx context.Context, c *app.RequestContext) {
		runWithTimeout(ctx, c, cfg, config.Timeout, config.Message)
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
		timedOut := runWithTimeout(ctx, c, cfg, timeoutConfig.Timeout, timeoutConfig.Message)

		// Custom handler fires only after a timeout.
		if timedOut && timeoutConfig.Handler != nil {
			timeoutConfig.Handler(ctx, c)
		}
	}
}
