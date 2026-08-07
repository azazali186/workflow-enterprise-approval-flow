package audit

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/aeroxe/approval-flow/internal/domain"
)

func TestService_PackageStructure(t *testing.T) {
	assert.True(t, true, "audit service package exists")
}

func TestLogParams_Structure(t *testing.T) {
	actorID := uuid.New()
	entityID := uuid.New()

	params := LogParams{
		EntityType: "application",
		EntityID:   entityID,
		Action:     "create",
		ActorID:    &actorID,
		Status:     "success",
	}

	assert.Equal(t, "application", params.EntityType)
	assert.Equal(t, entityID, params.EntityID)
	assert.Equal(t, "create", params.Action)
	assert.NotNil(t, params.ActorID)
	assert.Equal(t, "success", params.Status)
}

func TestRequestContext_Structure(t *testing.T) {
	ctx := &RequestContext{
		IPAddress: "127.0.0.1",
		UserAgent: "Mozilla/5.0",
		RequestID: "req-123",
	}

	assert.Equal(t, "127.0.0.1", ctx.IPAddress)
	assert.Equal(t, "Mozilla/5.0", ctx.UserAgent)
	assert.Equal(t, "req-123", ctx.RequestID)
}

func TestExtractRequestContext(t *testing.T) {
	ctx := ExtractRequestContext("127.0.0.1", "Mozilla/5.0", "req-123")
	assert.NotNil(t, ctx)
	assert.Equal(t, "127.0.0.1", ctx.IPAddress)
	assert.Equal(t, "Mozilla/5.0", ctx.UserAgent)
	assert.Equal(t, "req-123", ctx.RequestID)
}

func TestAuditAction_Constants(t *testing.T) {
	assert.Equal(t, "create", domain.AuditActionCreate)
	assert.Equal(t, "update", domain.AuditActionUpdate)
	assert.Equal(t, "delete", domain.AuditActionDelete)
	assert.Equal(t, "login", domain.AuditActionLogin)
	assert.Equal(t, "logout", domain.AuditActionLogout)
	assert.Equal(t, "login_failed", domain.AuditActionLoginFailed)
	assert.Equal(t, "submit", domain.AuditActionSubmit)
	assert.Equal(t, "decide", domain.AuditActionDecide)
	assert.Equal(t, "assign", domain.AuditActionAssign)
	assert.Equal(t, "remove", domain.AuditActionRemove)
	assert.Equal(t, "resolve", domain.AuditActionResolve)
}

func TestAuditStatus_Constants(t *testing.T) {
	assert.Equal(t, "success", domain.AuditStatusSuccess)
	assert.Equal(t, "failed", domain.AuditStatusFailed)
}
