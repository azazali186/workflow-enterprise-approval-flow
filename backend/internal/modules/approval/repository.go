package approval

import (
	"context"
	"time"

	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/pkg/pagination"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, approval *domain.Approval) error {
	return r.db.WithContext(ctx).Create(approval).Error
}

func (r *Repository) GetByID(ctx context.Context, id string) (*domain.Approval, error) {
	var approval domain.Approval
	err := r.db.WithContext(ctx).First(&approval, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &approval, nil
}

func (r *Repository) GetByApplicationID(ctx context.Context, applicationID string) ([]domain.Approval, error) {
	var approvals []domain.Approval
	err := r.db.WithContext(ctx).Where("application_id = ?", applicationID).Order("created_at ASC").Find(&approvals).Error
	return approvals, err
}

func (r *Repository) GetPendingByApproverID(ctx context.Context, approverID string) ([]domain.Approval, error) {
	var approvals []domain.Approval
	err := r.db.WithContext(ctx).Where("approver_id = ? AND status = ?", approverID, "pending").Order("created_at ASC").Find(&approvals).Error
	return approvals, err
}

func (r *Repository) Update(ctx context.Context, approval *domain.Approval) error {
	return r.db.WithContext(ctx).Save(approval).Error
}

// ValidateStepBelongsToApplication checks that the workflow step belongs to the
// same workflow as the application. An approval linking an application to a
// step of a *different* workflow is a data-integrity violation (the UI now
// prevents it, but the API must too — a hand-crafted request must not succeed).
func (r *Repository) ValidateStepBelongsToApplication(ctx context.Context, applicationID, stepID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.WorkflowStep{}).
		Joins("JOIN applications ON applications.workflow_id = workflow_steps.workflow_id").
		Where("workflow_steps.id = ? AND applications.id = ?", stepID, applicationID).
		Where("workflow_steps.deleted_at = 0 AND applications.deleted_at = 0").
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListOverduePending returns pending approvals whose step SLA (timeout_hours
// from the workflow_steps table) has expired. Used by the SLA escalation
// monitor to trigger escalations.
func (r *Repository) ListOverduePending(ctx context.Context, now time.Time) ([]domain.Approval, error) {
	var approvals []domain.Approval
	err := r.db.WithContext(ctx).
		Model(&domain.Approval{}).
		Joins("JOIN workflow_steps ON workflow_steps.id = approvals.workflow_step_id").
		Where("approvals.status = ?", "pending").
		Where("workflow_steps.timeout_hours > 0").
		Where("approvals.created_at + make_interval(hours => workflow_steps.timeout_hours) < ?", now).
		Find(&approvals).Error
	return approvals, err
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Approval{}).Where("id = ?", id).Update("deleted_at", &now).Error
}

// ListWithPagination returns paginated approvals with filters and sorting
func (r *Repository) ListWithPagination(
	ctx context.Context,
	filters *pagination.ApprovalFilters,
	cursor *pagination.Cursor,
	limit int,
) ([]domain.Approval, int64, error) {
	var approvals []domain.Approval
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&domain.Approval{})

	// Apply filters
	if filters != nil {
		// Search filter
		if filters.Search != nil && filters.Search.Query != "" {
			searchPattern := "%" + filters.Search.Query + "%"
			query = query.Where(
				"status ILIKE ? OR decision ILIKE ?",
				searchPattern, searchPattern,
			)
		}

		// Status filter
		if filters.Status != nil && len(filters.Status.Statuses) > 0 {
			query = query.Where("status IN ?", filters.Status.Statuses)
		}

		// Approver filter
		if filters.ApproverID != "" {
			query = query.Where("approver_id = ?", filters.ApproverID)
		}

		// Application filter
		if filters.ApplicationID != "" {
			query = query.Where("application_id = ?", filters.ApplicationID)
		}

		// Decision filter
		if filters.Decision != "" {
			query = query.Where("decision = ?", filters.Decision)
		}

		// Date range filter
		if filters.DateRange != nil {
			if filters.DateRange.Start != nil {
				query = query.Where("created_at >= ?", *filters.DateRange.Start)
			}
			if filters.DateRange.End != nil {
				query = query.Where("created_at <= ?", *filters.DateRange.End)
			}
		}
	}

	// Get total count
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply cursor pagination
	if cursor != nil {
		query = query.Where(
			"(created_at < ?) OR (created_at = ? AND id < ?)",
			cursor.CreatedAt, cursor.CreatedAt, cursor.ID,
		)
	}

	// Apply sorting
	sortConfig := pagination.ParseSortConfig(
		filters.SortBy,
		filters.SortOrder,
		"approval",
	)
	query = query.Order(pagination.GetSortBySQL(sortConfig, ""))

	// Apply limit + 1 to check if there are more results
	query = query.Limit(limit + 1)

	// Execute query
	if err := query.Find(&approvals).Error; err != nil {
		return nil, 0, err
	}

	return approvals, totalCount, nil
}

// GetSummary returns approval summary statistics
func (r *Repository) GetSummary(ctx context.Context, dateRange *pagination.DateRangeFilter) (*pagination.ApprovalSummary, error) {
	summary := &pagination.ApprovalSummary{
		ByStatus: make(map[string]int64),
	}

	query := r.db.WithContext(ctx).Model(&domain.Approval{})

	// Apply date range filter
	if dateRange != nil {
		if dateRange.Start != nil {
			query = query.Where("created_at >= ?", *dateRange.Start)
		}
		if dateRange.End != nil {
			query = query.Where("created_at <= ?", *dateRange.End)
		}
	}

	// Get total count
	if err := query.Count(&summary.TotalApprovals).Error; err != nil {
		return nil, err
	}

	// Count by status
	statusResults := []struct {
		Status string
		Count  int64
	}{}
	if err := query.Select("status, COUNT(*) as count").Group("status").Scan(&statusResults).Error; err != nil {
		return nil, err
	}
	for _, r := range statusResults {
		summary.ByStatus[r.Status] = r.Count
	}

	// Count pending approvals
	if err := r.db.WithContext(ctx).Model(&domain.Approval{}).
		Where("status = ?", "pending").
		Count(&summary.PendingApprovals).Error; err != nil {
		return nil, err
	}

	// Calculate average decision time
	var avgTime float64
	if err := r.db.WithContext(ctx).
		Select("AVG(EXTRACT(EPOCH FROM (decided_at - created_at)) / 3600)").
		Where("decided_at IS NOT NULL").
		Scan(&avgTime).Error; err == nil {
		summary.AvgDecisionTime = avgTime
	}

	// Calculate approval rate
	var approved, rejected int64
	if err := r.db.WithContext(ctx).Model(&domain.Approval{}).Where("status = ?", "approved").Count(&approved).Error; err == nil {
		if err := r.db.WithContext(ctx).Model(&domain.Approval{}).Where("status = ?", "rejected").Count(&rejected).Error; err == nil {
			total := approved + rejected
			if total > 0 {
				summary.ApprovalRate = float64(approved) / float64(total) * 100
			}
		}
	}

	return summary, nil
}
