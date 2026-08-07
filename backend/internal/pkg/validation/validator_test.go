package validation

import (
	"testing"
)

func TestValidator_ValidateUUID(t *testing.T) {
	v := New()

	// Test valid UUID
	v.ValidateUUID("id", "550e8400-e29b-41d4-a716-446655440000")
	if v.HasErrors() {
		t.Errorf("expected no errors for valid UUID, got: %v", v.Errors())
	}

	// Test invalid UUID
	v = New()
	v.ValidateUUID("id", "invalid-uuid")
	if !v.HasErrors() {
		t.Error("expected error for invalid UUID")
	}

	// Test empty UUID
	v = New()
	v.ValidateUUID("id", "")
	if !v.HasErrors() {
		t.Error("expected error for empty UUID")
	}
}

func TestValidator_ValidateRequired(t *testing.T) {
	v := New()

	// Test valid required field
	v.ValidateRequired("name", "John")
	if v.HasErrors() {
		t.Errorf("expected no errors for valid name, got: %v", v.Errors())
	}

	// Test empty required field
	v = New()
	v.ValidateRequired("name", "")
	if !v.HasErrors() {
		t.Error("expected error for empty name")
	}

	// Test whitespace only
	v = New()
	v.ValidateRequired("name", "   ")
	if !v.HasErrors() {
		t.Error("expected error for whitespace-only name")
	}
}

func TestValidator_ValidateEmail(t *testing.T) {
	v := New()

	// Test valid email
	v.ValidateEmail("email", "test@example.com")
	if v.HasErrors() {
		t.Errorf("expected no errors for valid email, got: %v", v.Errors())
	}

	// Test invalid email
	v = New()
	v.ValidateEmail("email", "invalid-email")
	if !v.HasErrors() {
		t.Error("expected error for invalid email")
	}

	// Test empty email
	v = New()
	v.ValidateEmail("email", "")
	if !v.HasErrors() {
		t.Error("expected error for empty email")
	}
}

func TestValidator_ValidateMinLength(t *testing.T) {
	v := New()

	// Test valid length
	v.ValidateMinLength("name", "John", 3)
	if v.HasErrors() {
		t.Errorf("expected no errors for valid length, got: %v", v.Errors())
	}

	// Test too short
	v = New()
	v.ValidateMinLength("name", "Jo", 3)
	if !v.HasErrors() {
		t.Error("expected error for too short name")
	}
}

func TestValidator_ValidateMaxLength(t *testing.T) {
	v := New()

	// Test valid length
	v.ValidateMaxLength("name", "John", 10)
	if v.HasErrors() {
		t.Errorf("expected no errors for valid length, got: %v", v.Errors())
	}

	// Test too long
	v = New()
	v.ValidateMaxLength("name", "John Smith Johnson", 10)
	if !v.HasErrors() {
		t.Error("expected error for too long name")
	}
}

func TestValidator_ValidateMin(t *testing.T) {
	v := New()

	// Test valid value
	v.ValidateMin("age", 18, 18)
	if v.HasErrors() {
		t.Errorf("expected no errors for valid age, got: %v", v.Errors())
	}

	// Test too small
	v = New()
	v.ValidateMin("age", 10, 18)
	if !v.HasErrors() {
		t.Error("expected error for too small age")
	}
}

func TestValidator_ValidateMax(t *testing.T) {
	v := New()

	// Test valid value
	v.ValidateMax("age", 50, 100)
	if v.HasErrors() {
		t.Errorf("expected no errors for valid age, got: %v", v.Errors())
	}

	// Test too large
	v = New()
	v.ValidateMax("age", 150, 100)
	if !v.HasErrors() {
		t.Error("expected error for too large age")
	}
}

func TestValidator_ValidateIn(t *testing.T) {
	v := New()

	// Test valid value
	v.ValidateIn("status", "active", []string{"active", "inactive", "pending"})
	if v.HasErrors() {
		t.Errorf("expected no errors for valid status, got: %v", v.Errors())
	}

	// Test invalid value
	v = New()
	v.ValidateIn("status", "unknown", []string{"active", "inactive", "pending"})
	if !v.HasErrors() {
		t.Error("expected error for invalid status")
	}

	// Test empty value (should pass - use ValidateRequired for required)
	v = New()
	v.ValidateIn("status", "", []string{"active", "inactive", "pending"})
	if v.HasErrors() {
		t.Errorf("expected no errors for empty status, got: %v", v.Errors())
	}
}

func TestValidator_ValidatePassword(t *testing.T) {
	v := New()

	// Test valid password
	v.ValidatePassword("password", "StrongPass1")
	if v.HasErrors() {
		t.Errorf("expected no errors for valid password, got: %v", v.Errors())
	}

	// Test too short
	v = New()
	v.ValidatePassword("password", "Short1")
	if !v.HasErrors() {
		t.Error("expected error for too short password")
	}

	// Test no uppercase
	v = New()
	v.ValidatePassword("password", "nouppercase1")
	if !v.HasErrors() {
		t.Error("expected error for password without uppercase")
	}

	// Test no lowercase
	v = New()
	v.ValidatePassword("password", "NOLOWERCASE1")
	if !v.HasErrors() {
		t.Error("expected error for password without lowercase")
	}

	// Test no digit
	v = New()
	v.ValidatePassword("password", "NoDigitHere")
	if !v.HasErrors() {
		t.Error("expected error for password without digit")
	}

	// Test empty
	v = New()
	v.ValidatePassword("password", "")
	if !v.HasErrors() {
		t.Error("expected error for empty password")
	}
}

func TestValidationErrors_Error(t *testing.T) {
	// Test empty errors
	var errors ValidationErrors
	if errors.Error() != "" {
		t.Errorf("expected empty error string, got: %s", errors.Error())
	}

	// Test single error
	errors = ValidationErrors{
		{Field: "name", Message: "is required"},
	}
	if errors.Error() != "name: is required" {
		t.Errorf("expected 'name: is required', got: %s", errors.Error())
	}

	// Test multiple errors
	errors = ValidationErrors{
		{Field: "name", Message: "is required"},
		{Field: "email", Message: "is invalid"},
	}
	expected := "name: is required; email: is invalid"
	if errors.Error() != expected {
		t.Errorf("expected '%s', got: %s", expected, errors.Error())
	}
}

func TestValidationErrors_HasErrors(t *testing.T) {
	// Test empty errors
	var errors ValidationErrors
	if errors.HasErrors() {
		t.Error("expected no errors")
	}

	// Test with errors
	errors = ValidationErrors{
		{Field: "name", Message: "is required"},
	}
	if !errors.HasErrors() {
		t.Error("expected errors")
	}
}

func TestParseUUID(t *testing.T) {
	// Test valid UUID
	id, err := ParseUUID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if id.String() != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected 550e8400-e29b-41d4-a716-446655440000, got: %s", id)
	}

	// Test invalid UUID
	_, err = ParseUUID("invalid")
	if err == nil {
		t.Error("expected error for invalid UUID")
	}
}
