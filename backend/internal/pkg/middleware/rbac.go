package middleware

import (
	"context"
	"fmt"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/modules/rbac"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RBACMiddleware creates a new RBAC authorization middleware
func RBACMiddleware(rbacService *rbac.Service, cfg *config.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// Get user ID from context
		userIDStr := GetUserIDFromContext(c)
		if userIDStr == "" {
			c.JSON(consts.StatusUnauthorized, map[string]string{
				"error": "user not authenticated",
			})
			c.Abort()
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			cfg.Error("invalid user ID", zap.Error(err))
			c.JSON(consts.StatusUnauthorized, map[string]string{
				"error": "invalid user ID",
			})
			c.Abort()
			return
		}

		// Build the route permission key (METHOD /path)
		method := string(c.Request.Method())
		path := string(c.Request.URI().Path())
		routeKey := fmt.Sprintf("%s %s", method, path)

		// Check if route is excluded (public routes)
		if rbac.IsExcludedRoute(path) {
			c.Next(ctx)
			return
		}

		// Check permission
		hasPermission, err := rbacService.CheckPermission(ctx, userID, routeKey)
		if err != nil {
			cfg.Error("failed to check permission", zap.Error(err))
			c.JSON(consts.StatusInternalServerError, map[string]string{
				"error": "failed to check permission",
			})
			c.Abort()
			return
		}

		if !hasPermission {
			cfg.Warn("permission denied",
				zap.String("user_id", userIDStr),
				zap.String("route", routeKey),
			)
			c.JSON(consts.StatusForbidden, map[string]string{
				"error": "insufficient permissions",
			})
			c.Abort()
			return
		}

		// Continue to next handler
		c.Next(ctx)
	}
}

// RequireRole creates a middleware that requires a specific role
func RequireRole(roleName string, rbacService *rbac.Service, cfg *config.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// Get user roles from context
		roles := GetUserRolesFromContext(c)
		if len(roles) == 0 {
			c.JSON(consts.StatusForbidden, map[string]string{
				"error": "no roles assigned",
			})
			c.Abort()
			return
		}

		// Check if user has the required role
		for _, role := range roles {
			if role == roleName {
				c.Next(ctx)
				return
			}
		}

		cfg.Warn("role required",
			zap.String("required_role", roleName),
			zap.Strings("user_roles", roles),
		)
		c.JSON(consts.StatusForbidden, map[string]string{
			"error": fmt.Sprintf("role '%s' required", roleName),
		})
		c.Abort()
	}
}

// RequireAnyRole creates a middleware that requires any of the specified roles
func RequireAnyRole(roleNames []string, rbacService *rbac.Service, cfg *config.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// Get user roles from context
		roles := GetUserRolesFromContext(c)
		if len(roles) == 0 {
			c.JSON(consts.StatusForbidden, map[string]string{
				"error": "no roles assigned",
			})
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		roleSet := make(map[string]bool)
		for _, role := range roles {
			roleSet[role] = true
		}

		for _, requiredRole := range roleNames {
			if roleSet[requiredRole] {
				c.Next(ctx)
				return
			}
		}

		cfg.Warn("role required",
			zap.Strings("required_roles", roleNames),
			zap.Strings("user_roles", roles),
		)
		c.JSON(consts.StatusForbidden, map[string]string{
			"error": "insufficient role permissions",
		})
		c.Abort()
	}
}
