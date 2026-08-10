package middleware

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/pkg/auth"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/redis/go-redis/v9"
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

// TokenRenewThreshold is the remaining TTL (in seconds) below which the session
// cache entry is renewed. Must match the TTL used at login time.
const tokenRenewThreshold = 30 * time.Minute

// AuthMiddleware creates a new authentication middleware
func AuthMiddleware(tokenService *auth.TokenService, cache *cache.Redis, cfg *config.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		claims, ok := ValidateBearerToken(ctx, c, tokenService, cache, cfg)
		if !ok {
			return
		}

		// Set user info in context
		c.Set(string(UserIDKey), claims.UserID)
		c.Set(string(UserEmailKey), claims.Email)
		c.Set(string(UserRolesKey), claims.Roles)

		// Continue to next handler
		c.Next(ctx)
	}
}

// ValidateBearerToken validates the Authorization: Bearer header and, when
// Redis is reachable, verifies the session (SSO) token hash. It writes the
// appropriate 401 response and returns ok=false on failure.
//
// If Redis is unavailable, validation degrades to the stateless JWT signature
// check so the API stays up; session revocation only resumes once Redis
// recovers. This is a deliberate availability/security trade-off, logged
// loudly for operators.
func ValidateBearerToken(ctx context.Context, c *app.RequestContext, tokenService *auth.TokenService, cache *cache.Redis, cfg *config.Config) (*auth.Claims, bool) {
	// Get Authorization header
	authorization := string(c.Request.Header.Peek("Authorization"))
	if len(authorization) < 8 || !strings.HasPrefix(authorization, "Bearer ") {
		c.JSON(consts.StatusUnauthorized, map[string]string{
			"error": "invalid or missing authorization header",
		})
		c.Abort()
		return nil, false
	}

	// Extract token
	tokenStr := authorization[7:]

	// Parse and validate JWT (stateless). ValidateAccess also rejects refresh
	// tokens (Subject "refresh-token"): they are signed with the same key but
	// live for 7 days, so they must never be accepted at authenticated
	// endpoints — otherwise a stolen refresh token would grant API access
	// during the Redis fail-open window.
	claims, err := tokenService.ValidateAccess(tokenStr)
	if err != nil {
		cfg.Error("invalid token", zap.Error(err))
		c.JSON(consts.StatusUnauthorized, map[string]string{
			"error": "invalid token",
		})
		c.Abort()
		return nil, false
	}

	// Check session against Redis (single sign-on)
	tokenHash := auth.ComputeTokenHash(tokenStr, claims.UserID)
	tokenKey := "auth:token:" + claims.UserID
	cachedHash, err := cache.Get(ctx, tokenKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// Key genuinely missing: session was logged out or expired.
			c.JSON(consts.StatusUnauthorized, map[string]string{
				"error": "token expired or invalidated",
			})
			c.Abort()
			return nil, false
		}
		// Redis connection failure: degrade to stateless JWT validation.
		cfg.Warn("redis unavailable; falling back to stateless JWT validation",
			zap.String("user_id", claims.UserID),
			zap.Error(err),
		)
		return claims, true
	}

	if cachedHash == "" || cachedHash != tokenHash {
		cfg.Error("token invalidated (single sign-on)", zap.String("user_id", claims.UserID))
		c.JSON(consts.StatusUnauthorized, map[string]string{
			"error": "token invalidated (single sign-on)",
		})
		c.Abort()
		return nil, false
	}

	// Renew the session TTL if it is close to expiry (sliding session).
	ttl, err := cache.TTL(ctx, tokenKey)
	if err == nil && ttl.Seconds() < tokenRenewThreshold.Seconds() {
		expiry := tokenService.Expiry()
		if err := cache.Set(ctx, tokenKey, tokenHash, expiry); err != nil {
			cfg.Error("failed to renew token", zap.Error(err))
		}
	}

	return claims, true
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
