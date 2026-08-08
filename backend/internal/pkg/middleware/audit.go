package middleware

import (
	"bytes"
	"context"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuditLogger provides audit logging functionality
type AuditLogger struct {
	db     *gorm.DB
	cfg    *config.Config
	logger *zap.Logger
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(db *gorm.DB, cfg *config.Config) *AuditLogger {
	return &AuditLogger{
		db:     db,
		cfg:    cfg,
		logger: cfg.Logger,
	}
}

// responseWriter wraps the Hertz response writer to capture the status code
type auditResponseWriter struct {
	statusCode int
	body       bytes.Buffer
	written    bool
}

// LogEntry represents an audit log entry
type LogEntry struct {
	EntityType string
	EntityID   uuid.UUID
	Action     string
	ActorID    *uuid.UUID
	Changes    domain.JSONMap
	IPAddress  string
	UserAgent  string
}

// Log writes an audit log entry to the database
func (al *AuditLogger) Log(ctx context.Context, entry LogEntry) {
	log := domain.AuditLog{
		EntityType: entry.EntityType,
		EntityID:   entry.EntityID,
		Action:     entry.Action,
		ActorID:    entry.ActorID,
		Changes:    entry.Changes,
		IPAddress:  &entry.IPAddress,
		UserAgent:  &entry.UserAgent,
	}

	if err := al.db.WithContext(ctx).Create(&log).Error; err != nil {
		al.logger.Error("failed to create audit log",
			zap.Error(err),
			zap.String("entity_type", entry.EntityType),
			zap.String("action", entry.Action),
		)
	}
}

// AuditMiddleware creates a middleware that logs all write operations
func AuditMiddleware(db *gorm.DB, cfg *config.Config) app.HandlerFunc {
	auditLogger := NewAuditLogger(db, cfg)

	return func(ctx context.Context, c *app.RequestContext) {
		// Only audit write operations (POST, PATCH, DELETE)
		method := string(c.Request.Method())
		if method != "POST" && method != "PATCH" && method != "DELETE" {
			c.Next(ctx)
			return
		}

		// Skip health checks, metrics, and swagger
		path := string(c.Request.URI().Path())
		if isAuditExcludedPath(path) {
			c.Next(ctx)
			return
		}

		// Get user ID from context (set by auth middleware)
		var actorID *uuid.UUID
		if userIDStr, exists := c.Get("user_id"); exists {
			if uid, err := uuid.Parse(userIDStr.(string)); err == nil {
				actorID = &uid
			}
		}

		start := time.Now()

		// Execute next handler
		c.Next(ctx)

		// Determine action from method
		action := determineAction(method, path)

		// Determine entity type from path
		entityType := determineEntityType(path)

		// Parse entity ID from response if available
		var entityID uuid.UUID
		// For now, use a nil UUID - in production, parse from request/response
		_ = entityID

		// Get IP address
		ipAddress := c.ClientIP()

		// Get user agent
		userAgent := string(c.Request.Header.UserAgent())

		// Create changes map
		changes := domain.JSONMap{
			"method":      method,
			"path":        path,
			"status_code": c.Response.StatusCode(),
			"duration_ms": time.Since(start).Milliseconds(),
		}

		// Log asynchronously to avoid blocking; detach from the request context so a
		// cancelled request does not silently drop the audit write.
		go auditLogger.Log(context.WithoutCancel(ctx), LogEntry{
			EntityType: entityType,
			EntityID:   entityID,
			Action:     action,
			ActorID:    actorID,
			Changes:    changes,
			IPAddress:  ipAddress,
			UserAgent:  userAgent,
		})

		// Log to application logs as well
		hlog.Info("audit",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", c.Response.StatusCode()),
			zap.Duration("duration", time.Since(start)),
		)
	}
}

// isAuditExcludedPath checks if a path should be excluded from audit logging
func isAuditExcludedPath(path string) bool {
	excludedPaths := []string{
		"/health",
		"/health/detailed",
		"/health/ready",
		"/health/live",
		"/metrics",
		"/docs",
		"/docs/swagger.json",
		"/version",
		"/ws",
	}

	for _, excluded := range excludedPaths {
		if path == excluded {
			return true
		}
	}
	return false
}

// determineAction determines the audit action from HTTP method
func determineAction(method, path string) string {
	switch method {
	case "POST":
		if contains(path, "/create") {
			return "create"
		}
		if contains(path, "/submit") {
			return "submit"
		}
		if contains(path, "/assign") {
			return "assign"
		}
		if contains(path, "/remove") {
			return "remove"
		}
		if contains(path, "/decide") {
			return "decide"
		}
		if contains(path, "/resolve") {
			return "resolve"
		}
		if contains(path, "/read") {
			return "mark_read"
		}
		return "update"
	case "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return "unknown"
	}
}

// determineEntityType determines the entity type from the request path
func determineEntityType(path string) string {
	entityMap := map[string]string{
		"/applications":  "application",
		"/approvals":     "approval",
		"/workflows":     "workflow",
		"/templates":     "template",
		"/escalations":   "escalation",
		"/notifications": "notification",
		"/admin/users":   "user",
		"/admin/roles":   "role",
		"/admin/permis":  "permission",
		"/reports":       "report",
		"/analytics":     "analytics",
	}

	for prefix, entityType := range entityMap {
		if contains(path, prefix) {
			return entityType
		}
	}

	return "unknown"
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
