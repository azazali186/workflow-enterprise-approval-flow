package middleware

import (
	"context"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

const (
	// MaxBodySize is the maximum request body size (10MB)
	MaxBodySize = 10 * 1024 * 1024
)

// BodySizeLimit creates a middleware that limits request body size
func BodySizeLimit(maxSize int64, cfg *config.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// Check Content-Length header
		contentLength := int64(c.Request.Header.ContentLength())
		if contentLength > maxSize {
			c.JSON(consts.StatusRequestEntityTooLarge, map[string]string{
				"error": "request body too large",
			})
			c.Abort()
			return
		}

		// Limit the request body reader
		c.Request.SetBodyStream(c.Request.BodyStream(), int(maxSize))

		c.Next(ctx)
	}
}
