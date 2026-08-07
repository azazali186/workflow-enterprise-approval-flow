package domain

import "github.com/google/uuid"

// Document represents a document attached to an application
type Document struct {
	Base
	ApplicationID uuid.UUID `gorm:"type:uuid;index" json:"application_id"`
	Name          string    `gorm:"size:255;not null" json:"name"`
	FileURL       string    `gorm:"type:text;not null" json:"file_url"`
	FileType      string    `gorm:"size:100" json:"file_type"`
	FileSize      int64     `json:"file_size"`
	UploadedBy    uuid.UUID `gorm:"type:uuid" json:"uploaded_by"`
}
