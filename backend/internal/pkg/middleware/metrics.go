package middleware

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/cloudwego/hertz/pkg/app"
)

// Metrics holds application metrics
type Metrics struct {
	mu sync.RWMutex

	// Request metrics
	TotalRequests      int64
	RequestsByMethod   map[string]int64
	RequestsByStatus   map[string]int64
	RequestsByPath     map[string]int64

	// Latency metrics
	RequestLatencies   []time.Duration
	AverageLatency     time.Duration

	// Error metrics
	TotalErrors        int64
	ErrorsByType       map[string]int64

	// Active connections
	ActiveConnections  int64

	// Start time
	StartTime          time.Time
}

var (
	globalMetrics *Metrics
	metricsOnce   sync.Once
)

// GetMetrics returns the global metrics instance
func GetMetrics() *Metrics {
	metricsOnce.Do(func() {
		globalMetrics = &Metrics{
			RequestsByMethod: make(map[string]int64),
			RequestsByStatus: make(map[string]int64),
			RequestsByPath:   make(map[string]int64),
			ErrorsByType:     make(map[string]int64),
			StartTime:        time.Now(),
		}
	})
	return globalMetrics
}

// RecordRequest records a request
func (m *Metrics) RecordRequest(method, path string, status int, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests++
	m.RequestsByMethod[method]++
	m.RequestsByStatus[strconv.Itoa(status)]++
	m.RequestsByPath[path]++

	// Update latency
	m.RequestLatencies = append(m.RequestLatencies, latency)
	if len(m.RequestLatencies) > 1000 {
		m.RequestLatencies = m.RequestLatencies[len(m.RequestLatencies)-1000:]
	}

	// Calculate average
	var total time.Duration
	for _, l := range m.RequestLatencies {
		total += l
	}
	m.AverageLatency = total / time.Duration(len(m.RequestLatencies))

	// Record errors
	if status >= 400 {
		m.TotalErrors++
		m.ErrorsByType[strconv.Itoa(status)]++
	}
}

// MetricsMiddleware collects request metrics
func MetricsMiddleware(cfg *config.Config) app.HandlerFunc {
	metrics := GetMetrics()

	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()

		c.Next(ctx)

		latency := time.Since(start)
		status := c.Response.StatusCode()
		method := string(c.Request.Method())
		path := string(c.Request.URI().Path())

		metrics.RecordRequest(method, path, status, latency)
	}
}

// GetMetricsJSON returns metrics as JSON
func GetMetricsJSON() map[string]interface{} {
	metrics := GetMetrics()
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()

	uptime := time.Since(metrics.StartTime)

	return map[string]interface{}{
		"uptime_seconds":     uptime.Seconds(),
		"total_requests":     metrics.TotalRequests,
		"requests_by_method": metrics.RequestsByMethod,
		"requests_by_status": metrics.RequestsByStatus,
		"average_latency_ms": metrics.AverageLatency.Milliseconds(),
		"total_errors":       metrics.TotalErrors,
		"errors_by_type":     metrics.ErrorsByType,
		"start_time":         metrics.StartTime.Format(time.RFC3339),
	}
}
