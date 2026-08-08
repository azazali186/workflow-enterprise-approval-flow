package validation

import "github.com/google/uuid"

// ==================== Auth Requests ====================

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LogoutRequest and GetProfileRequest accept an optional user_id for backward
// compatibility, but the identity always comes from the authenticated token.
type LogoutRequest struct {
	UserID string `json:"user_id"`
}

type GetProfileRequest struct {
	UserID string `json:"user_id"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ==================== User Requests ====================

type ListUsersRequest struct {
	Cursor    string `json:"cursor"`
	Limit     int    `json:"limit"`
	Search    string `json:"search"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
}

type GetUserRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type UpdateUserRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

type DeleteUserRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// ==================== Role Requests ====================

type ListRolesRequest struct {
	Cursor    string `json:"cursor"`
	Limit     int    `json:"limit"`
	Search    string `json:"search"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
}

type GetRoleRequest struct {
	RoleID string `json:"role_id" binding:"required"`
}

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"max=500"`
	IsDefault   bool   `json:"is_default"`
}

type UpdateRoleRequest struct {
	RoleID      string `json:"role_id" binding:"required"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDefault   *bool  `json:"is_default"`
}

type DeleteRoleRequest struct {
	RoleID string `json:"role_id" binding:"required"`
}

// ==================== Permission Requests ====================

type ListPermissionsRequest struct {
	Cursor    string `json:"cursor"`
	Limit     int    `json:"limit"`
	Search    string `json:"search"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
}

type GetPermissionRequest struct {
	PermissionID string `json:"permission_id" binding:"required"`
}

type CreatePermissionRequest struct {
	Name    string `json:"name" binding:"required"`
	Route   string `json:"route" binding:"required"`
	Path    string `json:"path" binding:"required"`
	Method  string `json:"method" binding:"required,oneof=POST PATCH DELETE"`
	Service string `json:"service" binding:"required"`
}

type UpdatePermissionRequest struct {
	PermissionID string `json:"permission_id" binding:"required"`
	Name         string `json:"name"`
	Route        string `json:"route"`
	Path         string `json:"path"`
	Method       string `json:"method"`
	Service      string `json:"service"`
}

type DeletePermissionRequest struct {
	PermissionID string `json:"permission_id" binding:"required"`
}

// ==================== Role-Permission Requests ====================

type GetRolePermissionsRequest struct {
	RoleID string `json:"role_id" binding:"required"`
}

type AssignPermissionToRoleRequest struct {
	RoleID       string `json:"role_id" binding:"required"`
	PermissionID string `json:"permission_id" binding:"required"`
}

type RemovePermissionFromRoleRequest struct {
	RoleID       string `json:"role_id" binding:"required"`
	PermissionID string `json:"permission_id" binding:"required"`
}

// ==================== User-Role Requests ====================

type GetUserRolesRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type AssignRoleToUserRequest struct {
	UserID string `json:"user_id" binding:"required"`
	RoleID string `json:"role_id" binding:"required"`
}

type RemoveRoleFromUserRequest struct {
	UserID string `json:"user_id" binding:"required"`
	RoleID string `json:"role_id" binding:"required"`
}

// ==================== Application Requests ====================

type SubmitApplicationRequest struct {
	ApplicantID string                 `json:"applicant_id" binding:"required"`
	WorkflowID  string                 `json:"workflow_id" binding:"required"`
	TemplateID  string                 `json:"template_id" binding:"required"`
	Title       string                 `json:"title" binding:"required"`
	Description string                 `json:"description"`
	Priority    string                 `json:"priority" binding:"required,oneof=low medium high urgent"`
	Data        map[string]interface{} `json:"data"`
}

type GetApplicationRequest struct {
	ApplicationID string `json:"application_id" binding:"required"`
}

type ListApplicationsRequest struct {
	Cursor      string `json:"cursor"`
	Limit       int    `json:"limit"`
	Search      string `json:"search"`
	SortBy      string `json:"sort_by"`
	SortOrder   string `json:"sort_order"`
	Status      string `json:"status"`
	Priority    string `json:"priority"`
	ApplicantID string `json:"applicant_id"`
	WorkflowID  string `json:"workflow_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
}

type UpdateApplicationRequest struct {
	ApplicationID string                 `json:"application_id" binding:"required"`
	Status        string                 `json:"status"`
	Priority      string                 `json:"priority"`
	Data          map[string]interface{} `json:"data"`
}

type DeleteApplicationRequest struct {
	ApplicationID string `json:"application_id" binding:"required"`
}

// ==================== Approval Requests ====================

type CreateApprovalRequest struct {
	ApplicationID  string `json:"application_id" binding:"required"`
	WorkflowStepID string `json:"workflow_step_id" binding:"required"`
	ApproverID     string `json:"approver_id" binding:"required"`
}

type GetApprovalRequest struct {
	ApprovalID string `json:"approval_id" binding:"required"`
}

type ListApprovalsRequest struct {
	Cursor        string `json:"cursor"`
	Limit         int    `json:"limit"`
	Search        string `json:"search"`
	SortBy        string `json:"sort_by"`
	SortOrder     string `json:"sort_order"`
	Status        string `json:"status"`
	Decision      string `json:"decision"`
	ApproverID    string `json:"approver_id"`
	ApplicationID string `json:"application_id"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
}

type DecideApprovalRequest struct {
	ApprovalID string `json:"approval_id" binding:"required"`
	Decision   string `json:"decision" binding:"required,oneof=approved rejected escalated"`
	Comment    string `json:"comment"`
}

type UpdateApprovalRequest struct {
	ApprovalID string `json:"approval_id" binding:"required"`
	Status     string `json:"status"`
	Comment    string `json:"comment"`
}

type DeleteApprovalRequest struct {
	ApprovalID string `json:"approval_id" binding:"required"`
}

type GetPendingApprovalsRequest struct {
	ApproverID string `json:"approver_id" binding:"required"`
	Cursor     string `json:"cursor"`
	Limit      int    `json:"limit"`
	SortBy     string `json:"sort_by"`
	SortOrder  string `json:"sort_order"`
}

// ==================== Workflow Requests ====================

type CreateWorkflowRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Category    string `json:"category" binding:"required"`
	IsActive    bool   `json:"is_active"`
}

type GetWorkflowRequest struct {
	WorkflowID string `json:"workflow_id" binding:"required"`
}

type ListWorkflowsRequest struct {
	Cursor    string `json:"cursor"`
	Limit     int    `json:"limit"`
	Search    string `json:"search"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
	Category  string `json:"category"`
	IsActive  *bool  `json:"is_active"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type UpdateWorkflowRequest struct {
	WorkflowID  string `json:"workflow_id" binding:"required"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	IsActive    *bool  `json:"is_active"`
}

type DeleteWorkflowRequest struct {
	WorkflowID string `json:"workflow_id" binding:"required"`
}

// ==================== Template Requests ====================

type CreateTemplateRequest struct {
	Name     string                 `json:"name" binding:"required"`
	Category string                 `json:"category" binding:"required"`
	Schema   map[string]interface{} `json:"schema"`
	UI       map[string]interface{} `json:"ui"`
}

type GetTemplateRequest struct {
	TemplateID string `json:"template_id" binding:"required"`
}

type ListTemplatesRequest struct {
	Cursor    string `json:"cursor"`
	Limit     int    `json:"limit"`
	Search    string `json:"search"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
	Category  string `json:"category"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type UpdateTemplateRequest struct {
	TemplateID string                 `json:"template_id" binding:"required"`
	Name       string                 `json:"name"`
	Category   string                 `json:"category"`
	Schema     map[string]interface{} `json:"schema"`
	UI         map[string]interface{} `json:"ui"`
}

type DeleteTemplateRequest struct {
	TemplateID string `json:"template_id" binding:"required"`
}

// ==================== Escalation Requests ====================

type CreateEscalationRequest struct {
	ApprovalID  string `json:"approval_id" binding:"required"`
	Level       int    `json:"level" binding:"min=0,max=10"`
	EscalatedTo string `json:"escalated_to" binding:"required"`
	Reason      string `json:"reason" binding:"required"`
}

type GetEscalationRequest struct {
	EscalationID string `json:"escalation_id" binding:"required"`
}

type ListEscalationsRequest struct {
	Cursor     string `json:"cursor"`
	Limit      int    `json:"limit"`
	Search     string `json:"search"`
	SortBy     string `json:"sort_by"`
	SortOrder  string `json:"sort_order"`
	ApprovalID string `json:"approval_id"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
}

type ResolveEscalationRequest struct {
	EscalationID string `json:"escalation_id" binding:"required"`
}

type GetEscalationsByApprovalRequest struct {
	ApprovalID string `json:"approval_id" binding:"required"`
	Cursor     string `json:"cursor"`
	Limit      int    `json:"limit"`
	SortBy     string `json:"sort_by"`
	SortOrder  string `json:"sort_order"`
}

// ==================== Notification Requests ====================

type SendNotificationRequest struct {
	UserID  string                 `json:"user_id" binding:"required"`
	Type    string                 `json:"type" binding:"required"`
	Channel string                 `json:"channel" binding:"required"`
	Title   string                 `json:"title" binding:"required"`
	Body    string                 `json:"body" binding:"required"`
	Data    map[string]interface{} `json:"data"`
}

type GetNotificationsRequest struct {
	UserID    string `json:"user_id"` // optional; the authenticated user is used
	Cursor    string `json:"cursor"`
	Limit     int    `json:"limit"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
	Type      string `json:"type"`
	Channel   string `json:"channel"`
	IsRead    *bool  `json:"is_read"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type GetUnreadNotificationsRequest struct {
	UserID string `json:"user_id"` // optional; the authenticated user is used
	Cursor string `json:"cursor"`
	Limit  int    `json:"limit"`
}

type MarkNotificationReadRequest struct {
	NotificationID string `json:"notification_id" binding:"required"`
}

type GetNotificationStatsRequest struct {
	UserID string `json:"user_id"` // optional; the authenticated user is used
}

// ==================== Report Requests ====================

type GetStatusesRequest struct {
	EntityType string `json:"entity_type" binding:"required"`
}

type GetCommentsRequest struct {
	ApprovalID string `json:"approval_id" binding:"required"`
	Cursor     string `json:"cursor"`
	Limit      int    `json:"limit"`
	SortBy     string `json:"sort_by"`
	SortOrder  string `json:"sort_order"`
}

type GetDocumentsRequest struct {
	ApplicationID string `json:"application_id" binding:"required"`
	Cursor        string `json:"cursor"`
	Limit         int    `json:"limit"`
	SortBy        string `json:"sort_by"`
	SortOrder     string `json:"sort_order"`
}

// ==================== Analytics Requests ====================

type GetApprovalStatsRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type GetWorkflowPerformanceRequest struct {
	WorkflowID string `json:"workflow_id" binding:"required"`
}

type GetApproverPerformanceRequest struct {
	ApproverID string `json:"approver_id" binding:"required"`
}

type GetEscalationMetricsRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// ==================== Pagination Request ====================

type PaginationRequest struct {
	Cursor    string `json:"cursor"`
	Limit     int    `json:"limit"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
	Search    string `json:"search"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// Validate sets defaults for pagination request
func (p *PaginationRequest) Validate() {
	if p.Limit <= 0 {
		p.Limit = 10
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
	if p.SortBy == "" {
		p.SortBy = "created_at"
	}
	if p.SortOrder == "" {
		p.SortOrder = "desc"
	}
}

// ==================== Login Log Requests ====================

type GetLoginHistoryRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Limit  int    `json:"limit"`
}

type GetLoginHistoryByEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	Limit int    `json:"limit"`
}

type GetLoginStatsRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type GetLoginLogsRequest struct {
	Email     string `json:"email"`
	Status    string `json:"status"`
	UserID    string `json:"user_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Page      int    `json:"page"`
	Limit     int    `json:"limit"`
}

// ==================== Audit Log Requests ====================

type GetAuditLogsRequest struct {
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
}

type GetEntityHistoryRequest struct {
	EntityType string `json:"entity_type" binding:"required"`
	EntityID   string `json:"entity_id" binding:"required"`
}

type GetUserActivityRequest struct {
	ActorID string `json:"actor_id" binding:"required"`
	Limit   int    `json:"limit"`
}

// ==================== Helper: Parse UUID ====================

func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
