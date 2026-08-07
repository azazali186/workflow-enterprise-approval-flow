package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetApprovalStats(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	var total, pending, approved, rejected int64
	err := r.db.WithContext(ctx).Model(&domain.Approval{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Count(&total).
		Error
	if err != nil {
		return nil, err
	}

	r.db.WithContext(ctx).Model(&domain.Approval{}).Where("created_at >= ? AND created_at <= ? AND status = ?", startDate, endDate, "pending").Count(&pending)
	r.db.WithContext(ctx).Model(&domain.Approval{}).Where("created_at >= ? AND created_at <= ? AND status = ?", startDate, endDate, "approved").Count(&approved)
	r.db.WithContext(ctx).Model(&domain.Approval{}).Where("created_at >= ? AND created_at <= ? AND status = ?", startDate, endDate, "rejected").Count(&rejected)

	result["total"] = total
	result["pending"] = pending
	result["approved"] = approved
	result["rejected"] = rejected
	if total > 0 {
		result["approval_rate"] = float64(approved) / float64(total)
	}

	return result, nil
}

func (r *Repository) GetWorkflowPerformance(ctx context.Context, workflowID string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	var avgTime float64
	err := r.db.WithContext(ctx).Model(&domain.Approval{}).
		Select("AVG(EXTRACT(EPOCH FROM (decided_at - created_at)))").
		Joins("JOIN applications ON applications.id = approvals.application_id").
		Where("applications.workflow_id = ? AND approvals.status IN ('approved', 'rejected')", workflowID).
		Scan(&avgTime).Error
	if err != nil {
		return nil, err
	}

	result["avg_decision_time_seconds"] = avgTime
	return result, nil
}

func (r *Repository) GetApproverPerformance(ctx context.Context, approverID string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	var total, approved, rejected int64
	var avgTime float64
	err := r.db.WithContext(ctx).Model(&domain.Approval{}).
		Select("COUNT(*), SUM(CASE WHEN status = 'approved' THEN 1 ELSE 0 END), SUM(CASE WHEN status = 'rejected' THEN 1 ELSE 0 END), AVG(EXTRACT(EPOCH FROM (decided_at - created_at)))").
		Where("approver_id = ? AND status IN ('approved', 'rejected')", approverID).
		Scan(&total, &approved, &rejected, &avgTime).Error
	if err != nil {
		return nil, err
	}

	result["total_decisions"] = total
	result["approved"] = approved
	result["rejected"] = rejected
	if total > 0 {
		result["approval_rate"] = float64(approved) / float64(total)
	}
	result["avg_response_time_seconds"] = avgTime

	return result, nil
}

func (r *Repository) GetEscalationMetrics(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	var totalEscalations, resolvedEscalations int64
	err := r.db.WithContext(ctx).Model(&domain.Escalation{}).
		Where("escalated_at >= ? AND escalated_at <= ?", startDate, endDate).
		Count(&totalEscalations).
		Error
	if err != nil {
		return nil, err
	}

	r.db.WithContext(ctx).Model(&domain.Escalation{}).Where("escalated_at >= ? AND escalated_at <= ? AND resolved_at IS NOT NULL", startDate, endDate).Count(&resolvedEscalations)

	result["total_escalations"] = totalEscalations
	result["resolved"] = resolvedEscalations
	if totalEscalations > 0 {
		result["resolution_rate"] = float64(resolvedEscalations) / float64(totalEscalations)
	}

	return result, nil
}

func (r *Repository) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

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
	s.Logger.Info("workflow performance retrieved", "workflow_id", workflowID)
	return perf, nil
}

func (s *Service) GetApproverPerformance(ctx context.Context, approverID string) (map[string]interface{}, error) {
	perf, err := s.Repo.GetApproverPerformance(ctx, approverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get approver performance: %w", err)
	}
	s.Logger.Info("approver performance retrieved", "approver_id", approverID)
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

func (s *Service) CreateAuditLog(ctx context.Context, entityType, entityID, action string, actorID *uuid.UUID, changes map[string]interface{}, ipAddress, userAgent *string) error {
	log := &domain.AuditLog{
		EntityType: entityType,
		EntityID:   entityID,
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
