package workflow

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

func (r *Repository) Create(ctx context.Context, workflow *domain.Workflow) error {
	return r.db.WithContext(ctx).Create(workflow).Error
}

func (r *Repository) GetByID(ctx context.Context, id string) (*domain.Workflow, error) {
	var workflow domain.Workflow
	err := r.db.WithContext(ctx).First(&workflow, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (r *Repository) GetByCategory(ctx context.Context, category string) ([]domain.Workflow, error) {
	var workflows []domain.Workflow
	err := r.db.WithContext(ctx).Where("category = ? AND is_active = ?", category, true).Order("created_at DESC").Find(&workflows).Error
	return workflows, err
}

func (r *Repository) GetAll(ctx context.Context, limit, offset int) ([]domain.Workflow, error) {
	var workflows []domain.Workflow
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Order("created_at DESC").Find(&workflows).Error
	return workflows, err
}

func (r *Repository) Update(ctx context.Context, workflow *domain.Workflow) error {
	return r.db.WithContext(ctx).Save(workflow).Error
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Workflow{}).Where("id = ?", id).Update("deleted_at", &now).Error
}

func (r *Repository) CreateStep(ctx context.Context, step *domain.WorkflowStep) error {
	return r.db.WithContext(ctx).Create(step).Error
}

func (r *Repository) GetStepsByWorkflowID(ctx context.Context, workflowID string) ([]domain.WorkflowStep, error) {
	var steps []domain.WorkflowStep
	err := r.db.WithContext(ctx).Where("workflow_id = ?", workflowID).Order("step_order ASC").Find(&steps).Error
	return steps, err
}

// ListWithPagination returns paginated workflows with filters and sorting
func (r *Repository) ListWithPagination(
	ctx context.Context,
	filters *pagination.WorkflowFilters,
	cursor *pagination.Cursor,
	limit int,
) ([]domain.Workflow, int64, error) {
	var workflows []domain.Workflow
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&domain.Workflow{})

	// Apply filters
	if filters != nil {
		// Search filter
		if filters.Search != nil && filters.Search.Query != "" {
			searchPattern := "%" + filters.Search.Query + "%"
			query = query.Where(
				"name ILIKE ? OR category ILIKE ? OR description ILIKE ?",
				searchPattern, searchPattern, searchPattern,
			)
		}

		// Category filter
		if filters.Category != "" {
			query = query.Where("category = ?", filters.Category)
		}

		// IsActive filter
		if filters.IsActive != nil {
			query = query.Where("is_active = ?", *filters.IsActive)
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
		"workflow",
	)
	query = query.Order(pagination.GetSortBySQL(sortConfig, ""))

	// Apply limit + 1 to check if there are more results
	query = query.Limit(limit + 1)

	// Execute query
	if err := query.Find(&workflows).Error; err != nil {
		return nil, 0, err
	}

	return workflows, totalCount, nil
}

// GetSummary returns workflow summary statistics
func (r *Repository) GetSummary(ctx context.Context, dateRange *pagination.DateRangeFilter) (*pagination.WorkflowSummary, error) {
	summary := &pagination.WorkflowSummary{
		ByCategory: make(map[string]int64),
	}

	query := r.db.WithContext(ctx).Model(&domain.Workflow{})

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
	if err := query.Count(&summary.TotalWorkflows).Error; err != nil {
		return nil, err
	}

	// Count active workflows
	if err := r.db.WithContext(ctx).Model(&domain.Workflow{}).
		Where("is_active = ?", true).
		Count(&summary.ActiveWorkflows).Error; err != nil {
		return nil, err
	}

	// Count by category
	categoryResults := []struct {
		Category string
		Count    int64
	}{}
	if err := query.Select("category, COUNT(*) as count").Group("category").Scan(&categoryResults).Error; err != nil {
		return nil, err
	}
	for _, r := range categoryResults {
		summary.ByCategory[r.Category] = r.Count
	}

	// Calculate average steps count
	var avgSteps float64
	if err := r.db.WithContext(ctx).
		Select("AVG(jsonb_array_length(steps::jsonb))").
		Where("steps IS NOT NULL").
		Scan(&avgSteps).Error; err == nil {
		summary.AvgStepsCount = avgSteps
	}

	return summary, nil
}
