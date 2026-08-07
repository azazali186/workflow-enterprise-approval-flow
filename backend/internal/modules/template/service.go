package template

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/modules/template/repository"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
	"github.com/aeroxe/approval-flow/internal/pkg/messaging"
)

type Service struct {
	Repo   *repository.Repository
	Cache  *cache.Redis
	NATS   *messaging.NATS
	Logger *config.Config
}

func NewService(repo *repository.Repository, cache *cache.Redis, nats *messaging.NATS, cfg *config.Config) *Service {
	return &Service{Repo: repo, Cache: cache, NATS: nats, Logger: cfg}
}

func (s *Service) CreateTemplate(ctx context.Context, template *domain.Template) error {
	if err := s.Repo.Create(ctx, template); err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	s.Cache.Delete(ctx, fmt.Sprintf("template:%s", template.ID))
	s.NATS.Publish("template.created", []byte(fmt.Sprintf(`{"template_id":"%s"}`, template.ID)))
	s.Logger.Info("template created", "template_id", template.ID)
	return nil
}

func (s *Service) GetTemplate(ctx context.Context, id string) (*domain.Template, error) {
	cacheKey := fmt.Sprintf("template:%s", id)
	if cached, err := s.Cache.Get(ctx, cacheKey); err == nil && cached != "" {
		var template domain.Template
		if err := json.Unmarshal([]byte(cached), &template); err == nil {
			return &template, nil
		}
	}

	template, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(template)
	s.Cache.Set(ctx, cacheKey, string(data), 15*time.Minute)

	return template, nil
}

func (s *Service) GetTemplatesByCategory(ctx context.Context, category string) ([]domain.Template, error) {
	return s.Repo.GetByCategory(ctx, category)
}

func (s *Service) GetAllTemplates(ctx context.Context, limit, offset int) ([]domain.Template, error) {
	return s.Repo.GetAll(ctx, limit, offset)
}
