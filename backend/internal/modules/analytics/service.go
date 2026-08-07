package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service struct {
	Repo   *Repository
	Logger *config.Config
}

func NewService(repo *Repository, cfg *config.Config) *Service {
	return &Service{Repo: repo, Logger: cfg}
}

func (s *Service) GetApprovalStats(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
	stats, err := s.Repo.GetApprovalStats(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get approval stats: %w", err)
	}
	s.Logger.Info("approval stats retrieved")
	return stats, nil
}

func (s *Service) GetWorkflowPerformance(ctx context.Context, workflowID string) (map[string]interface{}, error) {
	perf, err := s.Repo.GetWorkflowPerformance(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow performance: %w", err)
	}
	s.Logger.Info("workflow performance retrieved", zap.String("workflow_id", workflowID))
	return perf, nil
}

func (s *Service) GetApproverPerformance(ctx context.Context, approverID string) (map[string]interface{}, error) {
	perf, err := s.Repo.GetApproverPerformance(ctx, approverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get approver performance: %w", err)
	}
	s.Logger.Info("approver performance retrieved", zap.String("approver_id", approverID))
	return perf, nil
}

func (s *Service) GetEscalationMetrics(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
	metrics, err := s.Repo.GetEscalationMetrics(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get escalation metrics: %w", err)
	}
	s.Logger.Info("escalation metrics retrieved")
	return metrics, nil
}

func (s *Service) CreateAuditLog(ctx context.Context, entityType, entityID string, action string, actorID *uuid.UUID, changes map[string]interface{}, ipAddress, userAgent *string) error {
	parsedEntityID, err := uuid.Parse(entityID)
	if err != nil {
		return fmt.Errorf("invalid entity ID: %w", err)
	}

	log := &domain.AuditLog{
		EntityType: entityType,
		EntityID:   parsedEntityID,
		Action:     action,
		ActorID:    actorID,
		Changes:    changes,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}

	if err := s.Repo.CreateAuditLog(ctx, log); err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	return nil
}
