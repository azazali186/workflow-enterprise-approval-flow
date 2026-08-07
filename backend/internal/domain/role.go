package domain

// Role represents a system role
type Role struct {
	Base
	Name        string       `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Description string       `gorm:"size:500" json:"description"`
	IsDefault   bool         `gorm:"default:false" json:"is_default"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}
