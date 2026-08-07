package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
	"github.com/aeroxe/approval-flow/internal/pkg/messaging"
	"github.com/aeroxe/approval-flow/internal/pkg/websocket"
)

type Service struct {
	Repo   *Repository
	Cache  *cache.Redis
	NATS   *messaging.NATS
	Hub    *websocket.Hub
	Logger *config.Config
}

func NewService(repo *Repository, cache *cache.Redis, nats *messaging.NATS, hub *websocket.Hub, cfg *config.Config) *Service {
	return &Service{Repo: repo, Cache: cache, NATS: nats, Hub: hub, Logger: cfg}
}

func (s *Service) CreateWorkflow(ctx context.Context, workflow *domain.Workflow) error {
	if err := s.Repo.Create(ctx, workflow); err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	s.Cache.Delete(ctx, fmt.Sprintf("workflow:%s", workflow.ID))
	s.NATS.Publish("workflow.created", []byte(fmt.Sprintf(`{"workflow_id":"%s"}`, workflow.ID)))
	s.Logger.Info("workflow created", "workflow_id", workflow.ID)
	return nil
}

func (s *Service) GetWorkflow(ctx context.Context, id string) (*domain.Workflow, error) {
	cacheKey := fmt.Sprintf("workflow:%s", id)
	if cached, err := s.Cache.Get(ctx, cacheKey); err == nil && cached != "" {
		var workflow domain.Workflow
		if err := json.Unmarshal([]byte(cached), &workflow); err == nil {
			return &workflow, nil
		}
	}

	workflow, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(workflow)
	s.Cache.Set(ctx, cacheKey, string(data), 15*time.Minute)

	return workflow, nil
}

func (s *Service) GetWorkflowsByCategory(ctx context.Context, category string) ([]domain.Workflow, error) {
	return s.Repo.GetByCategory(ctx, category)
}

func (s *Service) GetAllWorkflows(ctx context.Context, limit, offset int) ([]domain.Workflow, error) {
	return s.Repo.GetAll(ctx, limit, offset)
}

func (s *Service) AddWorkflowStep(ctx context.Context, step *domain.WorkflowStep) error {
	if err := s.Repo.CreateStep(ctx, step); err != nil {
		return fmt.Errorf("failed to create workflow step: %w", err)
	}
	s.Cache.Delete(ctx, fmt.Sprintf("workflow:%s", step.WorkflowID))
	s.Logger.Info("workflow step created", "step_id", step.ID)
	return nil
}
