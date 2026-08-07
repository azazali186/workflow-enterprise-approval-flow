package report

import (
	"context"
	"fmt"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/modules/report/repository"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
)

type Service struct {
	Repo   *repository.Repository
	Cache  *cache.Redis
	Logger *config.Config
}

func NewService(repo *repository.Repository, cache *cache.Redis, cfg *config.Config) *Service {
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
