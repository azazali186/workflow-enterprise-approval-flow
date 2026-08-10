package middleware

import (
	"strings"
	"testing"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestIsOriginAllowed(t *testing.T) {
	tests := []struct {
		name     string
		origin   string
		allowed  []string
		expected bool
	}{
		{"exact match", "https://app.example.com", []string{"https://app.example.com"}, true},
		{"wildcard allows everything", "https://evil.example.com", []string{"*"}, true},
		{"subdomain wildcard", "https://app.example.com", []string{"*.example.com"}, true},
		{"subdomain wildcard rejects other domain", "https://example.com", []string{"*.example.com"}, false},
		{"not in allow-list", "https://attacker.example.net", []string{"https://app.example.com"}, false},
		{"empty allow-list rejects all", "https://app.example.com", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsOriginAllowed(tt.origin, tt.allowed))
		})
	}
}

func TestEffectiveCORSOrigins(t *testing.T) {
	assert.Equal(t, []string{"*"}, EffectiveCORSOrigins(nil), "nil list falls back to wildcard")
	assert.Equal(t, []string{"*"}, EffectiveCORSOrigins([]string{}), "empty list falls back to wildcard")
	assert.Equal(t, []string{"https://app.example.com"}, EffectiveCORSOrigins([]string{"https://app.example.com"}), "explicit list is preserved")
}

func TestNewCORSConfig(t *testing.T) {
	t.Run("wildcard disables credentials", func(t *testing.T) {
		cfg := &config.Config{CORSAllowedOrigins: []string{"*"}}
		cors := NewCORSConfig(cfg)
		assert.False(t, cors.AllowCredentials, "credentials must be disabled with a wildcard allow-list")
	})

	t.Run("explicit allow-list enables credentials", func(t *testing.T) {
		cfg := &config.Config{CORSAllowedOrigins: []string{"https://app.example.com", "https://admin.example.com"}}
		cors := NewCORSConfig(cfg)
		assert.True(t, cors.AllowCredentials, "credentials are safe with an explicit allow-list")
	})

	t.Run("empty config falls back to wildcard", func(t *testing.T) {
		cfg := &config.Config{}
		cors := NewCORSConfig(cfg)
		assert.Equal(t, []string{"*"}, cors.AllowOrigins)
	})
}

func TestIsCircuitExcludedPath(t *testing.T) {
	excluded := []string{
		"/health",
		"/health/live",
		"/health/ready",
		"/metrics",
		"/docs",
		"/docs/swagger.json",
		"/version",
		"/ws",
	}
	for _, path := range excluded {
		t.Run(path, func(t *testing.T) {
			assert.True(t, isCircuitExcludedPath(path), "infra path %s must bypass the circuit breaker", path)
		})
	}

	included := []string{
		"/api/v1/applications",
		"/api/v1/approvals/decide",
		"/api/v1/admin/users",
	}
	for _, path := range included {
		t.Run(path, func(t *testing.T) {
			assert.False(t, isCircuitExcludedPath(path), "API path %s must be covered by the circuit breaker", path)
		})
	}
}

func TestPrometheusTextFormat(t *testing.T) {
	metrics := GetPrometheusMetrics()

	// Seed a couple of samples so the labeled series are present in the output.
	metrics.RecordHTTPRequest("GET", "/health/live", 200, 1.5)
	metrics.RecordHTTPRequest("POST", "/api/v1/applications", 500, 42.0)

	text := metrics.ToPrometheusText()

	for _, want := range []string{
		"# HELP http_requests_total",
		"# TYPE http_requests_total counter",
		"http_requests_total ",
		"# TYPE http_active_connections gauge",
		"http_request_duration_ms_sum ",
		"# TYPE http_uptime_seconds gauge",
		"http_requests_by_method{method=\"GET\"}",
		"http_requests_by_status{status=\"500\"}",
		"http_requests_by_path{path=",
	} {
		assert.True(t, strings.Contains(text, want), "Prometheus text output should contain %q", want)
	}
}
