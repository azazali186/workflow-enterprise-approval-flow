package middleware

import (
	"testing"
	"time"

	"github.com/aeroxe/approval-flow/internal/modules/rbac"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
	"github.com/aeroxe/approval-flow/internal/pkg/pagination"
)

func TestIsExcludedRoute(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/health", true},
		{"/ws", true},
		{"/api/v1/auth/login", true},
		{"/api/v1/auth/register", true},
		{"/api/v1/auth/refresh", true},
		{"/metrics", true},
		{"/api/v1/applications", false},
		{"/api/v1/approvals", false},
		{"/api/v1/admin/users", false},
		{"/unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := rbac.IsExcludedRoute(tt.path)
			if result != tt.expected {
				t.Errorf("IsExcludedRoute(%s) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestFormatRouteName(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/ws", "WebSocket Connection"},
		{"/api/v1/applications", "Applications"},
		{"/api/v1/admin/users", "Admin Users"},
		{"/api/v1/auth/login", "Auth Login"},
		{"/api/v1/approvals/pending", "Approvals Pending"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := rbac.FormatRouteName(tt.path)
			if result != tt.expected {
				t.Errorf("FormatRouteName(%s) = %s, want %s", tt.path, result, tt.expected)
			}
		})
	}
}

func TestDistributedLock_Acquire(t *testing.T) {
	// Test that DistributedLock struct is properly initialized
	lock := &DistributedLock{
		key:   "test-lock",
		value: "test-value",
		ttl:   10 * time.Second,
	}

	if lock.key != "test-lock" {
		t.Errorf("expected key 'test-lock', got: %s", lock.key)
	}
	if lock.value != "test-value" {
		t.Errorf("expected value 'test-value', got: %s", lock.value)
	}
	if lock.ttl != 10*time.Second {
		t.Errorf("expected ttl 10s, got: %v", lock.ttl)
	}
	if lock.acquired {
		t.Error("expected acquired to be false initially")
	}
}

func TestCircuitBreaker_States(t *testing.T) {
	// Test circuit breaker state transitions
	cb := NewCircuitBreaker(3, 2, 30*time.Second, nil)

	if cb.GetState() != CircuitClosed {
		t.Errorf("expected state CircuitClosed, got: %v", cb.GetState())
	}

	// Record 3 failures to open
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.GetState() != CircuitOpen {
		t.Errorf("expected state CircuitOpen, got: %v", cb.GetState())
	}

	// Reset
	cb.Reset()
	if cb.GetState() != CircuitClosed {
		t.Errorf("expected state CircuitClosed after reset, got: %v", cb.GetState())
	}
}

func TestRetryConfig_Defaults(t *testing.T) {
	// Test that retry config has sensible defaults
	config := DefaultRetryConfig()

	if config.MaxRetries <= 0 {
		t.Error("expected positive max retries")
	}
	if config.BaseDelay <= 0 {
		t.Error("expected positive base delay")
	}
	if config.Multiplier <= 1 {
		t.Error("expected multiplier > 1")
	}
	if config.MaxDelay <= 0 {
		t.Error("expected positive max delay")
	}
}

func TestBodySizeLimit(t *testing.T) {
	// Test that MaxBodySize is set
	if MaxBodySize <= 0 {
		t.Error("expected positive MaxBodySize")
	}

	// Test that 10MB is the default
	expected := int64(10 * 1024 * 1024)
	if MaxBodySize != expected {
		t.Errorf("expected MaxBodySize %d, got: %d", expected, MaxBodySize)
	}
}

func TestDateRange(t *testing.T) {
	// Test NewDateRange
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	dr := pagination.NewDateRange(&start, &end)

	if dr.Start != start {
		t.Errorf("expected start %v, got: %v", start, dr.Start)
	}
	if dr.End != end {
		t.Errorf("expected end %v, got: %v", end, dr.End)
	}

	// Test IsEmpty
	if dr.IsEmpty() {
		t.Error("expected non-empty date range")
	}

	// Test date range with only start (End = now, not empty)
	startOnly := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dr2 := pagination.NewDateRange(&startOnly, nil)
	if dr2.Start != startOnly {
		t.Errorf("expected start %v, got: %v", startOnly, dr2.Start)
	}
	if dr2.End.IsZero() {
		t.Error("expected non-zero End when nil is passed")
	}
}

func TestCacheStats(t *testing.T) {
	// Test CacheStats struct
	stats := cache.CacheStats{
		Hits:      100,
		Misses:    10,
		Sets:      50,
		Deletes:   5,
		Evictions: 2,
	}

	if stats.Hits != 100 {
		t.Errorf("expected hits 100, got: %d", stats.Hits)
	}
	if stats.Misses != 10 {
		t.Errorf("expected misses 10, got: %d", stats.Misses)
	}
}

func TestParseSortConfig(t *testing.T) {
	tests := []struct {
		name      string
		sortBy    string
		sortOrder string
		entity    string
		wantField string
		wantDir   string
	}{
		{
			name:      "default sort",
			sortBy:    "",
			sortOrder: "",
			entity:    "application",
			wantField: "created_at",
			wantDir:   "desc",
		},
		{
			name:      "valid sort field",
			sortBy:    "status",
			sortOrder: "asc",
			entity:    "application",
			wantField: "status",
			wantDir:   "asc",
		},
		{
			name:      "invalid sort field",
			sortBy:    "invalid_field",
			sortOrder: "desc",
			entity:    "application",
			wantField: "created_at",
			wantDir:   "desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := pagination.ParseSortConfig(tt.sortBy, tt.sortOrder, tt.entity)
			if config.Field != tt.wantField {
				t.Errorf("expected field %s, got: %s", tt.wantField, config.Field)
			}
			if config.Direction != tt.wantDir {
				t.Errorf("expected direction %s, got: %s", tt.wantDir, config.Direction)
			}
		})
	}
}
