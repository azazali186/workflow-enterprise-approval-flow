package notification

import (
	"context"
	"fmt"
	"strings"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
	"github.com/aeroxe/approval-flow/internal/pkg/email"
	"github.com/aeroxe/approval-flow/internal/pkg/messaging"
	"github.com/aeroxe/approval-flow/internal/pkg/pagination"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service struct {
	Repo        *Repository
	Cache       *cache.Redis
	NATS        *messaging.NATS
	Logger      *config.Config
	db          *gorm.DB
	emailSender *email.Sender
}

func NewService(repo *Repository, cache *cache.Redis, nats *messaging.NATS, db *gorm.DB, cfg *config.Config) *Service {
	return &Service{Repo: repo, Cache: cache, NATS: nats, Logger: cfg, db: db, emailSender: email.New(cfg)}
}

func (s *Service) SendNotification(ctx context.Context, notification *domain.Notification) error {
	if err := s.Repo.Create(ctx, notification); err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Include the recipient so the saga can broadcast to the right user.
	payload := fmt.Sprintf(`{"notification_id":"%s","user_id":"%s","type":"%s"}`,
		notification.ID, notification.UserID, notification.Type)
	if err := s.NATS.Publish("notification.created", []byte(payload)); err != nil {
		s.Logger.Error("failed to publish notification.created", zap.Error(err), zap.String("notification_id", notification.ID.String()))
	}

	if notification.Channel == "email" || strings.Contains(notification.Channel, "email") {
		s.deliverEmail(ctx, notification)
	}

	s.Logger.Info("notification created",
		zap.String("notification_id", notification.ID.String()),
		zap.String("user_id", notification.UserID.String()),
		zap.String("channel", notification.Channel),
	)
	return nil
}

// deliverEmail resolves the recipient's email and sends it. Email delivery is
// best-effort: an unconfigured SMTP server or a failed send is logged, never
// fatal to the request (the in-app row is already persisted).
func (s *Service) deliverEmail(ctx context.Context, notification *domain.Notification) {
	if s.db == nil || s.emailSender == nil || !s.emailSender.Configured() {
		s.Logger.Warn("email notification skipped: SMTP not configured",
			zap.String("notification_id", notification.ID.String()),
		)
		return
	}

	var user domain.User
	if err := s.db.WithContext(ctx).First(&user, "id = ?", notification.UserID).Error; err != nil {
		s.Logger.Error("email notification skipped: recipient not found", zap.Error(err))
		return
	}

	if err := s.emailSender.Send(user.Email, notification.Title, notification.Body); err != nil {
		s.Logger.Error("email notification failed",
			zap.String("notification_id", notification.ID.String()),
			zap.String("recipient", user.Email),
			zap.Error(err),
		)
	}
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
		NextCursor: nextCursor,
		HasMore:    hasMore,
		TotalCount: totalCount,
		PageSize:   limit,
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
