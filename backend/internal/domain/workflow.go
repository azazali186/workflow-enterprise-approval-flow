package domain

import (
	"github.com/google/uuid"
)

// Workflow represents a workflow definition in the system
type Workflow struct {
	Base
	Name        string  `gorm:"size:255;not null" json:"name"`
	Description *string `gorm:"type:text" json:"description,omitempty"`
	Category    string  `gorm:"size:100;index" json:"category"`
	Version     int     `gorm:"default:1" json:"version"`
	IsActive    bool    `gorm:"default:true;index" json:"is_active"`
	Steps       JSONMap `gorm:"type:jsonb" json:"steps,omitempty"`
}

// WorkflowStep represents a step in a workflow
type WorkflowStep struct {
	Base
	WorkflowID   uuid.UUID  `gorm:"type:uuid;index" json:"workflow_id"`
	Name         string     `gorm:"size:255;not null" json:"name"`
	StepOrder    int        `gorm:"index" json:"step_order"`
	ApproverRole string     `gorm:"size:100;index" json:"approver_role"`
	ApproverID   *uuid.UUID `gorm:"type:uuid" json:"approver_id,omitempty"`
	Action       string     `gorm:"size:100" json:"action"`
	TimeoutHours int        `gorm:"default:24" json:"timeout_hours"`
	IsRequired   bool       `gorm:"default:true" json:"is_required"`
}
