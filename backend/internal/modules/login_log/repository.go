package login_log

import (
	"context"
	"time"

	"github.com/aeroxe/approval-flow/internal/domain"
	"gorm.io/gorm"
)

// Repository handles all login log database operations
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new login log repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new login log entry
func (r *Repository) Create(ctx context.Context, log *domain.LoginLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetByEmail retrieves login logs by email
func (r *Repository) GetByEmail(ctx context.Context, email string, limit int) ([]domain.LoginLog, error) {
	var logs []domain.LoginLog
	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		Order("attempted_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// GetByUserID retrieves login logs by user ID
func (r *Repository) GetByUserID(ctx context.Context, userID string, limit int) ([]domain.LoginLog, error) {
	var logs []domain.LoginLog
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("attempted_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// GetRecentFailedAttempts counts failed login attempts since a given time
func (r *Repository) GetRecentFailedAttempts(ctx context.Context, email string, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.LoginLog{}).
		Where("email = ? AND status = ? AND attempted_at >= ?", email, domain.LoginStatusFailed, since).
		Count(&count).Error
	return count, err
}

// GetRecentFailedAttemptsByIP counts failed login attempts from an IP since a given time
func (r *Repository) GetRecentFailedAttemptsByIP(ctx context.Context, ipAddress string, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.LoginLog{}).
		Where("ip_address = ? AND status = ? AND attempted_at >= ?", ipAddress, domain.LoginStatusFailed, since).
		Count(&count).Error
	return count, err
}

// GetSuccessfulLogin retrieves the most recent successful login for a user
func (r *Repository) GetSuccessfulLogin(ctx context.Context, userID string) (*domain.LoginLog, error) {
	var log domain.LoginLog
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, domain.LoginStatusSuccess).
		Order("attempted_at DESC").
		First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// GetLoginStats returns login statistics for a user
func (r *Repository) GetLoginStats(ctx context.Context, userID string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	var totalLogins, successfulLogins, failedLogins int64
	err := r.db.WithContext(ctx).
		Model(&domain.LoginLog{}).
		Where("user_id = ?", userID).
		Count(&totalLogins).Error
	if err != nil {
		return nil, err
	}

	r.db.WithContext(ctx).Model(&domain.LoginLog{}).
		Where("user_id = ? AND status = ?", userID, domain.LoginStatusSuccess).
		Count(&successfulLogins)

	r.db.WithContext(ctx).Model(&domain.LoginLog{}).
		Where("user_id = ? AND status = ?", userID, domain.LoginStatusFailed).
		Count(&failedLogins)

	// Get last successful login
	var lastLogin domain.LoginLog
	err = r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, domain.LoginStatusSuccess).
		Order("attempted_at DESC").
		First(&lastLogin).Error
	if err == nil {
		result["last_successful_login"] = lastLogin.AttemptedAt
	}

	result["total_logins"] = totalLogins
	result["successful_logins"] = successfulLogins
	result["failed_logins"] = failedLogins

	return result, nil
}

// GetLoginLogs retrieves login logs with pagination
func (r *Repository) GetLoginLogs(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]domain.LoginLog, int64, error) {
	var logs []domain.LoginLog
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&domain.LoginLog{})

	// Apply filters
	if email, ok := filters["email"]; ok && email != "" {
		query = query.Where("email = ?", email)
	}
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if userID, ok := filters["user_id"]; ok && userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if startDate, ok := filters["start_date"]; ok && startDate != "" {
		query = query.Where("attempted_at >= ?", startDate)
	}
	if endDate, ok := filters["end_date"]; ok && endDate != "" {
		query = query.Where("attempted_at <= ?", endDate)
	}

	// Get total count
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.
		Order("attempted_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error

	return logs, totalCount, err
}
