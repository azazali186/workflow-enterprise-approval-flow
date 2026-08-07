package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Approval struct {
	Base
	ApplicationID  uuid.UUID      `gorm:"type:uuid;index" json:"application_id"`
	WorkflowStepID uuid.UUID      `gorm:"type:uuid;index" json:"workflow_step_id"`
	ApproverID     uuid.UUID      `gorm:"type:uuid;index" json:"approver_id"`
	Status         string         `gorm:"size:50;index" json:"status"`
	Decision       *string        `gorm:"size:50" json:"decision,omitempty"`
	Comment        *string        `gorm:"type:text" json:"comment,omitempty"`
	DecidedAt      *time.Time     `json:"decided_at,omitempty"`
	EscalationLevel int           `gorm:"default:0" json:"escalation_level"`
	Metadata       JSONMap        `gorm:"type:jsonb" json:"metadata,omitempty"`
}

type Application struct {
	Base
	ApplicantID    uuid.UUID  `gorm:"type:uuid;index" json:"applicant_id"`
	WorkflowID     uuid.UUID  `gorm:"type:uuid;index" json:"workflow_id"`
	TemplateID     uuid.UUID  `gorm:"type:uuid;index" json:"template_id"`
	Status         string     `gorm:"size:50;index;default:draft" json:"status"`
	Priority       string     `gorm:"size:50;default:medium" json:"priority"`
	SubmittedAt    *time.Time `json:"submitted_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	Data           JSONMap    `gorm:"type:jsonb" json:"data,omitempty"`
}

type Workflow struct {
	Base
	Name         string     `gorm:"size:255;not null" json:"name"`
	Description  *string    `gorm:"type:text" json:"description,omitempty"`
	Category     string     `gorm:"size:100;index" json:"category"`
	Version      int        `gorm:"default:1" json:"version"`
	IsActive     bool       `gorm:"default:true;index" json:"is_active"`
	Steps        JSONMap    `gorm:"type:jsonb" json:"steps,omitempty"`
}

type WorkflowStep struct {
	Base
	WorkflowID   uuid.UUID `gorm:"type:uuid;index" json:"workflow_id"`
	Name         string    `gorm:"size:255;not null" json:"name"`
	StepOrder    int       `gorm:"index" json:"step_order"`
	ApproverRole string    `gorm:"size:100;index" json:"approver_role"`
	ApproverID   *uuid.UUID `gorm:"type:uuid" json:"approver_id,omitempty"`
	Action       string    `gorm:"size:100" json:"action"`
	TimeoutHours int       `gorm:"default:24" json:"timeout_hours"`
	IsRequired   bool      `gorm:"default:true" json:"is_required"`
}

type Notification struct {
	Base
	UserID      uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	Type        string    `gorm:"size:50;index" json:"type"`
	Channel     string    `gorm:"size:50" json:"channel"`
	Title       string    `gorm:"size:255" json:"title"`
	Body        string    `gorm:"type:text" json:"body"`
	Data        JSONMap   `gorm:"type:jsonb" json:"data,omitempty"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
	SentAt      *time.Time `json:"sent_at,omitempty"`
}

type Template struct {
	Base
	Name         string    `gorm:"size:255;not null" json:"name"`
	Category     string    `gorm:"size:100;index" json:"category"`
	Version      int       `gorm:"default:1" json:"version"`
	IsActive     bool      `gorm:"default:true;index" json:"is_active"`
	Schema       JSONMap   `gorm:"type:jsonb" json:"schema,omitempty"`
	UI           JSONMap   `gorm:"type:jsonb" json:"ui,omitempty"`
}

type Escalation struct {
	Base
	ApprovalID   uuid.UUID `gorm:"type:uuid;index" json:"approval_id"`
	Level        int       `gorm:"index" json:"level"`
	EscalatedTo  uuid.UUID `gorm:"type:uuid;index" json:"escalated_to"`
	Reason       string    `gorm:"type:text" json:"reason"`
	EscalatedAt  time.Time `gorm:"index" json:"escalated_at"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
}

type Status struct {
	Base
	EntityType string    `gorm:"size:100;index" json:"entity_type"`
	EntityID   uuid.UUID `gorm:"type:uuid;index" json:"entity_id"`
	State      string    `gorm:"size:50;index" json:"state"`
	Reason     *string   `gorm:"type:text" json:"reason,omitempty"`
}

type Comment struct {
	Base
	ApprovalID   uuid.UUID `gorm:"type:uuid;index" json:"approval_id"`
	UserID       uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	Content      string    `gorm:"type:text;not null" json:"content"`
	IsInternal   bool      `gorm:"default:false;index" json:"is_internal"`
}

type Document struct {
	Base
	ApplicationID uuid.UUID `gorm:"type:uuid;index" json:"application_id"`
	Name          string    `gorm:"size:255;not null" json:"name"`
	FileURL       string    `gorm:"type:text;not null" json:"file_url"`
	FileType      string    `gorm:"size:100" json:"file_type"`
	FileSize      int64     `json:"file_size"`
	UploadedBy    uuid.UUID `gorm:"type:uuid" json:"uploaded_by"`
}

type AuditLog struct {
	Base
	EntityType string      `gorm:"size:100;index" json:"entity_type"`
	EntityID   uuid.UUID   `gorm:"type:uuid;index" json:"entity_id"`
	Action     string      `gorm:"size:100;index" json:"action"`
	ActorID    *uuid.UUID  `gorm:"type:uuid" json:"actor_id,omitempty"`
	Changes    JSONMap     `gorm:"type:jsonb" json:"changes,omitempty"`
	IPAddress  *string     `gorm:"size:45" json:"ip_address,omitempty"`
	UserAgent  *string     `gorm:"type:text" json:"user_agent,omitempty"`
}
