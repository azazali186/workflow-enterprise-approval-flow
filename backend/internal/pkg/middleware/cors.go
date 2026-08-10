package middleware

import (
	"context"
	"strconv"
	"strings"

	"github.com/aeroxe/approval-flow/internal/config"
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

// DefaultCORSConfig returns the default (development) CORS configuration.
// Note: with a "*" allow-list, credentials cannot be enabled — browsers reject
// the combination, and reflecting arbitrary origins with credentials would be
// insecure. Use NewCORSConfig with an explicit allow-list in production.
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           43200, // 12 hours
	}
}

// EffectiveCORSOrigins returns the effective origin allow-list for a config:
// an empty or nil configured list falls back to "*" (development default).
// Both the CORS middleware and the WebSocket origin check must use this so
// they never disagree (an empty list previously allowed API requests via the
// middleware fallback while rejecting every browser WebSocket handshake).
func EffectiveCORSOrigins(configured []string) []string {
	if len(configured) == 0 {
		return []string{"*"}
	}
	return configured
}

// NewCORSConfig builds a CORS configuration from the app config.
// An empty or nil allow-list falls back to "*" (development default).
// A wildcard origin ("*") disables credentials; an explicit allow-list enables them.
func NewCORSConfig(cfg *config.Config) *CORSConfig {
	c := DefaultCORSConfig()
	c.AllowOrigins = EffectiveCORSOrigins(cfg.CORSAllowedOrigins)

	hasWildcard := false
	for _, origin := range c.AllowOrigins {
		if origin == "*" {
			hasWildcard = true
			break
		}
	}
	c.AllowCredentials = !hasWildcard && len(c.AllowOrigins) > 0
	return c
}

// CORS creates a new CORS middleware
func CORS(config *CORSConfig) app.HandlerFunc {
	if config == nil {
		config = DefaultCORSConfig()
	}

	return func(ctx context.Context, c *app.RequestContext) {
		// Get origin
		origin := string(c.Request.Header.Peek("Origin"))
		if origin == "" {
			c.Next(ctx)
			return
		}

		// Check if origin is allowed
		if !IsOriginAllowed(origin, config.AllowOrigins) {
			// Not allowed: do not set CORS headers; let the request proceed
			// (the browser will block the response for cross-origin reads).
			c.Next(ctx)
			return
		}

		// Set CORS headers
		if originIsWildcard(config.AllowOrigins) {
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
				c.Header("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
			}
			c.AbortWithStatus(consts.StatusNoContent)
			return
		}

		c.Next(ctx)
	}
}

// IsOriginAllowed checks whether an origin matches an allow-list.
// Supports a literal "*" wildcard and "*.example.com" subdomain wildcards.
func IsOriginAllowed(origin string, allowOrigins []string) bool {
	for _, allowedOrigin := range allowOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
		// Check for wildcard subdomain
		if strings.HasPrefix(allowedOrigin, "*.") {
			suffix := allowedOrigin[1:] // e.g., ".example.com"
			if strings.HasSuffix(origin, suffix) {
				return true
			}
		}
	}
	return false
}

// originIsWildcard reports whether the allow-list is a single "*" entry.
func originIsWildcard(allowOrigins []string) bool {
	return len(allowOrigins) == 1 && allowOrigins[0] == "*"
}
