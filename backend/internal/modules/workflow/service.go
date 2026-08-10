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
	"github.com/aeroxe/approval-flow/internal/pkg/pagination"
	"github.com/aeroxe/approval-flow/internal/pkg/websocket"
	"go.uber.org/zap"
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
	if err := s.NATS.Publish("workflow.created", []byte(fmt.Sprintf(`{"workflow_id":"%s"}`, workflow.ID))); err != nil {
		s.Logger.Error("failed to publish workflow.created", zap.Error(err), zap.String("workflow_id", workflow.ID.String()))
	}
	s.Logger.Info("workflow created", zap.String("workflow_id", workflow.ID.String()))
	return nil
}

// CreateWorkflowWithSteps creates a workflow and its approval steps together.
// If any step fails to persist, the workflow is rolled back so a partial
// failure never leaves an orphan definition.
func (s *Service) CreateWorkflowWithSteps(ctx context.Context, workflow *domain.Workflow, steps []domain.WorkflowStep) error {
	if err := s.CreateWorkflow(ctx, workflow); err != nil {
		return err
	}
	for i := range steps {
		steps[i].WorkflowID = workflow.ID
		if err := s.Repo.CreateStep(ctx, &steps[i]); err != nil {
			_ = s.Repo.Delete(ctx, workflow.ID.String())
			return fmt.Errorf("failed to create workflow step: %w", err)
		}
	}
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

// GetWorkflowSteps returns the ordered approval steps of a workflow.
func (s *Service) GetWorkflowSteps(ctx context.Context, workflowID string) ([]domain.WorkflowStep, error) {
	return s.Repo.GetStepsByWorkflowID(ctx, workflowID)
}

// UpdateWorkflow updates a workflow and, when steps are provided, replaces its
// approval steps wholesale.
func (s *Service) UpdateWorkflow(ctx context.Context, workflow *domain.Workflow, steps []domain.WorkflowStep) error {
	if err := s.Repo.Update(ctx, workflow); err != nil {
		return fmt.Errorf("failed to update workflow: %w", err)
	}
	if steps != nil {
		if err := s.Repo.DeleteStepsByWorkflowID(ctx, workflow.ID.String()); err != nil {
			return fmt.Errorf("failed to clear workflow steps: %w", err)
		}
		for i := range steps {
			steps[i].WorkflowID = workflow.ID
			if err := s.Repo.CreateStep(ctx, &steps[i]); err != nil {
				return fmt.Errorf("failed to replace workflow steps: %w", err)
			}
		}
	}
	s.Cache.Delete(ctx, fmt.Sprintf("workflow:%s", workflow.ID))
	s.Logger.Info("workflow updated", zap.String("workflow_id", workflow.ID.String()))
	return nil
}

// DeleteWorkflow soft-deletes a workflow.
func (s *Service) DeleteWorkflow(ctx context.Context, id string) error {
	if err := s.Repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}
	s.Cache.Delete(ctx, fmt.Sprintf("workflow:%s", id))
	s.Logger.Info("workflow deleted", zap.String("workflow_id", id))
	return nil
}

func (s *Service) AddWorkflowStep(ctx context.Context, step *domain.WorkflowStep) error {
	if err := s.Repo.CreateStep(ctx, step); err != nil {
		return fmt.Errorf("failed to create workflow step: %w", err)
	}
	s.Cache.Delete(ctx, fmt.Sprintf("workflow:%s", step.WorkflowID))
	s.Logger.Info("workflow step created", zap.String("step_id", step.ID.String()))
	return nil
}

// ListWorkflows returns paginated workflows with filters and sorting
func (s *Service) ListWorkflows(
	ctx context.Context,
	filters *pagination.WorkflowFilters,
	cursor *pagination.Cursor,
	limit int,
) (*pagination.ListResponse, error) {
	// Get paginated results
	workflows, totalCount, err := s.Repo.ListWithPagination(ctx, filters, cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflows: %w", err)
	}

	// Determine if there are more results
	hasMore := len(workflows) > limit
	if hasMore {
		// Remove the extra item
		workflows = workflows[:limit]
	}

	// Build next cursor
	var nextCursor string
	if hasMore && len(workflows) > 0 {
		lastWorkflow := workflows[len(workflows)-1]
		nextCursor = pagination.EncodeCursor(&pagination.Cursor{
			ID:        lastWorkflow.ID.String(),
			CreatedAt: lastWorkflow.CreatedAt,
		})
	}

	// Build pagination response
	paginationResp := &pagination.PaginationResponse{
		NextCursor: nextCursor,
		HasMore:    hasMore,
		TotalCount: totalCount,
		PageSize:   limit,
	}

	// Get summary
	summary, err := s.Repo.GetSummary(ctx, filters.DateRange)
	if err != nil {
		s.Logger.Error("failed to get summary", zap.Error(err))
	}

	return &pagination.ListResponse{
		Data:       workflows,
		Pagination: paginationResp,
		Summary:    summary,
	}, nil
}

// GetWorkflowSummary returns workflow summary statistics
func (s *Service) GetWorkflowSummary(ctx context.Context, dateRange *pagination.DateRangeFilter) (*pagination.WorkflowSummary, error) {
	return s.Repo.GetSummary(ctx, dateRange)
}
