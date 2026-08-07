package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	RoleID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name for UserRole
func (UserRole) TableName() string {
	return "user_roles"
}
