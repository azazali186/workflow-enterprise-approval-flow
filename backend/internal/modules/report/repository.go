package report

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

func (r *Repository) Create(ctx context.Context, status *domain.Status) error {
	return r.db.WithContext(ctx).Create(status).Error
}

func (r *Repository) GetByID(ctx context.Context, id string) (*domain.Status, error) {
	var status domain.Status
	err := r.db.WithContext(ctx).First(&status, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *Repository) GetByEntityType(ctx context.Context, entityType string) ([]domain.Status, error) {
	var statuses []domain.Status
	err := r.db.WithContext(ctx).Where("entity_type = ?", entityType).Order("created_at DESC").Find(&statuses).Error
	return statuses, err
}

func (r *Repository) Update(ctx context.Context, status *domain.Status) error {
	return r.db.WithContext(ctx).Save(status).Error
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Status{}).Where("id = ?", id).Update("deleted_at", &now).Error
}

func (r *Repository) CreateComment(ctx context.Context, comment *domain.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *Repository) GetCommentsByApprovalID(ctx context.Context, approvalID string) ([]domain.Comment, error) {
	var comments []domain.Comment
	err := r.db.WithContext(ctx).Where("approval_id = ?", approvalID).Order("created_at ASC").Find(&comments).Error
	return comments, err
}

func (r *Repository) CreateDocument(ctx context.Context, document *domain.Document) error {
	return r.db.WithContext(ctx).Create(document).Error
}

func (r *Repository) GetDocumentsByApplicationID(ctx context.Context, applicationID string) ([]domain.Document, error) {
	var documents []domain.Document
	err := r.db.WithContext(ctx).Where("application_id = ?", applicationID).Order("created_at DESC").Find(&documents).Error
	return documents, err
}
