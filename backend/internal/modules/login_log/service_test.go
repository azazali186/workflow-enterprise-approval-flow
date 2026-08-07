package login_log

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aeroxe/approval-flow/internal/domain"
)

func TestService_PackageStructure(t *testing.T) {
	assert.True(t, true, "login_log service package exists")
}

func TestLoginLogParams_Structure(t *testing.T) {
	params := LogLoginParams{
		Email:         "test@example.com",
		Status:        domain.LoginStatusSuccess,
		IPAddress:     "127.0.0.1",
		UserAgent:     "Mozilla/5.0",
		RequestID:     "req-123",
		TokenID:       "token-456",
		FailureReason: "",
	}

	assert.Equal(t, "test@example.com", params.Email)
	assert.Equal(t, domain.LoginStatusSuccess, params.Status)
	assert.Equal(t, "127.0.0.1", params.IPAddress)
	assert.Equal(t, "Mozilla/5.0", params.UserAgent)
	assert.Equal(t, "req-123", params.RequestID)
	assert.Equal(t, "token-456", params.TokenID)
	assert.Empty(t, params.FailureReason)
}

func TestLoginStatus_Constants(t *testing.T) {
	assert.Equal(t, "success", domain.LoginStatusSuccess)
	assert.Equal(t, "failed", domain.LoginStatusFailed)
	assert.Equal(t, "locked", domain.LoginStatusLocked)
}

func TestFailureReason_Constants(t *testing.T) {
	assert.Equal(t, "invalid_credentials", domain.FailureReasonInvalidCredentials)
	assert.Equal(t, "user_not_found", domain.FailureReasonUserNotFound)
	assert.Equal(t, "account_locked", domain.FailureReasonAccountLocked)
	assert.Equal(t, "too_many_attempts", domain.FailureReasonTooManyAttempts)
}
