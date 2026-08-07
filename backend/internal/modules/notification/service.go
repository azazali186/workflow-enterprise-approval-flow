package notification

import (
	"context"
	"fmt"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
	"github.com/aeroxe/approval-flow/internal/pkg/messaging"
	"github.com/aeroxe/approval-flow/internal/pkg/pagination"
	"go.uber.org/zap"
)

type Service struct {
	Repo   *Repository
	Cache  *cache.Redis
	NATS   *messaging.NATS
	Logger *config.Config
}

func NewService(repo *Repository, cache *cache.Redis, nats *messaging.NATS, cfg *config.Config) *Service {
	return &Service{Repo: repo, Cache: cache, NATS: nats, Logger: cfg}
}

func (s *Service) SendNotification(ctx context.Context, notification *domain.Notification) error {
	if err := s.Repo.Create(ctx, notification); err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	s.NATS.Publish("notification.created", []byte(fmt.Sprintf(`{"notification_id":"%s"}`, notification.ID)))
	s.Logger.Info("notification created",
		zap.String("notification_id", notification.ID.String()),
		zap.String("user_id", notification.UserID.String()),
	)
	return nil
}

func (s *Service) GetUserNotifications(ctx context.Context, userID string, limit, offset int) ([]domain.Notification, error) {
	return s.Repo.GetByUserID(ctx, userID, limit, offset)
}

func (s *Service) GetUnreadNotifications(ctx context.Context, userID string) ([]domain.Notification, error) {
	return s.Repo.GetUnreadByUserID(ctx, userID)
}

func (s *Service) MarkAsRead(ctx context.Context, id string) error {
	if err := s.Repo.MarkAsRead(ctx, id); err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	s.Cache.Delete(ctx, fmt.Sprintf("notification:%s", id))
	s.Logger.Info("notification marked as read", zap.String("notification_id", id))
	return nil
}

// ListNotifications returns paginated notifications with filters and sorting
func (s *Service) ListNotifications(
	ctx context.Context,
	filters *pagination.NotificationFilters,
	cursor *pagination.Cursor,
	limit int,
) (*pagination.ListResponse, error) {
	// Get paginated results
	notifications, totalCount, err := s.Repo.ListWithPagination(ctx, filters, cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}

	// Determine if there are more results
	hasMore := len(notifications) > limit
	if hasMore {
		// Remove the extra item
		notifications = notifications[:limit]
	}

	// Build next cursor
	var nextCursor string
	if hasMore && len(notifications) > 0 {
		lastNotification := notifications[len(notifications)-1]
		nextCursor = pagination.EncodeCursor(&pagination.Cursor{
			ID:        lastNotification.ID.String(),
			CreatedAt: lastNotification.CreatedAt,
		})
	}

	// Build pagination response
	paginationResp := &pagination.PaginationResponse{
		NextCursor:    nextCursor,
		HasMore:       hasMore,
		TotalCount:    totalCount,
		PageSize:      limit,
	}

	// Get summary
	summary, err := s.Repo.GetSummary(ctx, filters.DateRange)
	if err != nil {
		s.Logger.Error("failed to get summary", zap.Error(err))
	}

	return &pagination.ListResponse{
		Data:       notifications,
		Pagination: paginationResp,
		Summary:    summary,
	}, nil
}

// GetNotificationSummary returns notification summary statistics
func (s *Service) GetNotificationSummary(ctx context.Context, dateRange *pagination.DateRangeFilter) (*pagination.NotificationSummary, error) {
	return s.Repo.GetSummary(ctx, dateRange)
}
