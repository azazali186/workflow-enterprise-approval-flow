package application

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

func (r *Repository) Create(ctx context.Context, app *domain.Application) error {
	return r.db.WithContext(ctx).Create(app).Error
}

func (r *Repository) GetByID(ctx context.Context, id string) (*domain.Application, error) {
	var app domain.Application
	err := r.db.WithContext(ctx).First(&app, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *Repository) GetByApplicantID(ctx context.Context, applicantID string, limit, offset int) ([]domain.Application, error) {
	var apps []domain.Application
	err := r.db.WithContext(ctx).Where("applicant_id = ?", applicantID).Limit(limit).Offset(offset).Order("created_at DESC").Find(&apps).Error
	return apps, err
}

func (r *Repository) GetByWorkflowID(ctx context.Context, workflowID string) ([]domain.Application, error) {
	var apps []domain.Application
	err := r.db.WithContext(ctx).Where("workflow_id = ?", workflowID).Order("created_at DESC").Find(&apps).Error
	return apps, err
}

func (r *Repository) Update(ctx context.Context, app *domain.Application) error {
	return r.db.WithContext(ctx).Save(app).Error
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Application{}).Where("id = ?", id).Update("deleted_at", &now).Error
}

// ListWithPagination returns paginated applications with filters and sorting
func (r *Repository) ListWithPagination(
	ctx context.Context,
	filters *pagination.ApplicationFilters,
	cursor *pagination.Cursor,
	limit int,
) ([]domain.Application, int64, error) {
	var apps []domain.Application
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&domain.Application{})

	// Apply filters
	if filters != nil {
		// Search filter
		if filters.Search != nil && filters.Search.Query != "" {
			searchPattern := "%" + filters.Search.Query + "%"
			query = query.Where(
				"status ILIKE ? OR priority ILIKE ?",
				searchPattern, searchPattern,
			)
		}

		// Status filter
		if filters.Status != nil && len(filters.Status.Statuses) > 0 {
			query = query.Where("status IN ?", filters.Status.Statuses)
		}

		// Applicant filter
		if filters.ApplicantID != "" {
			query = query.Where("applicant_id = ?", filters.ApplicantID)
		}

		// Workflow filter
		if filters.WorkflowID != "" {
			query = query.Where("workflow_id = ?", filters.WorkflowID)
		}

		// Template filter
		if filters.TemplateID != "" {
			query = query.Where("template_id = ?", filters.TemplateID)
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
		"application",
	)
	query = query.Order(pagination.GetSortBySQL(sortConfig, ""))

	// Apply limit + 1 to check if there are more results
	query = query.Limit(limit + 1)

	// Execute query
	if err := query.Find(&apps).Error; err != nil {
		return nil, 0, err
	}

	return apps, totalCount, nil
}

// GetSummary returns application summary statistics
func (r *Repository) GetSummary(ctx context.Context, dateRange *pagination.DateRangeFilter) (*pagination.ApplicationSummary, error) {
	summary := &pagination.ApplicationSummary{
		ByStatus:   make(map[string]int64),
		ByPriority: make(map[string]int64),
	}

	query := r.db.WithContext(ctx).Model(&domain.Application{})

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
	if err := query.Count(&summary.TotalApplications).Error; err != nil {
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

	// Count by priority
	priorityResults := []struct {
		Priority string
		Count    int64
	}{}
	if err := query.Select("priority, COUNT(*) as count").Group("priority").Scan(&priorityResults).Error; err != nil {
		return nil, err
	}
	for _, r := range priorityResults {
		summary.ByPriority[r.Priority] = r.Count
	}

	// Get recent submissions (last 7 days)
	weekAgo := time.Now().AddDate(0, 0, -7)
	if err := r.db.WithContext(ctx).Model(&domain.Application{}).
		Where("created_at >= ?", weekAgo).
		Count(&summary.RecentSubmissions).Error; err != nil {
		return nil, err
	}

	// Calculate average processing time (submitted to completed)
	var avgTime float64
	if err := r.db.WithContext(ctx).
		Select("AVG(EXTRACT(EPOCH FROM (completed_at - submitted_at)) / 3600)").
		Where("completed_at IS NOT NULL AND submitted_at IS NOT NULL").
		Scan(&avgTime).Error; err == nil {
		summary.AvgProcessingTime = avgTime
	}

	return summary, nil
}
