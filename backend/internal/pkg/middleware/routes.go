package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/modules/rbac"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"go.uber.org/zap"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ExtractAndStoreRoutes extracts all routes and stores them as permissions
func ExtractAndStoreRoutes(h *server.Hertz, rbacService *rbac.Service, redis *cache.Redis, cfg *config.Config) error {
	// Get all routes
	routes := h.Routes()
	routeInfos := make([]rbac.RouteInfo, 0)
	seenRoutes := make(map[string]struct{})

	for _, r := range routes {
		routeKey := fmt.Sprintf("%s %s", r.Method, r.Path)

		// Skip duplicates
		if _, exists := seenRoutes[routeKey]; exists {
			hlog.Warnf("Skipping duplicate route: %s", routeKey)
			continue
		}

		// Skip excluded routes
		if rbac.IsExcludedRoute(r.Path) {
			hlog.Infof("Route excluded: %s", routeKey)
			continue
		}

		hlog.Infof("Route registered: %s", routeKey)
		seenRoutes[routeKey] = struct{}{}

		routeInfo := rbac.RouteInfo{
			Name:   FormatRouteName(r.Path),
			URL:    r.Path,
			Guard:  "API",
			Method: r.Method,
		}
		routeInfos = append(routeInfos, routeInfo)
	}

	// Store permissions in database and Redis
	ctx := context.Background()
	if err := rbacService.StoreRoutePermissions(ctx, routeInfos); err != nil {
		cfg.Error("failed to store route permissions", zap.Error(err))
		return err
	}

	cfg.Info("route permissions stored successfully",
		zap.Int("count", len(routeInfos)),
	)

	return nil
}

// FormatRouteName formats a route path into a human-readable name
func FormatRouteName(path string) string {
	if path == "/ws" {
		return "WebSocket Connection"
	}
	cleaned := strings.Replace(path, "/api/v1", "", 1)
	cleaned = strings.ReplaceAll(cleaned, "/", " ")
	titleCaser := cases.Title(language.English)
	return titleCaser.String(strings.TrimSpace(cleaned))
}
