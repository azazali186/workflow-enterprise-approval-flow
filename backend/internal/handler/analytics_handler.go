package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/modules/analytics"
	"github.com/aeroxe/approval-flow/internal/pkg/response"
	"github.com/aeroxe/approval-flow/internal/pkg/validation"
)

type AnalyticsHandler struct {
	svc *analytics.Service
	cfg *config.Config
}

func NewAnalyticsHandler(svc *analytics.Service, cfg *config.Config) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc, cfg: cfg}
}

// GetApprovalStats gets approval statistics
// @Summary      Get approval stats
// @Description  Get approval statistics with date range
// @Tags         Analytics
// @Accept       json
// @Produce      json
// @Param        request body     validation.GetApprovalStatsRequest true  "Date range"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/analytics/approval-stats [post]
func (h *AnalyticsHandler) GetApprovalStats(ctx context.Context, c *app.RequestContext) {
	var req validation.GetApprovalStatsRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil { startDate = t }
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil { endDate = t }
	}
	stats, err := h.svc.GetApprovalStats(ctx, startDate, endDate)
	if err != nil {
		h.cfg.Error("failed to get approval stats", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get stats")
		return
	}
	response.Success(c, stats)
}

// GetWorkflowPerformance gets workflow performance metrics
// @Summary      Get workflow performance
// @Description  Get performance metrics for a specific workflow
// @Tags         Analytics
// @Accept       json
// @Produce      json
// @Param        request body     validation.GetWorkflowPerformanceRequest true  "Workflow ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /api/v1/analytics/workflow-performance [post]
func (h *AnalyticsHandler) GetWorkflowPerformance(ctx context.Context, c *app.RequestContext) {
	var req validation.GetWorkflowPerformanceRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	perf, err := h.svc.GetWorkflowPerformance(ctx, req.WorkflowID)
	if err != nil {
		h.cfg.Error("failed to get workflow performance", zap.Error(err))
		response.Error(c, http.StatusNotFound, "workflow not found")
		return
	}
	response.Success(c, perf)
}

// GetApproverPerformance gets approver performance metrics
// @Summary      Get approver performance
// @Description  Get performance metrics for a specific approver
// @Tags         Analytics
// @Accept       json
// @Produce      json
// @Param        request body     validation.GetApproverPerformanceRequest true  "Approver ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /api/v1/analytics/approver-performance [post]
func (h *AnalyticsHandler) GetApproverPerformance(ctx context.Context, c *app.RequestContext) {
	var req validation.GetApproverPerformanceRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	perf, err := h.svc.GetApproverPerformance(ctx, req.ApproverID)
	if err != nil {
		h.cfg.Error("failed to get approver performance", zap.Error(err))
		response.Error(c, http.StatusNotFound, "approver not found")
		return
	}
	response.Success(c, perf)
}

// GetEscalationMetrics gets escalation metrics
// @Summary      Get escalation metrics
// @Description  Get escalation metrics with date range
// @Tags         Analytics
// @Accept       json
// @Produce      json
// @Param        request body     validation.GetEscalationMetricsRequest true  "Date range"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/analytics/escalation-metrics [post]
func (h *AnalyticsHandler) GetEscalationMetrics(ctx context.Context, c *app.RequestContext) {
	var req validation.GetEscalationMetricsRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil { startDate = t }
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil { endDate = t }
	}
	metrics, err := h.svc.GetEscalationMetrics(ctx, startDate, endDate)
	if err != nil {
		h.cfg.Error("failed to get escalation metrics", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get metrics")
		return
	}
	response.Success(c, metrics)
}
