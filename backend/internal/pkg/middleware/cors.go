package middleware

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           43200, // 12 hours
	}
}

// CORS creates a new CORS middleware
func CORS(config *CORSConfig) app.HandlerFunc {
	if config == nil {
		config = DefaultCORSConfig()
	}

	return func(ctx context.Context, c *app.RequestContext) {
		// Handle preflight OPTIONS request
		if string(c.Request.Method()) == "OPTIONS" {
			c.AbortWithStatus(consts.StatusNoContent)
			return
		}

		// Get origin
		origin := string(c.Request.Header.Peek("Origin"))
		if origin == "" {
			c.Next(ctx)
			return
		}

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range config.AllowOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
			// Check for wildcard subdomain
			if strings.HasPrefix(allowedOrigin, "*.") {
				suffix := allowedOrigin[1:] // e.g., ".example.com"
				if strings.HasSuffix(origin, suffix) {
					allowed = true
					break
				}
			}
		}

		if !allowed {
			c.Next(ctx)
			return
		}

		// Set CORS headers
		if config.AllowOrigins[0] == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}

		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if len(config.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
		}

		// For preflight requests
		if string(c.Request.Method()) == "OPTIONS" {
			if len(config.AllowMethods) > 0 {
				c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
			}
			if len(config.AllowHeaders) > 0 {
				c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
			}
			if config.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", strings.TrimSpace(string(rune(config.MaxAge+'0'))))
			}
			c.AbortWithStatus(consts.StatusNoContent)
			return
		}

		c.Next(ctx)
	}
}
