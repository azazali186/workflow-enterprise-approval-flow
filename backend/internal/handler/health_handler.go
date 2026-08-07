package handler

import (
	"context"
	"runtime"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
	"github.com/aeroxe/approval-flow/internal/pkg/database"
	"github.com/aeroxe/approval-flow/internal/pkg/messaging"
	"github.com/aeroxe/approval-flow/internal/pkg/middleware"
	"github.com/aeroxe/approval-flow/internal/pkg/response"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"go.uber.org/zap"
)

type HealthHandler struct {
	db    *database.DB
	redis *cache.Redis
	nats  *messaging.NATS
	cfg   *config.Config
}

func NewHealthHandler(db *database.DB, redis *cache.Redis, nats *messaging.NATS, cfg *config.Config) *HealthHandler {
	return &HealthHandler{db: db, redis: redis, nats: nats, cfg: cfg}
}

// HealthCheck godoc
// @Summary      Health check
// @Description  Returns basic health status
// @Tags         Health
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.Response
// @Router       /health [post]
func (h *HealthHandler) HealthCheck(ctx context.Context, c *app.RequestContext) {
	response.Success(c, map[string]string{"status": "ok"})
}

// DetailedHealthCheck godoc
// @Summary      Detailed health check
// @Description  Returns detailed health status with all dependencies
// @Tags         Health
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.Response{data=DetailedHealthStatus}
// @Router       /health/detailed [post]
func (h *HealthHandler) DetailedHealthCheck(ctx context.Context, c *app.RequestContext) {
	status := DetailedHealthStatus{
		Status:   "ok",
		Services: make(map[string]ServiceHealth),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Check PostgreSQL
	dbStart := time.Now()
	if err := h.db.Ping(ctx); err != nil {
		status.Services["postgresql"] = ServiceHealth{
			Status:  "error",
			Message: err.Error(),
			Latency: time.Since(dbStart).Milliseconds(),
		}
		status.Status = "degraded"
	} else {
		status.Services["postgresql"] = ServiceHealth{
			Status:  "ok",
			Latency: time.Since(dbStart).Milliseconds(),
		}
	}

	// Check Redis
	redisStart := time.Now()
	if err := h.redis.Ping(ctx); err != nil {
		status.Services["redis"] = ServiceHealth{
			Status:  "error",
			Message: err.Error(),
			Latency: time.Since(redisStart).Milliseconds(),
		}
		status.Status = "degraded"
	} else {
		status.Services["redis"] = ServiceHealth{
			Status:  "ok",
			Latency: time.Since(redisStart).Milliseconds(),
		}
	}

	// Check NATS
	natsStart := time.Now()
	if h.nats != nil {
		if err := h.nats.Ping(); err != nil {
			status.Services["nats"] = ServiceHealth{
				Status:  "error",
				Message: err.Error(),
				Latency: time.Since(natsStart).Milliseconds(),
			}
			status.Status = "degraded"
		} else {
			status.Services["nats"] = ServiceHealth{
				Status:  "ok",
				Latency: time.Since(natsStart).Milliseconds(),
			}
		}
	} else {
		status.Services["nats"] = ServiceHealth{
			Status:  "not_configured",
			Message: "NATS not configured",
		}
	}

	// Add system metrics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	status.System = SystemHealth{
		GoVersion:    runtime.Version(),
		Goroutines:   runtime.NumGoroutine(),
		MemoryAlloc:  memStats.Alloc,
		MemoryTotal:  memStats.TotalAlloc,
		MemorySys:    memStats.Sys,
		Uptime:       time.Since(time.Now()).Seconds(), // This would need a start time
	}

	h.cfg.Info("health check", zap.String("status", status.Status))
	response.Success(c, status)
}

// ReadyCheck godoc
// @Summary      Readiness check
// @Description  Returns readiness status for Kubernetes
// @Tags         Health
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.Response
// @Failure      503  {object}  response.Response
// @Router       /health/ready [post]
func (h *HealthHandler) ReadyCheck(ctx context.Context, c *app.RequestContext) {
	ready := true
	checks := make(map[string]bool)

	if err := h.db.Ping(ctx); err != nil {
		checks["postgresql"] = false
		ready = false
	} else {
		checks["postgresql"] = true
	}

	if err := h.redis.Ping(ctx); err != nil {
		checks["redis"] = false
		ready = false
	} else {
		checks["redis"] = true
	}

	if !ready {
		response.Error(c, consts.StatusServiceUnavailable, "service not ready")
		return
	}

	response.Success(c, map[string]interface{}{
		"status": "ready",
		"checks": checks,
	})
}

// LiveCheck godoc
// @Summary      Liveness check
// @Description  Returns liveness status for Kubernetes
// @Tags         Health
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.Response
// @Router       /health/live [post]
func (h *HealthHandler) LiveCheck(ctx context.Context, c *app.RequestContext) {
	response.Success(c, map[string]string{"status": "alive"})
}

// Version godoc
// @Summary      Version info
// @Description  Returns API version information
// @Tags         Health
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.Response{data=VersionInfo}
// @Router       /version [post]
func (h *HealthHandler) Version(ctx context.Context, c *app.RequestContext) {
	response.Success(c, VersionInfo{
		Version:   "1.0.0",
		GoVersion: runtime.Version(),
		BuildTime: "2024-01-01",
	})
}

// Metrics godoc
// @Summary      Prometheus metrics
// @Description  Returns Prometheus-style metrics
// @Tags         Health
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.Response{data=map[string]interface{}}
// @Router       /metrics [post]
func (h *HealthHandler) Metrics(ctx context.Context, c *app.RequestContext) {
	metrics := middleware.GetPrometheusMetrics()
	response.Success(c, metrics.ToJSON())
}

// ==================== Types ====================

type DetailedHealthStatus struct {
	Status    string                    `json:"status"`
	Services  map[string]ServiceHealth  `json:"services"`
	System    SystemHealth              `json:"system"`
	Timestamp string                    `json:"timestamp"`
}

type ServiceHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency int64  `json:"latency_ms"`
}

type SystemHealth struct {
	GoVersion  string  `json:"go_version"`
	Goroutines int     `json:"goroutines"`
	MemoryAlloc uint64 `json:"memory_alloc_bytes"`
	MemoryTotal uint64 `json:"memory_total_bytes"`
	MemorySys   uint64 `json:"memory_sys_bytes"`
	Uptime     float64 `json:"uptime_seconds"`
}

type VersionInfo struct {
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
	BuildTime string `json:"build_time"`
}
