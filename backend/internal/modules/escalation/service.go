package escalation

import (
	"context"
	"fmt"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/pkg/messaging"
	"github.com/aeroxe/approval-flow/internal/pkg/websocket"
	"github.com/google/uuid"
)

type Service struct {
	Repo    *Repository
	NATS    *messaging.NATS
	Hub     *websocket.Hub
	Logger  *config.Config
}

func NewService(repo *Repository, nats *messaging.NATS, hub *websocket.Hub, cfg *config.Config) *Service {
	return &Service{Repo: repo, NATS: nats, Hub: hub, Logger: cfg}
}

func (s *Service) Escalate(ctx context.Context, approvalID string, level int, escalatedTo uuid.UUID, reason string) error {
	escalation := &domain.Escalation{
		ApprovalID:   approvalID,
		Level:        level,
		EscalatedTo:  escalatedTo,
		Reason:       reason,
		EscalatedAt:  time.Now(),
	}

	if err := s.Repo.Create(ctx, escalation); err != nil {
		return fmt.Errorf("failed to create escalation: %w", err)
	}

	s.NATS.Publish("escalation.created", []byte(fmt.Sprintf(`{"escalation_id":"%s"}`, escalation.ID)))
	s.Hub.SendToUser(escalatedTo.String(), "escalation_trigger", map[string]interface{}{
		"escalation_id": escalation.ID,
		"approval_id":   approvalID,
		"level":         level,
	})

	s.Logger.Info("escalation created", "escalation_id", escalation.ID, "approval_id", approvalID)
	return nil
}

func (s *Service) GetEscalations(ctx context.Context, approvalID string) ([]domain.Escalation, error) {
	return s.Repo.GetByApprovalID(ctx, approvalID)
}

func (s *Service) GetActiveEscalations(ctx context.Context) ([]domain.Escalation, error) {
	return s.Repo.GetActive(ctx)
}

func (s *Service) ResolveEscalation(ctx context.Context, id string) error {
	if err := s.Repo.Resolve(ctx, id); err != nil {
		return fmt.Errorf("failed to resolve escalation: %w", err)
	}
	s.Logger.Info("escalation resolved", "escalation_id", id)
	return nil
}
