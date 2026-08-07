package application

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
