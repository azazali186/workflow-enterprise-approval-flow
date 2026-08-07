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

	s.Logger.Info("approval created", "approval_id", approval.ID, "approver_id", approval.ApproverID)
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

	s.NATS.Publish("approval.decided", []byte(fmt.Sprintf(`{"approval_id":"%s","decision":"%s"}`, approval.ID, decision)))
	s.Hub.SendToUser(approval.ApplicationID.String(), "decision_made", map[string]interface{}{
		"approval_id": approval.ID,
		"decision":    decision,
	})

	s.Logger.Info("approval decided", "approval_id", approval.ID, "decision", decision)
	return approval, nil
}
