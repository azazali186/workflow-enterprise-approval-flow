package domain

import (
	"time"

	"github.com/google/uuid"
)

// Notification represents a notification in the system
type Notification struct {
	Base
	UserID  uuid.UUID  `gorm:"type:uuid;index" json:"user_id"`
	Type    string     `gorm:"size:50;index" json:"type"`
	Channel string     `gorm:"size:50" json:"channel"`
	Title   string     `gorm:"size:255" json:"title"`
	Body    string     `gorm:"type:text" json:"body"`
	Data    JSONMap    `gorm:"type:jsonb" json:"data,omitempty"`
	ReadAt  *time.Time `json:"read_at,omitempty"`
	SentAt  *time.Time `json:"sent_at,omitempty"`
}
