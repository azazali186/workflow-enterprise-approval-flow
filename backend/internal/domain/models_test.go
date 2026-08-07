package domain

import (
	"encoding/json"
	"testing"
)

func TestJSONMap_Scan(t *testing.T) {
	// Test scanning nil value
	var j JSONMap
	err := j.Scan(nil)
	if err != nil {
		t.Errorf("expected no error for nil scan, got: %v", err)
	}
	if j == nil {
		t.Error("expected non-nil JSONMap after nil scan")
	}

	// Test scanning byte slice
	j = nil
	err = j.Scan([]byte(`{"key": "value"}`))
	if err != nil {
		t.Errorf("expected no error for byte scan, got: %v", err)
	}
	if j["key"] != "value" {
		t.Errorf("expected 'value', got: %v", j["key"])
	}

	// Test scanning invalid type
	j = nil
	err = j.Scan("invalid")
	if err == nil {
		t.Error("expected error for invalid type")
	}
}

func TestJSONMap_Value(t *testing.T) {
	// Test nil JSONMap
	var j JSONMap
	val, err := j.Value()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if val != nil {
		t.Errorf("expected nil value, got: %v", val)
	}

	// Test valid JSONMap
	j = JSONMap{"key": "value"}
	val, err = j.Value()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if val == nil {
		t.Error("expected non-nil value")
	}

	// Verify serialization
	data, err := json.Marshal(j)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("expected 'value', got: %v", result["key"])
	}
}

func TestStringArray_Scan(t *testing.T) {
	// Test scanning nil value
	var s StringArray
	err := s.Scan(nil)
	if err != nil {
		t.Errorf("expected no error for nil scan, got: %v", err)
	}
	if s == nil {
		t.Error("expected non-nil StringArray after nil scan")
	}

	// Test scanning byte slice
	s = nil
	err = s.Scan([]byte(`["a", "b", "c"]`))
	if err != nil {
		t.Errorf("expected no error for byte scan, got: %v", err)
	}
	if len(s) != 3 {
		t.Errorf("expected 3 elements, got: %d", len(s))
	}
	if s[0] != "a" || s[1] != "b" || s[2] != "c" {
		t.Errorf("expected [a, b, c], got: %v", s)
	}

	// Test scanning invalid type
	s = nil
	err = s.Scan("invalid")
	if err == nil {
		t.Error("expected error for invalid type")
	}
}

func TestStringArray_Value(t *testing.T) {
	// Test nil StringArray
	var s StringArray
	val, err := s.Value()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if val != nil {
		t.Errorf("expected nil value, got: %v", val)
	}

	// Test valid StringArray
	s = StringArray{"a", "b", "c"}
	val, err = s.Value()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if val == nil {
		t.Error("expected non-nil value")
	}

	// Verify serialization
	data, err := json.Marshal(s)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	var result []string
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 elements, got: %d", len(result))
	}
}

func TestUser_HashPassword(t *testing.T) {
	user := &User{}

	err := user.HashPassword("password123")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if user.Password == "password123" {
		t.Error("expected password to be hashed")
	}

	if user.Password == "" {
		t.Error("expected non-empty hashed password")
	}
}

func TestUser_CheckPassword(t *testing.T) {
	user := &User{}

	// Hash a password
	err := user.HashPassword("password123")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Check correct password
	if !user.CheckPassword("password123") {
		t.Error("expected correct password to return true")
	}

	// Check incorrect password
	if user.CheckPassword("wrongpassword") {
		t.Error("expected incorrect password to return false")
	}
}

func TestBase_JSONTags(t *testing.T) {
	// Test that Base struct has correct JSON tags
	base := Base{}

	data, err := json.Marshal(base)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Verify JSON keys are snake_case
	if _, ok := result["id"]; !ok {
		t.Error("expected 'id' key in JSON")
	}
	if _, ok := result["created_at"]; !ok {
		t.Error("expected 'created_at' key in JSON")
	}
	if _, ok := result["updated_at"]; !ok {
		t.Error("expected 'updated_at' key in JSON")
	}
}

func TestDomainEntities_JSONSerialization(t *testing.T) {
	// Test that all domain entities serialize correctly
	tests := []struct {
		name   string
		entity interface{}
	}{
		{
			name:   "User",
			entity: &User{Email: "test@example.com", Name: "Test User", Status: "active"},
		},
		{
			name:   "Role",
			entity: &Role{Name: "admin", Description: "Admin role"},
		},
		{
			name:   "Permission",
			entity: &Permission{Name: "read", Route: "GET /api/v1/users", Path: "/api/v1/users", Method: "GET"},
		},
		{
			name:   "Workflow",
			entity: &Workflow{Name: "Expense Approval", Category: "finance"},
		},
		{
			name:   "Template",
			entity: &Template{Name: "Purchase Request", Category: "procurement"},
		},
		{
			name:   "Application",
			entity: &Application{Status: "pending", Priority: "high"},
		},
		{
			name:   "Approval",
			entity: &Approval{Status: "pending", Decision: nil},
		},
		{
			name:   "Escalation",
			entity: &Escalation{Level: 1, Reason: "Timeout"},
		},
		{
			name:   "Notification",
			entity: &Notification{Type: "email", Title: "New Approval", Body: "You have a new approval request"},
		},
		{
			name:   "AuditLog",
			entity: &AuditLog{EntityType: "user", Action: "create"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.entity)
			if err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
			if len(data) == 0 {
				t.Error("expected non-empty JSON")
			}
		})
	}
}
