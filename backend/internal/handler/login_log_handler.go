package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/modules/login_log"
	"github.com/aeroxe/approval-flow/internal/pkg/response"
	"github.com/aeroxe/approval-flow/internal/pkg/validation"
)

type LoginLogHandler struct {
	svc *login_log.Service
	cfg *config.Config
}

func NewLoginLogHandler(svc *login_log.Service, cfg *config.Config) *LoginLogHandler {
	return &LoginLogHandler{svc: svc, cfg: cfg}
}

// GetLoginHistory godoc
// @Summary      Get login history
// @Description  Get login history for a user
// @Tags         Auth - Login Logs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body validation.GetLoginHistoryRequest true "Login history request"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /api/v1/auth/login-history [post]
func (h *LoginLogHandler) GetLoginHistory(ctx context.Context, c *app.RequestContext) {
	var req validation.GetLoginHistoryRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	logs, err := h.svc.GetLoginHistory(ctx, req.UserID, req.Limit)
	if err != nil {
		h.cfg.Error("failed to get login history", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get login history")
		return
	}

	response.Success(c, logs)
}

// GetLoginHistoryByEmail godoc
// @Summary      Get login history by email
// @Description  Get login history by email address
// @Tags         Auth - Login Logs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body validation.GetLoginHistoryByEmailRequest true "Login history request"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /api/v1/auth/login-history/email [post]
func (h *LoginLogHandler) GetLoginHistoryByEmail(ctx context.Context, c *app.RequestContext) {
	var req validation.GetLoginHistoryByEmailRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	logs, err := h.svc.GetLoginHistoryByEmail(ctx, req.Email, req.Limit)
	if err != nil {
		h.cfg.Error("failed to get login history", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get login history")
		return
	}

	response.Success(c, logs)
}

// GetLoginStats godoc
// @Summary      Get login stats
// @Description  Get login statistics for a user
// @Tags         Auth - Login Logs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body validation.GetLoginStatsRequest true "Login stats request"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /api/v1/auth/login-stats [post]
func (h *LoginLogHandler) GetLoginStats(ctx context.Context, c *app.RequestContext) {
	var req validation.GetLoginStatsRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	stats, err := h.svc.GetLoginStats(ctx, req.UserID)
	if err != nil {
		h.cfg.Error("failed to get login stats", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get login stats")
		return
	}

	response.Success(c, stats)
}

// GetLoginLogs godoc
// @Summary      Get login logs (admin)
// @Description  Get all login logs with filters (admin only)
// @Tags         Auth - Login Logs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body validation.GetLoginLogsRequest true "Login logs request"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /api/v1/admin/login-logs [post]
func (h *LoginLogHandler) GetLoginLogs(ctx context.Context, c *app.RequestContext) {
	var req validation.GetLoginLogsRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	offset := 0
	if req.Page > 1 {
		offset = (req.Page - 1) * req.Limit
	}

	filters := map[string]interface{}{
		"email":      req.Email,
		"status":     req.Status,
		"user_id":    req.UserID,
		"start_date": req.StartDate,
		"end_date":   req.EndDate,
	}

	logs, totalCount, err := h.svc.GetLoginLogs(ctx, req.Limit, offset, filters)
	if err != nil {
		h.cfg.Error("failed to get login logs", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get login logs")
		return
	}

	totalPages := int(totalCount) / req.Limit
	if int(totalCount)%req.Limit > 0 {
		totalPages++
	}

	response.Paginated(c, logs, map[string]interface{}{
		"total_count": totalCount,
		"page":        req.Page,
		"page_size":   req.Limit,
		"total_pages": totalPages,
		"has_more":    req.Page < totalPages,
	})
}
