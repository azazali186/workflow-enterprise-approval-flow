package domain

import "github.com/google/uuid"

// AuditLog represents an audit log entry
type AuditLog struct {
	Base
	EntityType   string     `gorm:"size:100;index" json:"entity_type"`
	EntityID     uuid.UUID  `gorm:"type:uuid;index" json:"entity_id"`
	Action       string     `gorm:"size:100;index" json:"action"`
	ActorID      *uuid.UUID `gorm:"type:uuid" json:"actor_id,omitempty"`
	ActorEmail   *string    `gorm:"size:255" json:"actor_email,omitempty"`
	Changes      JSONMap    `gorm:"type:jsonb" json:"changes,omitempty"`
	BeforeState  JSONMap    `gorm:"type:jsonb" json:"before_state,omitempty"`
	AfterState   JSONMap    `gorm:"type:jsonb" json:"after_state,omitempty"`
	IPAddress    *string    `gorm:"size:45" json:"ip_address,omitempty"`
	UserAgent    *string    `gorm:"type:text" json:"user_agent,omitempty"`
	RequestID    *string    `gorm:"size:100" json:"request_id,omitempty"`
	Status       string     `gorm:"size:20;default:success;index" json:"status"`
	ErrorMessage *string    `gorm:"type:text" json:"error_message,omitempty"`
	DurationMs   *int64     `json:"duration_ms,omitempty"`
}

// Audit log status constants
const (
	AuditStatusSuccess = "success"
	AuditStatusFailed  = "failed"
	AuditStatusError   = "error"
)

// Audit action constants
const (
	AuditActionCreate         = "create"
	AuditActionUpdate         = "update"
	AuditActionDelete         = "delete"
	AuditActionLogin          = "login"
	AuditActionLogout         = "logout"
	AuditActionLoginFailed    = "login_failed"
	AuditActionRegister       = "register"
	AuditActionSubmit         = "submit"
	AuditActionDecide         = "decide"
	AuditActionAssign         = "assign"
	AuditActionRemove         = "remove"
	AuditActionResolve        = "resolve"
	AuditActionMarkRead       = "mark_read"
	AuditActionSend           = "send"
	AuditActionRefreshToken   = "refresh_token"
	AuditActionChangePassword = "change_password"
)
