package approval

import (
	"context"
	"fmt"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
	"github.com/aeroxe/approval-flow/internal/pkg/messaging"
	"github.com/aeroxe/approval-flow/internal/pkg/websocket"
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

func (r *Repository) Delete(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Approval{}).Where("id = ?", id).Update("deleted_at", &now).Error
}
