package handler

import (
	"net/http"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/modules/approval"
)

type ApprovalHandler struct {
	service *approval.Service
	logger  *config.Config
}

func NewApprovalHandler(service *approval.Service, cfg *config.Config) *ApprovalHandler {
	return &ApprovalHandler{service: service, logger: cfg}
}

func (h *ApprovalHandler) GetPendingApprovals(ctx app.RequestContext) {
	approverID := ctx.Param("approver_id")
	if approverID == "" {
		ctx.JSON(consts.StatusBadRequest, map[string]string{"error": "approver_id is required"})
		return
	}

	limit, _ := strconv.Atoi(ctx.Query("limit"))
	if limit == 0 {
		limit = 10
	}
	offset, _ := strconv.Atoi(ctx.Query("offset"))

	approvals, err := h.service.GetPendingApprovals(ctx, approverID)
	if err != nil {
		h.logger.Error("failed to get pending approvals", "error", err)
		ctx.JSON(consts.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	ctx.JSON(consts.StatusOK, map[string]interface{}{
		"data": approvals,
		"count": len(approvals),
	})
}

func (h *ApprovalHandler) DecideApproval(ctx app.RequestContext) {
	approvalID := ctx.Param("id")
	if approvalID == "" {
		ctx.JSON(consts.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}

	var req struct {
		Decision string `json:"decision"`
		Comment  string `json:"comment"`
	}

	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(consts.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Decision == "" {
		ctx.JSON(consts.StatusBadRequest, map[string]string{"error": "decision is required"})
		return
	}

	approval, err := h.service.DecideApproval(ctx, approvalID, req.Decision, req.Comment)
	if err != nil {
		h.logger.Error("failed to decide approval", "error", err)
		ctx.JSON(consts.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	ctx.JSON(consts.StatusOK, approval)
}

func (h *ApprovalHandler) GetApproval(ctx app.RequestContext) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(consts.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}

	approval, err := h.service.GetApproval(ctx, id)
	if err != nil {
		h.logger.Error("failed to get approval", "error", err)
		ctx.JSON(consts.StatusNotFound, map[string]string{"error": "approval not found"})
		return
	}

	ctx.JSON(consts.StatusOK, approval)
}
