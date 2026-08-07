package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
	"github.com/aeroxe/approval-flow/internal/pkg/messaging"
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
