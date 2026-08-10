package pagination

import "time"

// PaginationSummary represents a summary of the pagination result
type PaginationSummary struct {
	TotalRecords    int64  `json:"total_records"`
	FilteredRecords int64  `json:"filtered_records"`
	Page            int    `json:"page"`
	PageSize        int    `json:"page_size"`
	TotalPages      int    `json:"total_pages"`
	HasMore         bool   `json:"has_more"`
	HasPrevious     bool   `json:"has_previous"`
	NextCursor      string `json:"next_cursor,omitempty"`
	PreviousCursor  string `json:"previous_cursor,omitempty"`
}

// ApplicationSummary represents application-specific summary data
type ApplicationSummary struct {
	TotalApplications int64            `json:"total_applications"`
	ByStatus          map[string]int64 `json:"by_status"`
	ByPriority        map[string]int64 `json:"by_priority"`
	RecentSubmissions int64            `json:"recent_submissions"`
	AvgProcessingTime float64          `json:"avg_processing_time_hours"`
}

// ApprovalSummary represents approval-specific summary data
type ApprovalSummary struct {
	TotalApprovals   int64            `json:"total_approvals"`
	ByStatus         map[string]int64 `json:"by_status"`
	PendingApprovals int64            `json:"pending_approvals"`
	AvgDecisionTime  float64          `json:"avg_decision_time_hours"`
	ApprovalRate     float64          `json:"approval_rate"`
}

// WorkflowSummary represents workflow-specific summary data
type WorkflowSummary struct {
	TotalWorkflows  int64            `json:"total_workflows"`
	ActiveWorkflows int64            `json:"active_workflows"`
	ByCategory      map[string]int64 `json:"by_category"`
	AvgStepsCount   float64          `json:"avg_steps_count"`
}

// NotificationSummary represents notification-specific summary data
type NotificationSummary struct {
	TotalNotifications int64            `json:"total_notifications"`
	ByType             map[string]int64 `json:"by_type"`
	ByChannel          map[string]int64 `json:"by_channel"`
	UnreadCount        int64            `json:"unread_count"`
	SentToday          int64            `json:"sent_today"`
}

// UserSummary represents user-specific summary data
type UserSummary struct {
	TotalUsers    int64            `json:"total_users"`
	ActiveUsers   int64            `json:"active_users"`
	ByStatus      map[string]int64 `json:"by_status"`
	RecentLogins  int64            `json:"recent_logins"`
	AvgLoginCount float64          `json:"avg_login_count"`
}

// RoleSummary represents role-specific summary data
type RoleSummary struct {
	TotalRoles      int64            `json:"total_roles"`
	WithPermissions int64            `json:"with_permissions"`
	ByDefault       map[string]int64 `json:"by_default"`
	AvgPermissions  float64          `json:"avg_permissions_per_role"`
}

// PermissionSummary represents permission-specific summary data
type PermissionSummary struct {
	TotalPermissions int64            `json:"total_permissions"`
	ByMethod         map[string]int64 `json:"by_method"`
	ByService        map[string]int64 `json:"by_service"`
	AssignedToRoles  int64            `json:"assigned_to_roles"`
}

// EscalationSummary represents escalation-specific summary data
type EscalationSummary struct {
	TotalEscalations    int64            `json:"total_escalations"`
	ResolvedEscalations int64            `json:"resolved_escalations"`
	PendingEscalations  int64            `json:"pending_escalations"`
	ByLevel             map[string]int64 `json:"by_level"`
	AvgResolutionTime   float64          `json:"avg_resolution_time_hours"`
}

// AuditLogSummary represents audit log-specific summary data
type AuditLogSummary struct {
	TotalLogs      int64            `json:"total_logs"`
	ByAction       map[string]int64 `json:"by_action"`
	ByEntity       map[string]int64 `json:"by_entity"`
	RecentActivity int64            `json:"recent_activity"`
	TodayLogs      int64            `json:"today_logs"`
}

// DateRange represents a date range for summary queries
type DateRange struct {
	Start time.Time
	End   time.Time
}

// NewDateRange creates a new date range
func NewDateRange(start, end *time.Time) DateRange {
	if start == nil {
		t := time.Time{}
		start = &t
	}
	if end == nil {
		now := time.Now()
		end = &now
	}
	return DateRange{Start: *start, End: *end}
}

// IsEmpty checks if the date range is empty
func (d DateRange) IsEmpty() bool {
	return d.Start.IsZero() && d.End.IsZero()
}
