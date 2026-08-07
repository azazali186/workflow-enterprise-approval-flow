package report

import (
	"context"
	"fmt"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
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

type Service struct {
	Repo   *Repository
	Cache  *cache.Redis
	Logger *config.Config
}

func NewService(repo *Repository, cache *cache.Redis, cfg *config.Config) *Service {
	return &Service{Repo: repo, Cache: cache, Logger: cfg}
}

func (s *Service) CreateStatus(ctx context.Context, status *domain.Status) error {
	if err := s.Repo.Create(ctx, status); err != nil {
		return fmt.Errorf("failed to create status: %w", err)
	}
	s.Logger.Info("status created", "status_id", status.ID)
	return nil
}

func (s *Service) GetStatus(ctx context.Context, id string) (*domain.Status, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *Service) GetStatusesByEntityType(ctx context.Context, entityType string) ([]domain.Status, error) {
	return s.Repo.GetByEntityType(ctx, entityType)
}

func (s *Service) CreateComment(ctx context.Context, comment *domain.Comment) error {
	if err := s.Repo.CreateComment(ctx, comment); err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}
	s.Logger.Info("comment created", "comment_id", comment.ID)
	return nil
}

func (s *Service) GetComments(ctx context.Context, approvalID string) ([]domain.Comment, error) {
	return s.Repo.GetCommentsByApprovalID(ctx, approvalID)
}

func (s *Service) CreateDocument(ctx context.Context, document *domain.Document) error {
	if err := s.Repo.CreateDocument(ctx, document); err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}
	s.Logger.Info("document created", "document_id", document.ID)
	return nil
}

func (s *Service) GetDocuments(ctx context.Context, applicationID string) ([]domain.Document, error) {
	return s.Repo.GetDocumentsByApplicationID(ctx, applicationID)
}
