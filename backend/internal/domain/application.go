package domain

import (
	"time"

	"github.com/google/uuid"
)

// Application represents an application submission in the system
type Application struct {
	Base
	ApplicantID uuid.UUID  `gorm:"type:uuid;index" json:"applicant_id"`
	WorkflowID  uuid.UUID  `gorm:"type:uuid;index" json:"workflow_id"`
	TemplateID  uuid.UUID  `gorm:"type:uuid;index" json:"template_id"`
	Status      string     `gorm:"size:50;index;default:draft" json:"status"`
	Priority    string     `gorm:"size:50;default:medium" json:"priority"`
	SubmittedAt *time.Time `json:"submitted_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Data        JSONMap    `gorm:"type:jsonb" json:"data,omitempty"`
}
