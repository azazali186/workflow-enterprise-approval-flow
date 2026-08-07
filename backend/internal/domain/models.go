package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	pkguuid "github.com/aeroxe/approval-flow/internal/pkg/uuid"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

// Base contains common fields for all entities
type Base struct {
	ID        uuid.UUID              `gorm:"type:uuid;primary_key" json:"id"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate generates UUID v7 before creating a record
func (b *Base) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = pkguuid.NewV7()
	}
	return nil
}

// StringArray represents a JSON array of strings
type StringArray []string

// Scan implements the sql.Scanner interface
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = StringArray{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into StringArray", value)
	}
	return json.Unmarshal(bytes, s)
}

// Value implements the driver.Valuer interface
func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// JSONMap represents a JSON object stored as JSONB in PostgreSQL
type JSONMap map[string]interface{}

// Scan implements the sql.Scanner interface
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = JSONMap{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONMap", value)
	}
	return json.Unmarshal(bytes, j)
}

// Value implements the driver.Valuer interface
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// SensitiveString is a string type that masks its value in JSON output
type SensitiveString string

// MarshalJSON masks sensitive data in JSON output
func (s SensitiveString) MarshalJSON() ([]byte, error) {
	if s == "" {
		return json.Marshal("")
	}
	return json.Marshal("[REDACTED]")
}

// String returns the raw value (for internal use only)
func (s SensitiveString) String() string {
	return string(s)
}

// Value returns the raw value for database storage
func (s SensitiveString) Value() (driver.Value, error) {
	return string(s), nil
}

// Scan implements the sql.Scanner interface
func (s *SensitiveString) Scan(value interface{}) error {
	if value == nil {
		*s = ""
		return nil
	}
	switch v := value.(type) {
	case string:
		*s = SensitiveString(v)
	case []byte:
		*s = SensitiveString(v)
	}
	return nil
}
