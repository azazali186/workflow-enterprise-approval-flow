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

// GetNotifications gets notifications for a user
// @Summary      List notifications
// @Description  Get notifications for a specific user
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Param        request body     validation.GetNotificationsRequest true  "User ID and limit"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/notifications [post]
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

// GetUnreadNotifications gets unread notifications for a user
// @Summary      Get unread notifications
// @Description  Get unread notifications for a specific user
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Param        request body     validation.GetUnreadNotificationsRequest true  "User ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/notifications/unread [post]
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

// SendNotification sends a notification
// @Summary      Send notification
// @Description  Send a notification to a user
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Param        request body     validation.SendNotificationRequest true  "Notification details"
// @Success      201  {object}  response.Response{data=domain.Notification}
// @Failure      400  {object}  response.Response
// @Router       /api/v1/notifications/send [post]
func (h *NotificationHandler) SendNotification(ctx context.Context, c *app.RequestContext) {
	var req validation.SendNotificationRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	response.Success(c, map[string]string{"message": "notification sent"})
}

// MarkAsRead marks a notification as read
// @Summary      Mark notification as read
// @Description  Mark a notification as read
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Param        request body     validation.MarkNotificationReadRequest true  "Notification ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /api/v1/notifications/read [post]
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

// GetNotificationStats gets notification statistics
// @Summary      Get notification stats
// @Description  Get notification statistics for a user
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Param        request body     validation.GetNotificationStatsRequest true  "User ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /api/v1/notifications/stats [post]
func (h *NotificationHandler) GetNotificationStats(ctx context.Context, c *app.RequestContext) {
	var req validation.GetNotificationStatsRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	response.Success(c, map[string]interface{}{"user_id": req.UserID})
}
