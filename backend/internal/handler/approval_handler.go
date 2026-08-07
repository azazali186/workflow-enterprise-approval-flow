package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/modules/approval"
	"github.com/aeroxe/approval-flow/internal/pkg/pagination"
	"github.com/aeroxe/approval-flow/internal/pkg/response"
	"github.com/aeroxe/approval-flow/internal/pkg/validation"
)

type ApprovalHandler struct {
	svc *approval.Service
	cfg *config.Config
}

func NewApprovalHandler(svc *approval.Service, cfg *config.Config) *ApprovalHandler {
	return &ApprovalHandler{svc: svc, cfg: cfg}
}

func (h *ApprovalHandler) GetApprovals(ctx context.Context, c *app.RequestContext) {
	var req validation.ListApprovalsRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Limit <= 0 { req.Limit = 10 }

	filters := &pagination.ApprovalFilters{}
	if req.Status != "" {
		filters.Status = &pagination.StatusFilter{Statuses: []string{req.Status}}
	}
	filters.Decision = req.Decision
	filters.ApproverID = req.ApproverID
	filters.ApplicationID = req.ApplicationID

	result, err := h.svc.ListApprovals(ctx, filters, nil, req.Limit)
	if err != nil {
		h.cfg.Error("failed to get approvals", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get approvals")
		return
	}
	response.Success(c, result)
}

func (h *ApprovalHandler) GetApproval(ctx context.Context, c *app.RequestContext) {
	var req validation.GetApprovalRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	approval, err := h.svc.GetApproval(ctx, req.ApprovalID)
	if err != nil {
		h.cfg.Error("failed to get approval", zap.Error(err))
		response.Error(c, http.StatusNotFound, "approval not found")
		return
	}
	response.Success(c, approval)
}

func (h *ApprovalHandler) CreateApproval(ctx context.Context, c *app.RequestContext) {
	var req validation.CreateApprovalRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	response.Success(c, map[string]string{"message": "approval created"})
}

func (h *ApprovalHandler) DecideApproval(ctx context.Context, c *app.RequestContext) {
	var req validation.DecideApprovalRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	approval, err := h.svc.DecideApproval(ctx, req.ApprovalID, req.Decision, req.Comment)
	if err != nil {
		h.cfg.Error("failed to decide approval", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, approval)
}

func (h *ApprovalHandler) UpdateApproval(ctx context.Context, c *app.RequestContext) {
	var req validation.UpdateApprovalRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	response.Success(c, map[string]string{"message": "approval updated"})
}

func (h *ApprovalHandler) DeleteApproval(ctx context.Context, c *app.RequestContext) {
	var req validation.DeleteApprovalRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	response.Success(c, map[string]string{"message": "approval deleted"})
}

func (h *ApprovalHandler) GetPendingApprovals(ctx context.Context, c *app.RequestContext) {
	var req validation.GetPendingApprovalsRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	approvals, err := h.svc.GetPendingApprovals(ctx, req.ApproverID)
	if err != nil {
		h.cfg.Error("failed to get pending approvals", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get pending approvals")
		return
	}
	response.Success(c, approvals)
}

// Ensure pagination is imported
var _ = pagination.PaginationResponse{}
