package domain

// Template represents a template definition in the system
type Template struct {
	Base
	Name     string `gorm:"size:255;not null" json:"name"`
	Category string `gorm:"size:100;index" json:"category"`
	Version  int    `gorm:"default:1" json:"version"`
	IsActive bool   `gorm:"default:true;index" json:"is_active"`
	Schema   JSONMap `gorm:"type:jsonb" json:"schema,omitempty"`
	UI       JSONMap `gorm:"type:jsonb" json:"ui,omitempty"`
}
