package handler

import (
	"net/http"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/modules/notification"
)

type NotificationHandler struct {
	service *notification.Service
	logger  *config.Config
}

func NewNotificationHandler(service *notification.Service, cfg *config.Config) *NotificationHandler {
	return &NotificationHandler{service: service, logger: cfg}
}

func (h *NotificationHandler) GetNotifications(ctx app.RequestContext) {
	userID := ctx.Param("user_id")
	if userID == "" {
		ctx.JSON(consts.StatusBadRequest, map[string]string{"error": "user_id is required"})
		return
	}

	limit, _ := strconv.Atoi(ctx.Query("limit"))
	if limit == 0 {
		limit = 10
	}
	offset, _ := strconv.Atoi(ctx.Query("offset"))

	notifications, err := h.service.GetUserNotifications(ctx, userID, limit, offset)
	if err != nil {
		h.logger.Error("failed to get notifications", "error", err)
		ctx.JSON(consts.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	ctx.JSON(consts.StatusOK, map[string]interface{}{
		"data": notifications,
		"count": len(notifications),
	})
}

func (h *NotificationHandler) GetUnreadNotifications(ctx app.RequestContext) {
	userID := ctx.Param("user_id")
	if userID == "" {
		ctx.JSON(consts.StatusBadRequest, map[string]string{"error": "user_id is required"})
		return
	}

	notifications, err := h.service.GetUnreadNotifications(ctx, userID)
	if err != nil {
		h.logger.Error("failed to get unread notifications", "error", err)
		ctx.JSON(consts.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	ctx.JSON(consts.StatusOK, map[string]interface{}{
		"data": notifications,
		"count": len(notifications),
	})
}

func (h *NotificationHandler) MarkAsRead(ctx app.RequestContext) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(consts.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}

	if err := h.service.MarkAsRead(ctx, id); err != nil {
		h.logger.Error("failed to mark notification as read", "error", err)
		ctx.JSON(consts.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	ctx.JSON(consts.StatusOK, map[string]string{"status": "ok"})
}

func (h *NotificationHandler) HealthCheck(ctx app.RequestContext) {
	ctx.JSON(consts.StatusOK, map[string]string{"status": "ok"})
}
