package notification

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

func (r *Repository) Create(ctx context.Context, notification *domain.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *Repository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.Notification, error) {
	var notifications []domain.Notification
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Limit(limit).Offset(offset).Order("created_at DESC").Find(&notifications).Error
	return notifications, err
}

func (r *Repository) GetUnreadByUserID(ctx context.Context, userID string) ([]domain.Notification, error) {
	var notifications []domain.Notification
	err := r.db.WithContext(ctx).Where("user_id = ? AND read_at IS NULL", userID).Order("created_at DESC").Find(&notifications).Error
	return notifications, err
}

func (r *Repository) MarkAsRead(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Notification{}).Where("id = ?", id).Update("read_at", &now).Error
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Notification{}).Where("id = ?", id).Update("deleted_at", &now).Error
}

// ListWithPagination returns paginated notifications with filters and sorting
func (r *Repository) ListWithPagination(
	ctx context.Context,
	filters *pagination.NotificationFilters,
	cursor *pagination.Cursor,
	limit int,
) ([]domain.Notification, int64, error) {
	var notifications []domain.Notification
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&domain.Notification{})

	// Apply filters
	if filters != nil {
		// Search filter
		if filters.Search != nil && filters.Search.Query != "" {
			searchPattern := "%" + filters.Search.Query + "%"
			query = query.Where(
				"title ILIKE ? OR body ILIKE ? OR type ILIKE ?",
				searchPattern, searchPattern, searchPattern,
			)
		}

		// User filter
		if filters.UserID != "" {
			query = query.Where("user_id = ?", filters.UserID)
		}

		// Type filter
		if filters.Type != "" {
			query = query.Where("type = ?", filters.Type)
		}

		// Channel filter
		if filters.Channel != "" {
			query = query.Where("channel = ?", filters.Channel)
		}

		// ReadAt filter
		if filters.ReadAt != nil {
			if *filters.ReadAt {
				query = query.Where("read_at IS NOT NULL")
			} else {
				query = query.Where("read_at IS NULL")
			}
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
		"notification",
	)
	query = query.Order(pagination.GetSortBySQL(sortConfig, ""))

	// Apply limit + 1 to check if there are more results
	query = query.Limit(limit + 1)

	// Execute query
	if err := query.Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, totalCount, nil
}

// GetSummary returns notification summary statistics
func (r *Repository) GetSummary(ctx context.Context, dateRange *pagination.DateRangeFilter) (*pagination.NotificationSummary, error) {
	summary := &pagination.NotificationSummary{
		ByType:    make(map[string]int64),
		ByChannel: make(map[string]int64),
	}

	query := r.db.WithContext(ctx).Model(&domain.Notification{})

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
	if err := query.Count(&summary.TotalNotifications).Error; err != nil {
		return nil, err
	}

	// Count by type
	typeResults := []struct {
		Type  string
		Count int64
	}{}
	if err := query.Select("type, COUNT(*) as count").Group("type").Scan(&typeResults).Error; err != nil {
		return nil, err
	}
	for _, r := range typeResults {
		summary.ByType[r.Type] = r.Count
	}

	// Count by channel
	channelResults := []struct {
		Channel string
		Count   int64
	}{}
	if err := query.Select("channel, COUNT(*) as count").Group("channel").Scan(&channelResults).Error; err != nil {
		return nil, err
	}
	for _, r := range channelResults {
		summary.ByChannel[r.Channel] = r.Count
	}

	// Count unread notifications
	if err := r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("read_at IS NULL").
		Count(&summary.UnreadCount).Error; err != nil {
		return nil, err
	}

	// Count notifications sent today
	today := time.Now().Truncate(24 * time.Hour)
	if err := r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("created_at >= ?", today).
		Count(&summary.SentToday).Error; err != nil {
		return nil, err
	}

	return summary, nil
}
