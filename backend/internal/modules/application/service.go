package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/modules/approval"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
	"github.com/aeroxe/approval-flow/internal/pkg/messaging"
	"github.com/aeroxe/approval-flow/internal/pkg/websocket"
)

type Service struct {
	Repo      *Repository
	Cache     *cache.Redis
	NATS      *messaging.NATS
	Hub       *websocket.Hub
	ApprovalSvc *approval.Service
	Logger    *config.Config
}

func NewService(repo *Repository, cache *cache.Redis, nats *messaging.NATS, hub *websocket.Hub, approvalSvc *approval.Service, cfg *config.Config) *Service {
	return &Service{Repo: repo, Cache: cache, NATS: nats, Hub: hub, ApprovalSvc: approvalSvc, Logger: cfg}
}

func (s *Service) SubmitApplication(ctx context.Context, app *domain.Application) error {
	if err := s.Repo.Create(ctx, app); err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}

	s.Cache.Delete(ctx, fmt.Sprintf("application:%s", app.ID))

	s.NATS.Publish("application.submitted", []byte(fmt.Sprintf(`{"application_id":"%s"}`, app.ID)))
	s.Hub.SendToUser(app.ApplicantID.String(), "application_submitted", map[string]interface{}{
		"application_id": app.ID,
		"status":         app.Status,
	})

	s.Logger.Info("application submitted", "application_id", app.ID, "applicant_id", app.ApplicantID)
	return nil
}

func (s *Service) GetApplication(ctx context.Context, id string) (*domain.Application, error) {
	cacheKey := fmt.Sprintf("application:%s", id)
	if cached, err := s.Cache.Get(ctx, cacheKey); err == nil && cached != "" {
		var app domain.Application
		if err := json.Unmarshal([]byte(cached), &app); err == nil {
			return &app, nil
		}
	}

	app, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(app)
	s.Cache.Set(ctx, cacheKey, string(data), 15*time.Minute)

	return app, nil
}

func (s *Service) GetApplicationsByApplicant(ctx context.Context, applicantID string, limit, offset int) ([]domain.Application, error) {
	return s.Repo.GetByApplicantID(ctx, applicantID, limit, offset)
}

func (s *Service) UpdateApplication(ctx context.Context, app *domain.Application) error {
	if err := s.Repo.Update(ctx, app); err != nil {
		return fmt.Errorf("failed to update application: %w", err)
	}

	s.Cache.Delete(ctx, fmt.Sprintf("application:%s", app.ID))
	s.Logger.Info("application updated", "application_id", app.ID)
	return nil
}
