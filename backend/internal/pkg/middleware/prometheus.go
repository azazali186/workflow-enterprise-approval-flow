package middleware

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/cloudwego/hertz/pkg/app"
)

// PrometheusMetrics holds comprehensive Prometheus-style metrics
type PrometheusMetrics struct {
	// Counters
	requestsTotal     int64
	errorsTotal       int64
	activeConnections int64

	// Histograms (latency buckets in ms)
	latencyBuckets []float64
	latencyCounts  []int64
	latencySum     int64

	// Gauges
	gaugeValues sync.Map

	// Labels
	requestsByMethod sync.Map
	requestsByStatus sync.Map
	requestsByPath   sync.Map
	errorsByType     sync.Map

	// Start time
	startTime time.Time
}

var (
	prometheusOnce     sync.Once
	prometheusInstance *PrometheusMetrics
)

// GetPrometheusMetrics returns the singleton Prometheus metrics instance
func GetPrometheusMetrics() *PrometheusMetrics {
	prometheusOnce.Do(func() {
		// Define latency buckets in milliseconds: 1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000
		buckets := []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000}
		counts := make([]int64, len(buckets)+1) // +1 for +Inf

		prometheusInstance = &PrometheusMetrics{
			latencyBuckets: buckets,
			latencyCounts:  counts,
			startTime:      time.Now(),
		}
	})
	return prometheusInstance
}

// RecordHTTPRequest records an HTTP request with full metrics
func (pm *PrometheusMetrics) RecordHTTPRequest(method, path string, status int, latencyMs float64) {
	atomic.AddInt64(&pm.requestsTotal, 1)

	// Record by method
	pm.incrementMapCounter(&pm.requestsByMethod, method)

	// Record by status
	statusStr := strconv.Itoa(status)
	pm.incrementMapCounter(&pm.requestsByStatus, statusStr)

	// Record by path (normalized)
	normalizedPath := normalizePath(path)
	pm.incrementMapCounter(&pm.requestsByPath, normalizedPath)

	// Record errors
	if status >= 400 {
		atomic.AddInt64(&pm.errorsTotal, 1)
		pm.incrementMapCounter(&pm.errorsByType, statusStr)
	}

	// Record latency in histogram
	pm.recordLatency(latencyMs)
}

// RecordConnection records connection open/close
func (pm *PrometheusMetrics) RecordConnection(open bool) {
	if open {
		atomic.AddInt64(&pm.activeConnections, 1)
	} else {
		atomic.AddInt64(&pm.activeConnections, -1)
	}
}

// SetGauge sets a gauge value
func (pm *PrometheusMetrics) SetGauge(name string, value float64) {
	pm.gaugeValues.Store(name, value)
}

// GetGauge gets a gauge value
func (pm *PrometheusMetrics) GetGauge(name string) float64 {
	if val, ok := pm.gaugeValues.Load(name); ok {
		return val.(float64)
	}
	return 0
}

// recordLatency records latency in histogram buckets
func (pm *PrometheusMetrics) recordLatency(latencyMs float64) {
	atomic.AddInt64(&pm.latencySum, int64(latencyMs))

	// Find the bucket
	for i, bucket := range pm.latencyBuckets {
		if latencyMs <= bucket {
			atomic.AddInt64(&pm.latencyCounts[i], 1)
			return
		}
	}
	// Overflow bucket (+Inf)
	atomic.AddInt64(&pm.latencyCounts[len(pm.latencyCounts)-1], 1)
}

// incrementMapCounter atomically increments a counter in a sync.Map
func (pm *PrometheusMetrics) incrementMapCounter(m *sync.Map, key string) {
	for {
		val, _ := m.LoadOrStore(key, int64(0))
		old := val.(int64)
		if m.CompareAndSwap(key, old, old+1) {
			break
		}
	}
}

// normalizePath normalizes path for metrics (removes IDs)
func normalizePath(path string) string {
	// Simple normalization - replace UUIDs and numbers with placeholders
	// In production, use a more sophisticated approach
	return path
}

// ToJSON returns metrics as JSON
func (pm *PrometheusMetrics) ToJSON() map[string]interface{} {
	uptime := time.Since(pm.startTime)

	// Calculate percentile latencies
	p50 := pm.calculatePercentile(0.50)
	p95 := pm.calculatePercentile(0.95)
	p99 := pm.calculatePercentile(0.99)

	// Convert sync.Maps to regular maps
	methods := convertSyncMap(&pm.requestsByMethod)
	statuses := convertSyncMap(&pm.requestsByStatus)
	paths := convertSyncMap(&pm.requestsByPath)
	errors := convertSyncMap(&pm.errorsByType)

	return map[string]interface{}{
		"uptime_seconds": uptime.Seconds(),
		"counters": map[string]interface{}{
			"http_requests_total": atomic.LoadInt64(&pm.requestsTotal),
			"http_errors_total":   atomic.LoadInt64(&pm.errorsTotal),
			"active_connections":  atomic.LoadInt64(&pm.activeConnections),
		},
		"histograms": map[string]interface{}{
			"http_request_duration_ms": map[string]interface{}{
				"sum":   atomic.LoadInt64(&pm.latencySum),
				"count": atomic.LoadInt64(&pm.requestsTotal),
				"p50":   p50,
				"p95":   p95,
				"p99":   p99,
			},
		},
		"labels": map[string]interface{}{
			"by_method": methods,
			"by_status": statuses,
			"by_path":   paths,
			"errors":    errors,
		},
		"start_time": pm.startTime.Format(time.RFC3339),
	}
}

// calculatePercentile calculates percentile from histogram
func (pm *PrometheusMetrics) calculatePercentile(percentile float64) float64 {
	totalCount := atomic.LoadInt64(&pm.requestsTotal)
	if totalCount == 0 {
		return 0
	}

	targetCount := int64(math.Ceil(float64(totalCount) * percentile))
	var cumulativeCount int64

	for i, bucket := range pm.latencyBuckets {
		cumulativeCount += atomic.LoadInt64(&pm.latencyCounts[i])
		if cumulativeCount >= targetCount {
			return bucket
		}
	}

	return pm.latencyBuckets[len(pm.latencyBuckets)-1]
}

// convertSyncMap converts sync.Map to regular map
func convertSyncMap(m *sync.Map) map[string]int64 {
	result := make(map[string]int64)
	m.Range(func(key, value interface{}) bool {
		result[key.(string)] = value.(int64)
		return true
	})
	return result
}

// ToPrometheusText renders the metrics in Prometheus text exposition format so a
// real Prometheus scraper can consume them at GET /metrics.
func (pm *PrometheusMetrics) ToPrometheusText() string {
	var b strings.Builder

	b.WriteString("# HELP http_requests_total Total number of HTTP requests processed.\n")
	b.WriteString("# TYPE http_requests_total counter\n")
	fmt.Fprintf(&b, "http_requests_total %d\n", atomic.LoadInt64(&pm.requestsTotal))

	b.WriteString("# HELP http_errors_total Total number of HTTP error responses (status >= 400).\n")
	b.WriteString("# TYPE http_errors_total counter\n")
	fmt.Fprintf(&b, "http_errors_total %d\n", atomic.LoadInt64(&pm.errorsTotal))

	b.WriteString("# HELP http_active_connections Current number of in-flight requests.\n")
	b.WriteString("# TYPE http_active_connections gauge\n")
	fmt.Fprintf(&b, "http_active_connections %d\n", atomic.LoadInt64(&pm.activeConnections))

	b.WriteString("# HELP http_request_duration_ms_sum Total request latency in milliseconds.\n")
	b.WriteString("# TYPE http_request_duration_ms_sum counter\n")
	fmt.Fprintf(&b, "http_request_duration_ms_sum %d\n", atomic.LoadInt64(&pm.latencySum))

	b.WriteString("# HELP http_request_duration_ms_count Total number of requests observed.\n")
	b.WriteString("# TYPE http_request_duration_ms_count counter\n")
	fmt.Fprintf(&b, "http_request_duration_ms_count %d\n", atomic.LoadInt64(&pm.requestsTotal))

	b.WriteString("# HELP http_uptime_seconds Server uptime in seconds.\n")
	b.WriteString("# TYPE http_uptime_seconds gauge\n")
	fmt.Fprintf(&b, "http_uptime_seconds %g\n", time.Since(pm.startTime).Seconds())

	writeLabeled := func(name, help, typ, label string, m *sync.Map) {
		b.WriteString("# HELP " + name + " " + help + "\n")
		b.WriteString("# TYPE " + name + " " + typ + "\n")
		m.Range(func(key, value interface{}) bool {
			fmt.Fprintf(&b, "%s{%s=%q} %d\n", name, label, key.(string), value.(int64))
			return true
		})
	}

	writeLabeled("http_requests_by_method", "HTTP requests by HTTP method.", "counter", "method", &pm.requestsByMethod)
	writeLabeled("http_requests_by_status", "HTTP requests by response status code.", "counter", "status", &pm.requestsByStatus)
	writeLabeled("http_requests_by_path", "HTTP requests by request path.", "counter", "path", &pm.requestsByPath)
	writeLabeled("http_errors_by_status", "HTTP error responses by response status code.", "counter", "status", &pm.errorsByType)

	return b.String()
}

// PrometheusMiddleware collects comprehensive Prometheus-style metrics
func PrometheusMiddleware(cfg *config.Config) app.HandlerFunc {
	metrics := GetPrometheusMetrics()

	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()

		// Track connection
		metrics.RecordConnection(true)
		defer metrics.RecordConnection(false)

		c.Next(ctx)

		latencyMs := float64(time.Since(start).Microseconds()) / 1000.0
		status := c.Response.StatusCode()
		method := string(c.Request.Method())
		path := string(c.Request.URI().Path())

		metrics.RecordHTTPRequest(method, path, status, latencyMs)
	}
}
