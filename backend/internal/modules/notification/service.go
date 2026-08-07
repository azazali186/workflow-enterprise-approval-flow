package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
	"github.com/aeroxe/approval-flow/internal/pkg/messaging"
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
	s.Logger.Info("notification created", "notification_id", notification.ID, "user_id", notification.UserID)
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
	s.Logger.Info("notification marked as read", "notification_id", id)
	return nil
}
