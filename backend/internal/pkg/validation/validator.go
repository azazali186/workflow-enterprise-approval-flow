package validation

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	messages := make([]string, len(ve))
	for i, e := range ve {
		messages[i] = fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return strings.Join(messages, "; ")
}

// HasErrors returns true if there are validation errors
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// Add adds a validation error
func (ve *ValidationErrors) Add(field, message string) {
	*ve = append(*ve, ValidationError{Field: field, Message: message})
}

// Validator provides validation methods
type Validator struct {
	errors ValidationErrors
}

// New creates a new validator
func New() *Validator {
	return &Validator{}
}

// Errors returns the validation errors
func (v *Validator) Errors() ValidationErrors {
	return v.errors
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return v.errors.HasErrors()
}

// ValidateUUID validates a UUID string
func (v *Validator) ValidateUUID(field, value string) {
	if value == "" {
		v.errors.Add(field, "is required")
		return
	}
	if _, err := uuid.Parse(value); err != nil {
		v.errors.Add(field, "must be a valid UUID")
	}
}

// ValidateRequired validates that a field is not empty
func (v *Validator) ValidateRequired(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.errors.Add(field, "is required")
	}
}

// ValidateEmail validates an email address
func (v *Validator) ValidateEmail(field, value string) {
	if value == "" {
		v.errors.Add(field, "is required")
		return
	}
	if _, err := mail.ParseAddress(value); err != nil {
		v.errors.Add(field, "must be a valid email address")
	}
}

// ValidateMinLength validates minimum length
func (v *Validator) ValidateMinLength(field, value string, min int) {
	if len(value) < min {
		v.errors.Add(field, fmt.Sprintf("must be at least %d characters", min))
	}
}

// ValidateMaxLength validates maximum length
func (v *Validator) ValidateMaxLength(field, value string, max int) {
	if len(value) > max {
		v.errors.Add(field, fmt.Sprintf("must be at most %d characters", max))
	}
}

// ValidateMin validates minimum value
func (v *Validator) ValidateMin(field string, value, min int) {
	if value < min {
		v.errors.Add(field, fmt.Sprintf("must be at least %d", min))
	}
}

// ValidateMax validates maximum value
func (v *Validator) ValidateMax(field string, value, max int) {
	if value > max {
		v.errors.Add(field, fmt.Sprintf("must be at most %d", max))
	}
}

// ValidateIn validates that value is in the allowed list
func (v *Validator) ValidateIn(field, value string, allowed []string) {
	if value == "" {
		return // Skip if empty (use ValidateRequired for required fields)
	}
	for _, a := range allowed {
		if value == a {
			return
		}
	}
	v.errors.Add(field, fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", ")))
}

// ValidatePattern validates against a regex pattern
func (v *Validator) ValidatePattern(field, value, pattern string) {
	if value == "" {
		return
	}
	matched, err := regexp.MatchString(pattern, value)
	if err != nil || !matched {
		v.errors.Add(field, "invalid format")
	}
}

// ValidatePassword validates password strength
func (v *Validator) ValidatePassword(field, value string) {
	if value == "" {
		v.errors.Add(field, "is required")
		return
	}
	if len(value) < 8 {
		v.errors.Add(field, "must be at least 8 characters")
		return
	}
	if len(value) > 128 {
		v.errors.Add(field, "must be at most 128 characters")
		return
	}
	// Check for at least one uppercase, one lowercase, and one digit
	hasUpper := false
	hasLower := false
	hasDigit := false
	for _, c := range value {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		v.errors.Add(field, "must contain at least one uppercase letter, one lowercase letter, and one digit")
	}
}

// ValidateURL validates a URL
func (v *Validator) ValidateURL(field, value string) {
	if value == "" {
		return
	}
	if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
		v.errors.Add(field, "must be a valid URL")
	}
}

// ValidateSlice validates a slice is not empty
func (v *Validator) ValidateSliceNotEmpty(field string, slice interface{}) {
	switch s := slice.(type) {
	case []string:
		if len(s) == 0 {
			v.errors.Add(field, "must not be empty")
		}
	case []uuid.UUID:
		if len(s) == 0 {
			v.errors.Add(field, "must not be empty")
		}
	}
}

// Add adds a validation error to the validator
func (v *Validator) Add(field, message string) {
	v.errors.Add(field, message)
}
