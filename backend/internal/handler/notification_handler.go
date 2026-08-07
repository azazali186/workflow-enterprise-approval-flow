package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/modules/notification"
	"github.com/aeroxe/approval-flow/internal/pkg/response"
	"github.com/aeroxe/approval-flow/internal/pkg/validation"
)

type NotificationHandler struct {
	svc *notification.Service
	cfg *config.Config
}

func NewNotificationHandler(svc *notification.Service, cfg *config.Config) *NotificationHandler {
	return &NotificationHandler{svc: svc, cfg: cfg}
}

func (h *NotificationHandler) GetNotifications(ctx context.Context, c *app.RequestContext) {
	var req validation.GetNotificationsRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Limit <= 0 { req.Limit = 10 }
	notifications, err := h.svc.GetUserNotifications(ctx, req.UserID, req.Limit, 0)
	if err != nil {
		h.cfg.Error("failed to get notifications", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get notifications")
		return
	}
	response.Success(c, notifications)
}

func (h *NotificationHandler) GetUnreadNotifications(ctx context.Context, c *app.RequestContext) {
	var req validation.GetUnreadNotificationsRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	notifications, err := h.svc.GetUnreadNotifications(ctx, req.UserID)
	if err != nil {
		h.cfg.Error("failed to get unread notifications", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get unread notifications")
		return
	}
	response.Success(c, notifications)
}

func (h *NotificationHandler) SendNotification(ctx context.Context, c *app.RequestContext) {
	var req validation.SendNotificationRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	response.Success(c, map[string]string{"message": "notification sent"})
}

func (h *NotificationHandler) MarkAsRead(ctx context.Context, c *app.RequestContext) {
	var req validation.MarkNotificationReadRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.MarkAsRead(ctx, req.NotificationID); err != nil {
		h.cfg.Error("failed to mark notification as read", zap.Error(err))
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	response.Success(c, map[string]string{"message": "notification marked as read"})
}

func (h *NotificationHandler) GetNotificationStats(ctx context.Context, c *app.RequestContext) {
	var req validation.GetNotificationStatsRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	response.Success(c, map[string]interface{}{"user_id": req.UserID})
}
