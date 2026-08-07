package domain

// Permission represents a system permission (API route)
type Permission struct {
	Base
	Name    string `gorm:"size:255;not null" json:"name"`
	Route   string `gorm:"size:500;uniqueIndex;not null" json:"route"`
	Path    string `gorm:"size:500;not null" json:"path"`
	Method  string `gorm:"size:10;not null" json:"method"`
	Service string `gorm:"size:100;not null;default:approval-flow" json:"service"`
}
