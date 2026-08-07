package domain

import "github.com/google/uuid"

// AuditLog represents an audit log entry
type AuditLog struct {
	Base
	EntityType string     `gorm:"size:100;index" json:"entity_type"`
	EntityID   uuid.UUID  `gorm:"type:uuid;index" json:"entity_id"`
	Action     string     `gorm:"size:100;index" json:"action"`
	ActorID    *uuid.UUID `gorm:"type:uuid" json:"actor_id,omitempty"`
	Changes    JSONMap    `gorm:"type:jsonb" json:"changes,omitempty"`
	IPAddress  *string    `gorm:"size:45" json:"ip_address,omitempty"`
	UserAgent  *string    `gorm:"type:text" json:"user_agent,omitempty"`
}
