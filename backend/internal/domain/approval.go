package domain

import (
	"time"

	"github.com/google/uuid"
)

// Approval represents an approval record in the system
type Approval struct {
	Base
	ApplicationID   uuid.UUID  `gorm:"type:uuid;index" json:"application_id"`
	WorkflowStepID  uuid.UUID  `gorm:"type:uuid;index" json:"workflow_step_id"`
	ApproverID      uuid.UUID  `gorm:"type:uuid;index" json:"approver_id"`
	Status          string     `gorm:"size:50;index" json:"status"`
	Decision        *string    `gorm:"size:50" json:"decision,omitempty"`
	Comment         *string    `gorm:"type:text" json:"comment,omitempty"`
	DecidedAt       *time.Time `json:"decided_at,omitempty"`
	EscalationLevel int        `gorm:"default:0" json:"escalation_level"`
	Metadata        JSONMap    `gorm:"type:jsonb" json:"metadata,omitempty"`
}
