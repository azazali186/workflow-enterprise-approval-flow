package domain

import (
	"time"

	"github.com/google/uuid"
)

// Escalation represents an escalation record in the system
type Escalation struct {
	Base
	ApprovalID  uuid.UUID  `gorm:"type:uuid;index" json:"approval_id"`
	Level       int        `gorm:"index" json:"level"`
	EscalatedTo uuid.UUID  `gorm:"type:uuid;index" json:"escalated_to"`
	Reason      string     `gorm:"type:text" json:"reason"`
	EscalatedAt time.Time  `gorm:"index" json:"escalated_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}
