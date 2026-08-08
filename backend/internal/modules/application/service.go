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
	"github.com/aeroxe/approval-flow/internal/pkg/pagination"
	"github.com/aeroxe/approval-flow/internal/pkg/websocket"
	"go.uber.org/zap"
)

type Service struct {
	Repo        *Repository
	Cache       *cache.Redis
	NATS        *messaging.NATS
	Hub         *websocket.Hub
	ApprovalSvc *approval.Service
	Logger      *config.Config
	// onSubmitted is invoked synchronously after a successful submission so the
	// workflow engine can create the first approval task. Wired by the server
	// to avoid circular module dependencies.
	onSubmitted func(ctx context.Context, applicationID string) error
}

func NewService(repo *Repository, cache *cache.Redis, nats *messaging.NATS, hub *websocket.Hub, approvalSvc *approval.Service, cfg *config.Config) *Service {
	return &Service{Repo: repo, Cache: cache, NATS: nats, Hub: hub, ApprovalSvc: approvalSvc, Logger: cfg}
}

// SetOnSubmitted registers the post-submission workflow routing callback.
func (s *Service) SetOnSubmitted(fn func(ctx context.Context, applicationID string) error) {
	s.onSubmitted = fn
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

	s.Logger.Info("application submitted",
		zap.String("application_id", app.ID.String()),
		zap.String("applicant_id", app.ApplicantID.String()),
	)

	if s.onSubmitted != nil {
		if err := s.onSubmitted(ctx, app.ID.String()); err != nil {
			return fmt.Errorf("failed to route application to approval: %w", err)
		}
	}
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
	s.Logger.Info("application updated", zap.String("application_id", app.ID.String()))
	return nil
}

// ListApplications returns paginated applications with filters and sorting
func (s *Service) ListApplications(
	ctx context.Context,
	filters *pagination.ApplicationFilters,
	cursor *pagination.Cursor,
	limit int,
) (*pagination.ListResponse, error) {
	// Get paginated results
	apps, totalCount, err := s.Repo.ListWithPagination(ctx, filters, cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list applications: %w", err)
	}

	// Determine if there are more results
	hasMore := len(apps) > limit
	if hasMore {
		// Remove the extra item
		apps = apps[:limit]
	}

	// Build next cursor
	var nextCursor string
	if hasMore && len(apps) > 0 {
		lastApp := apps[len(apps)-1]
		nextCursor = pagination.EncodeCursor(&pagination.Cursor{
			ID:        lastApp.ID.String(),
			CreatedAt: lastApp.CreatedAt,
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
		Data:       apps,
		Pagination: paginationResp,
		Summary:    summary,
	}, nil
}

// GetApplicationSummary returns application summary statistics
func (s *Service) GetApplicationSummary(ctx context.Context, dateRange *pagination.DateRangeFilter) (*pagination.ApplicationSummary, error) {
	return s.Repo.GetSummary(ctx, dateRange)
}
