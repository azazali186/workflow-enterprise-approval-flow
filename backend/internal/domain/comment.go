package domain

import "github.com/google/uuid"

// Comment represents a comment on an approval
type Comment struct {
	Base
	ApprovalID uuid.UUID `gorm:"type:uuid;index" json:"approval_id"`
	UserID     uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	IsInternal bool      `gorm:"default:false;index" json:"is_internal"`
}
