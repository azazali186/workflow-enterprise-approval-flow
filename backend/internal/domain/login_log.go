package domain

import (
	"time"

	"github.com/google/uuid"
)

// LoginLog represents a login attempt log entry
type LoginLog struct {
	Base
	UserID        *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Email         string     `gorm:"size:255;index" json:"email"`
	Status        string     `gorm:"size:20;index;not null" json:"status"` // success, failed, locked
	FailureReason *string    `gorm:"size:255" json:"failure_reason,omitempty"`
	IPAddress     string     `gorm:"size:45" json:"ip_address"`
	UserAgent     string     `gorm:"type:text" json:"user_agent"`
	RequestID     *string    `gorm:"size:100" json:"request_id,omitempty"`
	TokenID       *string    `gorm:"size:100" json:"token_id,omitempty"` // JWT ID (jti) for successful logins
	AttemptedAt   time.Time  `gorm:"index;not null" json:"attempted_at"`
}

// Login log status constants
const (
	LoginStatusSuccess = "success"
	LoginStatusFailed  = "failed"
	LoginStatusLocked  = "locked"
)

// Login failure reason constants
const (
	FailureReasonInvalidCredentials = "invalid_credentials"
	FailureReasonUserNotFound       = "user_not_found"
	FailureReasonAccountLocked      = "account_locked"
	FailureReasonAccountDisabled    = "account_disabled"
	FailureReasonPasswordExpired    = "password_expired"
	FailureReasonTooManyAttempts    = "too_many_attempts"
	FailureReasonTokenExpired       = "token_expired"
	FailureReasonTokenInvalid       = "token_invalid"
	FailureReasonTokenRevoked       = "token_revoked"
	FailureReasonTokenReuse         = "token_reuse_detected"
)
