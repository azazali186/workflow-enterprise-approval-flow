package pagination

import (
	"fmt"
	"strings"
)

// SortConfig represents a sort configuration
type SortConfig struct {
	Field     string
	Direction string
}

// AllowedSortFields defines the allowed sort fields for each entity
var AllowedSortFields = map[string][]string{
	"application":  {"created_at", "updated_at", "status", "priority", "submitted_at"},
	"approval":     {"created_at", "updated_at", "status", "decided_at", "escalation_level"},
	"workflow":     {"created_at", "updated_at", "name", "category", "version"},
	"notification": {"created_at", "updated_at", "type", "sent_at", "read_at"},
	"user":         {"created_at", "updated_at", "email", "name", "last_login_at"},
	"role":         {"created_at", "updated_at", "name"},
	"permission":   {"created_at", "updated_at", "name", "route", "method"},
	"escalation":   {"created_at", "updated_at", "level", "escalated_at", "resolved_at"},
	"audit_log":    {"created_at", "updated_at", "action"},
}

// DefaultSortFields defines the default sort fields for each entity
var DefaultSortFields = map[string]string{
	"application":  "created_at",
	"approval":     "created_at",
	"workflow":     "created_at",
	"notification": "created_at",
	"user":         "created_at",
	"role":         "created_at",
	"permission":   "created_at",
	"escalation":   "created_at",
	"audit_log":    "created_at",
}

// ParseSortConfig parses sort parameters into a SortConfig
func ParseSortConfig(sortBy, sortOrder, entityType string) SortConfig {
	// Get allowed fields for entity
	allowedFields, exists := AllowedSortFields[entityType]
	if !exists {
		allowedFields = AllowedSortFields["application"]
	}

	// Get default field
	defaultField := DefaultSortFields[entityType]
	if defaultField == "" {
		defaultField = "created_at"
	}

	// Validate sort field
	validSortBy := defaultField
	for _, field := range allowedFields {
		if sortBy == field {
			validSortBy = sortBy
			break
		}
	}

	// Validate sort order
	validSortOrder := "desc"
	if sortOrder == "asc" || sortOrder == "desc" {
		validSortOrder = sortOrder
	}

	return SortConfig{
		Field:     validSortBy,
		Direction: validSortOrder,
	}
}

// ApplySorting applies sorting to a GORM query (returns order clause string)
func ApplySorting(sortConfig SortConfig, allowedFields map[string]string) string {
	// Check if field is allowed
	if allowedFields != nil {
		if dbField, exists := allowedFields[sortConfig.Field]; exists {
			sortConfig.Field = dbField
		}
	}

	// Build order clause
	return fmt.Sprintf("%s %s", sortConfig.Field, strings.ToUpper(sortConfig.Direction))
}

// GetSortBySQL returns a SQL-safe sort clause
func GetSortBySQL(sortConfig SortConfig, tableAlias string) string {
	if tableAlias != "" {
		return fmt.Sprintf("%s.%s %s", tableAlias, sortConfig.Field, strings.ToUpper(sortConfig.Direction))
	}
	return fmt.Sprintf("%s %s", sortConfig.Field, strings.ToUpper(sortConfig.Direction))
}

// ColumnMapping maps API field names to database column names
var ColumnMapping = map[string]string{
	"id":               "id",
	"created_at":       "created_at",
	"updated_at":       "updated_at",
	"deleted_at":       "deleted_at",
	"name":             "name",
	"email":            "email",
	"status":           "status",
	"priority":         "priority",
	"type":             "type",
	"channel":          "channel",
	"title":            "title",
	"route":            "route",
	"path":             "path",
	"method":           "method",
	"service":          "service",
	"action":           "action",
	"level":            "level",
	"escalated_at":     "escalated_at",
	"resolved_at":      "resolved_at",
	"decided_at":       "decided_at",
	"submitted_at":     "submitted_at",
	"completed_at":     "completed_at",
	"last_login_at":    "last_login_at",
	"category":         "category",
	"version":          "version",
	"is_active":        "is_active",
	"is_default":       "is_default",
	"is_internal":      "is_internal",
	"is_required":      "is_required",
	"step_order":       "step_order",
	"escalation_level": "escalation_level",
}

// MapFieldName maps an API field name to a database column name
func MapFieldName(fieldName string) string {
	if dbField, exists := ColumnMapping[fieldName]; exists {
		return dbField
	}
	return fieldName
}
