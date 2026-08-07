package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/pkg/audit"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Service handles audit logging business logic
type Service struct {
	db     *gorm.DB
	logger *config.Config
}

// NewService creates a new audit service
func NewService(db *gorm.DB, cfg *config.Config) *Service {
	return &Service{
		db:     db,
		logger: cfg,
	}
}

// LogParams holds parameters for creating an audit log entry
type LogParams struct {
	EntityType   string
	EntityID     uuid.UUID
	Action       string
	ActorID      *uuid.UUID
	ActorEmail   *string
	BeforeState  interface{}
	AfterState   interface{}
	Request      *RequestContext
	Status       string
	ErrorMessage *string
	DurationMs   *int64
}

// RequestContext holds request metadata for audit logging
type RequestContext struct {
	IPAddress string
	UserAgent string
	RequestID string
}

// Log creates an audit log entry with before/after state
func (s *Service) Log(ctx context.Context, params LogParams) error {
	// Compute changes if both before and after are provided
	var changes domain.JSONMap
	var beforeState domain.JSONMap
	var afterState domain.JSONMap

	if params.BeforeState != nil {
		beforeState = audit.StructToMapWithJSONTags(params.BeforeState)
		// Filter sensitive fields from before state
		beforeState = audit.FilterSensitiveFields(beforeState)
	}

	if params.AfterState != nil {
		afterState = audit.StructToMapWithJSONTags(params.AfterState)
		afterState = audit.FilterSensitiveFields(afterState)
	}

	if params.BeforeState != nil && params.AfterState != nil {
		diff := audit.ComputeChangesFromMaps(beforeState, afterState)
		if diff.HasChanges {
			changes = audit.ChangesToJSONMap(diff)
		}
	}

	// Set default status
	if params.Status == "" {
		params.Status = domain.AuditStatusSuccess
	}

	log := &domain.AuditLog{
		EntityType:   params.EntityType,
		EntityID:     params.EntityID,
		Action:       params.Action,
		ActorID:      params.ActorID,
		ActorEmail:   params.ActorEmail,
		Changes:      changes,
		BeforeState:  beforeState,
		AfterState:   afterState,
		Status:       params.Status,
		ErrorMessage: params.ErrorMessage,
		DurationMs:   params.DurationMs,
	}

	if params.Request != nil {
		log.IPAddress = &params.Request.IPAddress
		log.UserAgent = &params.Request.UserAgent
		log.RequestID = &params.Request.RequestID
	}

	if err := s.db.WithContext(ctx).Create(log).Error; err != nil {
		s.logger.Error("failed to create audit log",
			zap.Error(err),
			zap.String("entity_type", params.EntityType),
			zap.String("action", params.Action),
		)
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// LogCreate logs a create operation
func (s *Service) LogCreate(ctx context.Context, entityType string, entityID uuid.UUID, actorID *uuid.UUID, entity interface{}, reqCtx *RequestContext) error {
	return s.Log(ctx, LogParams{
		EntityType:  entityType,
		EntityID:    entityID,
		Action:      domain.AuditActionCreate,
		ActorID:     actorID,
		AfterState:  entity,
		Request:     reqCtx,
		Status:      domain.AuditStatusSuccess,
	})
}

// LogUpdate logs an update operation with before/after state
func (s *Service) LogUpdate(ctx context.Context, entityType string, entityID uuid.UUID, actorID *uuid.UUID, before, after interface{}, reqCtx *RequestContext) error {
	return s.Log(ctx, LogParams{
		EntityType:  entityType,
		EntityID:    entityID,
		Action:      domain.AuditActionUpdate,
		ActorID:     actorID,
		BeforeState: before,
		AfterState:  after,
		Request:     reqCtx,
		Status:      domain.AuditStatusSuccess,
	})
}

// LogDelete logs a delete operation with before state
func (s *Service) LogDelete(ctx context.Context, entityType string, entityID uuid.UUID, actorID *uuid.UUID, entity interface{}, reqCtx *RequestContext) error {
	return s.Log(ctx, LogParams{
		EntityType:  entityType,
		EntityID:    entityID,
		Action:      domain.AuditActionDelete,
		ActorID:     actorID,
		BeforeState: entity,
		Request:     reqCtx,
		Status:      domain.AuditStatusSuccess,
	})
}

// LogAction logs a generic action with optional before/after state
func (s *Service) LogAction(ctx context.Context, params LogParams) error {
	return s.Log(ctx, params)
}

// LogError logs a failed action
func (s *Service) LogError(ctx context.Context, entityType string, entityID uuid.UUID, action string, actorID *uuid.UUID, err error, reqCtx *RequestContext) error {
	errMsg := err.Error()
	return s.Log(ctx, LogParams{
		EntityType:   entityType,
		EntityID:     entityID,
		Action:       action,
		ActorID:      actorID,
		Request:      reqCtx,
		Status:       domain.AuditStatusFailed,
		ErrorMessage: &errMsg,
	})
}

// ExtractRequestContext extracts audit request context from Hertz request
func ExtractRequestContext(ipAddress, userAgent, requestID string) *RequestContext {
	return &RequestContext{
		IPAddress: ipAddress,
		UserAgent: userAgent,
		RequestID: requestID,
	}
}

// GetAuditLogs retrieves audit logs with filters
func (s *Service) GetAuditLogs(ctx context.Context, limit, offset int, entityType, entityID string) ([]domain.AuditLog, int64, error) {
	var logs []domain.AuditLog
	var totalCount int64

	query := s.db.WithContext(ctx).Model(&domain.AuditLog{})

	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}
	if entityID != "" {
		parsedID, err := uuid.Parse(entityID)
		if err == nil {
			query = query.Where("entity_id = ?", parsedID)
		}
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error

	return logs, totalCount, err
}

// GetEntityHistory retrieves the audit history for a specific entity
func (s *Service) GetEntityHistory(ctx context.Context, entityType string, entityID uuid.UUID) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	err := s.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

// GetUserActivity retrieves all activity by a specific user
func (s *Service) GetUserActivity(ctx context.Context, actorID uuid.UUID, limit int) ([]domain.AuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	var logs []domain.AuditLog
	err := s.db.WithContext(ctx).
		Where("actor_id = ?", actorID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// helper function to check if time is zero
func isZeroTime(t time.Time) bool {
	return t.IsZero()
}
