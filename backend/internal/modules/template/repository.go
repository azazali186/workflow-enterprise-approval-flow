package template

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

func (r *Repository) Create(ctx context.Context, template *domain.Template) error {
	return r.db.WithContext(ctx).Create(template).Error
}

func (r *Repository) GetByID(ctx context.Context, id string) (*domain.Template, error) {
	var template domain.Template
	err := r.db.WithContext(ctx).First(&template, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *Repository) GetByCategory(ctx context.Context, category string) ([]domain.Template, error) {
	var templates []domain.Template
	err := r.db.WithContext(ctx).Where("category = ? AND is_active = ?", category, true).Order("created_at DESC").Find(&templates).Error
	return templates, err
}

func (r *Repository) GetAll(ctx context.Context, limit, offset int) ([]domain.Template, error) {
	var templates []domain.Template
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Order("created_at DESC").Find(&templates).Error
	return templates, err
}

func (r *Repository) Update(ctx context.Context, template *domain.Template) error {
	return r.db.WithContext(ctx).Save(template).Error
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Template{}).Where("id = ?", id).Update("deleted_at", &now).Error
}
