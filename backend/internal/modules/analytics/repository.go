package analytics

import (
	"context"
	"time"

	"github.com/aeroxe/approval-flow/internal/domain"
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
		// Columns are qualified because both approvals and applications have
		// created_at/decided_at columns after the JOIN.
		Select("COALESCE(AVG(EXTRACT(EPOCH FROM (approvals.decided_at - approvals.created_at))), 0)").
		Joins("JOIN applications ON applications.id = approvals.application_id").
		Where("applications.workflow_id = ? AND approvals.status IN ('approved', 'rejected')", workflowID).
		Scan(&avgTime).Error
	if err != nil {
		return nil, err
	}

	result["avg_decision_time_seconds"] = avgTime
	return result, nil
}

type approverPerformance struct {
	Total    int64
	Approved int64
	Rejected int64
	AvgTime  float64
}

func (r *Repository) GetApproverPerformance(ctx context.Context, approverID string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	var perf approverPerformance
	err := r.db.WithContext(ctx).Model(&domain.Approval{}).
		Select("COUNT(*) as total, SUM(CASE WHEN status = 'approved' THEN 1 ELSE 0 END) as approved, SUM(CASE WHEN status = 'rejected' THEN 1 ELSE 0 END) as rejected, COALESCE(AVG(EXTRACT(EPOCH FROM (decided_at - created_at))), 0) as avg_time").
		Where("approver_id = ? AND status IN ('approved', 'rejected')", approverID).
		Scan(&perf).Error
	if err != nil {
		return nil, err
	}

	result["total_decisions"] = perf.Total
	result["approved"] = perf.Approved
	result["rejected"] = perf.Rejected
	if perf.Total > 0 {
		result["approval_rate"] = float64(perf.Approved) / float64(perf.Total)
	}
	result["avg_response_time_seconds"] = perf.AvgTime

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
