package escalation

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

func (r *Repository) Create(ctx context.Context, escalation *domain.Escalation) error {
	return r.db.WithContext(ctx).Create(escalation).Error
}

func (r *Repository) GetByApprovalID(ctx context.Context, approvalID string) ([]domain.Escalation, error) {
	var escalations []domain.Escalation
	err := r.db.WithContext(ctx).Where("approval_id = ?", approvalID).Order("level ASC").Find(&escalations).Error
	return escalations, err
}

func (r *Repository) GetActive(ctx context.Context) ([]domain.Escalation, error) {
	var escalations []domain.Escalation
	err := r.db.WithContext(ctx).Where("resolved_at IS NULL").Find(&escalations).Error
	return escalations, err
}

func (r *Repository) Resolve(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Escalation{}).Where("id = ?", id).Update("resolved_at", &now).Error
}
