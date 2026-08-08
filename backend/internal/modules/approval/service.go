package approval

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
	// advanceFn is invoked synchronously after a decision so the workflow
	// engine can route to the next step or complete the application. It is
	// wired by the server to avoid circular module dependencies.
	advanceFn func(ctx context.Context, applicationID, decision, decidedApprovalID string) error
}

func NewService(repo *Repository, cache *cache.Redis, nats *messaging.NATS, hub *websocket.Hub, cfg *config.Config) *Service {
	return &Service{Repo: repo, Cache: cache, NATS: nats, Hub: hub, Logger: cfg}
}

// SetAdvanceHandler registers the post-decision workflow advance callback.
func (s *Service) SetAdvanceHandler(fn func(ctx context.Context, applicationID, decision, decidedApprovalID string) error) {
	s.advanceFn = fn
}

func (s *Service) CreateApproval(ctx context.Context, approval *domain.Approval) error {
	if err := s.Repo.Create(ctx, approval); err != nil {
		return fmt.Errorf("failed to create approval: %w", err)
	}

	s.Cache.Delete(ctx, fmt.Sprintf("approval:%s", approval.ID))

	s.NATS.Publish("approval.created", []byte(fmt.Sprintf(`{"approval_id":"%s"}`, approval.ID)))
	s.Hub.SendToUser(approval.ApproverID.String(), "approval_needed", map[string]interface{}{
		"approval_id": approval.ID,
		"status":      approval.Status,
	})

	s.Logger.Info("approval created",
		zap.String("approval_id", approval.ID.String()),
		zap.String("approver_id", approval.ApproverID.String()),
	)
	return nil
}

func (s *Service) GetApproval(ctx context.Context, id string) (*domain.Approval, error) {
	cacheKey := fmt.Sprintf("approval:%s", id)
	if cached, err := s.Cache.Get(ctx, cacheKey); err == nil && cached != "" {
		var approval domain.Approval
		if err := json.Unmarshal([]byte(cached), &approval); err == nil {
			return &approval, nil
		}
	}

	approval, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(approval)
	s.Cache.Set(ctx, cacheKey, string(data), 15*time.Minute)

	return approval, nil
}

func (s *Service) GetPendingApprovals(ctx context.Context, approverID string) ([]domain.Approval, error) {
	return s.Repo.GetPendingByApproverID(ctx, approverID)
}

// GetApprovalsByApplication returns all approvals for an application.
func (s *Service) GetApprovalsByApplication(ctx context.Context, applicationID string) ([]domain.Approval, error) {
	return s.Repo.GetByApplicationID(ctx, applicationID)
}

// DeleteApproval soft-deletes an approval record.
func (s *Service) DeleteApproval(ctx context.Context, id string) error {
	if err := s.Repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete approval: %w", err)
	}
	s.Cache.Delete(ctx, fmt.Sprintf("approval:%s", id))
	s.Logger.Info("approval deleted", zap.String("approval_id", id))
	return nil
}

func (s *Service) DecideApproval(ctx context.Context, id, decision, comment string) (*domain.Approval, error) {
	approval, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	approval.Status = decision
	approval.Decision = &decision
	approval.Comment = &comment
	approval.DecidedAt = &now

	if err := s.Repo.Update(ctx, approval); err != nil {
		return nil, fmt.Errorf("failed to update approval: %w", err)
	}

	s.Cache.Delete(ctx, fmt.Sprintf("approval:%s", approval.ID))

	s.NATS.Publish("approval.decided", []byte(fmt.Sprintf(`{"approval_id":"%s","application_id":"%s","decision":"%s"}`, approval.ID, approval.ApplicationID, decision)))
	s.Hub.SendToUser(approval.ApplicationID.String(), "decision_made", map[string]interface{}{
		"approval_id": approval.ID,
		"decision":    decision,
	})

	// Advance the workflow (next step or application completion) synchronously
	// so the caller sees a consistent state.
	if s.advanceFn != nil {
		if err := s.advanceFn(ctx, approval.ApplicationID.String(), decision, approval.ID.String()); err != nil {
			s.Logger.Error("workflow advance failed",
				zap.String("approval_id", approval.ID.String()),
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to advance workflow: %w", err)
		}
	}

	s.Logger.Info("approval decided",
		zap.String("approval_id", approval.ID.String()),
		zap.String("decision", decision),
	)
	return approval, nil
}

// ListApprovals returns paginated approvals with filters and sorting
func (s *Service) ListApprovals(
	ctx context.Context,
	filters *pagination.ApprovalFilters,
	cursor *pagination.Cursor,
	limit int,
) (*pagination.ListResponse, error) {
	// Get paginated results
	approvals, totalCount, err := s.Repo.ListWithPagination(ctx, filters, cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list approvals: %w", err)
	}

	// Determine if there are more results
	hasMore := len(approvals) > limit
	if hasMore {
		// Remove the extra item
		approvals = approvals[:limit]
	}

	// Build next cursor
	var nextCursor string
	if hasMore && len(approvals) > 0 {
		lastApproval := approvals[len(approvals)-1]
		nextCursor = pagination.EncodeCursor(&pagination.Cursor{
			ID:        lastApproval.ID.String(),
			CreatedAt: lastApproval.CreatedAt,
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
		Data:       approvals,
		Pagination: paginationResp,
		Summary:    summary,
	}, nil
}

// GetApprovalSummary returns approval summary statistics
func (s *Service) GetApprovalSummary(ctx context.Context, dateRange *pagination.DateRangeFilter) (*pagination.ApprovalSummary, error) {
	return s.Repo.GetSummary(ctx, dateRange)
}
