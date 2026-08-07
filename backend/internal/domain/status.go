package domain

import "github.com/google/uuid"

// Status represents a status record in the system
type Status struct {
	Base
	EntityType string    `gorm:"size:100;index" json:"entity_type"`
	EntityID   uuid.UUID `gorm:"type:uuid;index" json:"entity_id"`
	State      string    `gorm:"size:50;index" json:"state"`
	Reason     *string   `gorm:"type:text" json:"reason,omitempty"`
}
