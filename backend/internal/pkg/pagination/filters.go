package pagination

import (
	"time"
)

// DateRangeFilter represents a date range filter
type DateRangeFilter struct {
	Start *time.Time `json:"start"`
	End   *time.Time `json:"end"`
}

// SearchFilter represents a search filter
type SearchFilter struct {
	Query  string   `json:"query"`
	Fields []string `json:"fields"`
}

// StatusFilter represents a status filter
type StatusFilter struct {
	Statuses []string `json:"statuses"`
}

// IDFilter represents an ID filter
type IDFilter struct {
	IDs []string `json:"ids"`
}

// CommonFilters represents common filter parameters
type CommonFilters struct {
	Search    *SearchFilter    `json:"search"`
	Status    *StatusFilter    `json:"status"`
	DateRange *DateRangeFilter `json:"date_range"`
	CreatedBy *IDFilter        `json:"created_by"`
	SortBy    string           `json:"sort_by"`
	SortOrder string           `json:"sort_order"`
}

// ValidateSortOrder validates and returns the sort order
func (f *CommonFilters) ValidateSortOrder() string {
	if f.SortOrder == "" || (f.SortOrder != "asc" && f.SortOrder != "desc") {
		return "desc"
	}
	return f.SortOrder
}

// ValidateSortBy validates and returns the sort field
func (f *CommonFilters) ValidateSortBy(allowedFields []string, defaultField string) string {
	if f.SortBy == "" {
		return defaultField
	}
	for _, field := range allowedFields {
		if f.SortBy == field {
			return f.SortBy
		}
	}
	return defaultField
}

// ApplicationFilters represents application-specific filters
type ApplicationFilters struct {
	CommonFilters
	ApplicantID string `json:"applicant_id"`
	WorkflowID  string `json:"workflow_id"`
	TemplateID  string `json:"template_id"`
	Priority    string `json:"priority"`
}

// ApprovalFilters represents approval-specific filters
type ApprovalFilters struct {
	CommonFilters
	ApproverID     string `json:"approver_id"`
	ApplicationID  string `json:"application_id"`
	WorkflowStepID string `json:"workflow_step_id"`
	Decision       string `json:"decision"`
}

// WorkflowFilters represents workflow-specific filters
type WorkflowFilters struct {
	CommonFilters
	Category string `json:"category"`
	IsActive *bool  `json:"is_active"`
}

// NotificationFilters represents notification-specific filters
type NotificationFilters struct {
	CommonFilters
	UserID  string `json:"user_id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	ReadAt  *bool  `json:"read_at"`
}

// UserFilters represents user-specific filters
type UserFilters struct {
	CommonFilters
	Email  string `json:"email"`
	Status string `json:"status"`
	RoleID string `json:"role_id"`
}

// RoleFilters represents role-specific filters
type RoleFilters struct {
	CommonFilters
	Name      string `json:"name"`
	IsDefault *bool  `json:"is_default"`
}

// PermissionFilters represents permission-specific filters
type PermissionFilters struct {
	CommonFilters
	Name    string `json:"name"`
	Route   string `json:"route"`
	Method  string `json:"method"`
	Service string `json:"service"`
}

// EscalationFilters represents escalation-specific filters
type EscalationFilters struct {
	CommonFilters
	ApprovalID  string `json:"approval_id"`
	EscalatedTo string `json:"escalated_to"`
	IsResolved  *bool  `json:"is_resolved"`
}
