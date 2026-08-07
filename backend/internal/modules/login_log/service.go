package login_log

import (
	"context"
	"fmt"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles login log business logic
type Service struct {
	Repo   *Repository
	Cache  *cache.Redis
	Logger *config.Config
}

// NewService creates a new login log service
func NewService(repo *Repository, cache *cache.Redis, cfg *config.Config) *Service {
	return &Service{
		Repo:   repo,
		Cache:  cache,
		Logger: cfg,
	}
}

// LogLoginAttempt logs a login attempt (success or failure)
func (s *Service) LogLoginAttempt(ctx context.Context, params LogLoginParams) error {
	log := &domain.LoginLog{
		UserID:      params.UserID,
		Email:       params.Email,
		Status:      params.Status,
		IPAddress:   params.IPAddress,
		UserAgent:   params.UserAgent,
		AttemptedAt: time.Now(),
	}
	if params.RequestID != "" {
		log.RequestID = &params.RequestID
	}
	if params.TokenID != "" {
		log.TokenID = &params.TokenID
	}

	if params.FailureReason != "" {
		log.FailureReason = &params.FailureReason
	}

	if err := s.Repo.Create(ctx, log); err != nil {
		s.Logger.Error("failed to create login log",
			zap.Error(err),
			zap.String("email", params.Email),
			zap.String("status", params.Status),
		)
		return fmt.Errorf("failed to log login attempt: %w", err)
	}

	s.Logger.Info("login attempt logged",
		zap.String("email", params.Email),
		zap.String("status", params.Status),
		zap.String("ip_address", params.IPAddress),
	)

	return nil
}

// LogLoginSuccess logs a successful login
func (s *Service) LogLoginSuccess(ctx context.Context, userID uuid.UUID, email, ipAddress, userAgent, requestID, tokenID string) error {
	return s.LogLoginAttempt(ctx, LogLoginParams{
		UserID:    &userID,
		Email:     email,
		Status:    domain.LoginStatusSuccess,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		RequestID: requestID,
		TokenID:   tokenID,
	})
}

// LogLoginFailure logs a failed login
func (s *Service) LogLoginFailure(ctx context.Context, email, failureReason, ipAddress, userAgent, requestID string) error {
	return s.LogLoginAttempt(ctx, LogLoginParams{
		Email:         email,
		Status:        domain.LoginStatusFailed,
		FailureReason: failureReason,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		RequestID:     requestID,
	})
}

// LogAccountLocked logs when an account is locked due to too many attempts
func (s *Service) LogAccountLocked(ctx context.Context, email, ipAddress, userAgent, requestID string) error {
	return s.LogLoginAttempt(ctx, LogLoginParams{
		Email:         email,
		Status:        domain.LoginStatusLocked,
		FailureReason: domain.FailureReasonTooManyAttempts,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		RequestID:     requestID,
	})
}

// IsAccountLocked checks if an account is locked due to too many failed attempts
// Returns true if more than 5 failed attempts in the last 15 minutes
func (s *Service) IsAccountLocked(ctx context.Context, email string) (bool, error) {
	since := time.Now().Add(-15 * time.Minute)
	count, err := s.Repo.GetRecentFailedAttempts(ctx, email, since)
	if err != nil {
		return false, err
	}
	return count >= 5, nil
}

// IsIPLocked checks if an IP is locked due to too many failed attempts
// Returns true if more than 20 failed attempts from this IP in the last 15 minutes
func (s *Service) IsIPLocked(ctx context.Context, ipAddress string) (bool, error) {
	since := time.Now().Add(-15 * time.Minute)
	count, err := s.Repo.GetRecentFailedAttemptsByIP(ctx, ipAddress, since)
	if err != nil {
		return false, err
	}
	return count >= 20, nil
}

// GetLoginHistory retrieves login history for a user
func (s *Service) GetLoginHistory(ctx context.Context, userID string, limit int) ([]domain.LoginLog, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return s.Repo.GetByUserID(ctx, userID, limit)
}

// GetLoginHistoryByEmail retrieves login history by email
func (s *Service) GetLoginHistoryByEmail(ctx context.Context, email string, limit int) ([]domain.LoginLog, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return s.Repo.GetByEmail(ctx, email, limit)
}

// GetLoginStats returns login statistics for a user
func (s *Service) GetLoginStats(ctx context.Context, userID string) (map[string]interface{}, error) {
	return s.Repo.GetLoginStats(ctx, userID)
}

// GetLoginLogs retrieves login logs with filters and pagination
func (s *Service) GetLoginLogs(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]domain.LoginLog, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return s.Repo.GetLoginLogs(ctx, limit, offset, filters)
}

// LogLoginParams holds parameters for logging a login attempt
type LogLoginParams struct {
	UserID        *uuid.UUID
	Email         string
	Status        string
	FailureReason string
	IPAddress     string
	UserAgent     string
	RequestID     string
	TokenID       string
}
