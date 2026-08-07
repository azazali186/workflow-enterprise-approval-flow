package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	pkguuid "github.com/aeroxe/approval-flow/internal/pkg/uuid"
)

const (
	// RequestIDHeader is the header name for request ID
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey is the context key for request ID
	RequestIDKey = "requestID"
)

// RequestID creates a new request ID middleware
func RequestID() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// Check if request ID already exists in header
		requestID := string(c.Request.Header.Peek(RequestIDHeader))

		// Generate new request ID if not provided
		if requestID == "" {
			requestID = pkguuid.GenerateID()
		}

		// Set request ID in context
		c.Set(RequestIDKey, requestID)

		// Set response header
		c.Header(RequestIDHeader, requestID)

		// Continue to next handler
		c.Next(ctx)
	}
}

// GetRequestIDFromContext retrieves the request ID from the context
func GetRequestIDFromContext(c *app.RequestContext) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		return requestID.(string)
	}
	return ""
}
