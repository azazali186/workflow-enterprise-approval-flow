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

// BodySizeLimit creates a middleware that limits request body size.
//
// The Content-Length header is checked for a fast rejection; the hard cap for
// chunked/streamed bodies is enforced by the server itself via
// server.WithMaxRequestBodySize (see internal/server/router.go). A previous
// implementation here called SetBodyStream(BodyStream(), maxSize) which wiped
// the body of every request (BodyStream() is nil for ordinary JSON bodies),
// causing every body-based endpoint to fail binding.
func BodySizeLimit(maxSize int64, cfg *config.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		contentLength := int64(c.Request.Header.ContentLength())
		if contentLength > maxSize {
			c.JSON(consts.StatusRequestEntityTooLarge, map[string]string{
				"error": "request body too large",
			})
			c.Abort()
			return
		}

		c.Next(ctx)
	}
}
