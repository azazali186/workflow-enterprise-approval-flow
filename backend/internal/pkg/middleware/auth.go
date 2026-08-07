package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/pkg/auth"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"go.uber.org/zap"
)

// contextKey is a custom type for context keys
type contextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "userID"
	// UserEmailKey is the context key for user email
	UserEmailKey contextKey = "userEmail"
	// UserRolesKey is the context key for user roles
	UserRolesKey contextKey = "userRoles"
)

// AuthMiddleware creates a new authentication middleware
func AuthMiddleware(tokenService *auth.TokenService, cache *cache.Redis, cfg *config.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// Get Authorization header
		authorization := string(c.Request.Header.Peek("Authorization"))
		if len(authorization) < 8 || !strings.HasPrefix(authorization, "Bearer ") {
			c.JSON(consts.StatusUnauthorized, map[string]string{
				"error": "invalid or missing authorization header",
			})
			c.Abort()
			return
		}

		// Extract token
		tokenStr := authorization[7:]

		// Parse and validate token
		claims, err := tokenService.Validate(tokenStr)
		if err != nil {
			cfg.Error("invalid token", zap.Error(err))
			c.JSON(consts.StatusUnauthorized, map[string]string{
				"error": "invalid token",
			})
			c.Abort()
			return
		}

		// Check if token is in Redis (single sign-on)
		tokenHash := computeTokenHash(tokenStr, claims.UserID)
		tokenKey := "auth:token:" + claims.UserID
		cachedHash, err := cache.Get(ctx, tokenKey)
		if err != nil || cachedHash == "" {
			cfg.Error("token expired or invalidated", zap.String("user_id", claims.UserID))
			c.JSON(consts.StatusUnauthorized, map[string]string{
				"error": "token expired or invalidated",
			})
			c.Abort()
			return
		}

		// Verify token hash matches
		if cachedHash != tokenHash {
			cfg.Error("token invalidated (single sign-on)", zap.String("user_id", claims.UserID))
			c.JSON(consts.StatusUnauthorized, map[string]string{
				"error": "token invalidated (single sign-on)",
			})
			c.Abort()
			return
		}

		// Renew token if needed (less than 30 minutes remaining)
		ttl, err := cache.TTL(ctx, tokenKey)
		if err == nil && ttl.Seconds() < 1800 {
			if err := cache.Set(ctx, tokenKey, tokenHash, 5400); err != nil {
				cfg.Error("failed to renew token", zap.Error(err))
			}
		}

		// Set user info in context
		c.Set(string(UserIDKey), claims.UserID)
		c.Set(string(UserEmailKey), claims.Email)
		c.Set(string(UserRolesKey), claims.Roles)

		// Continue to next handler
		c.Next(ctx)
	}
}

// computeTokenHash computes a hash of the token and user ID for SSO validation
func computeTokenHash(token, userID string) string {
	// Simple hash for demonstration - in production use a more secure hash
	return fmt.Sprintf("%x", []byte(token+userID))
}

// GetUserIDFromContext retrieves the user ID from the context
func GetUserIDFromContext(c *app.RequestContext) string {
	if userID, exists := c.Get(string(UserIDKey)); exists {
		return fmt.Sprintf("%v", userID)
	}
	return ""
}

// GetUserEmailFromContext retrieves the user email from the context
func GetUserEmailFromContext(c *app.RequestContext) string {
	if email, exists := c.Get(string(UserEmailKey)); exists {
		return fmt.Sprintf("%v", email)
	}
	return ""
}

// GetUserRolesFromContext retrieves the user roles from the context
func GetUserRolesFromContext(c *app.RequestContext) []string {
	if roles, exists := c.Get(string(UserRolesKey)); exists {
		if roleList, ok := roles.([]string); ok {
			return roleList
		}
	}
	return nil
}
