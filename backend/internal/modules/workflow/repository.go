package workflow

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
