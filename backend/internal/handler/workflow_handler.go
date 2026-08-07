package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/modules/workflow"
	"github.com/aeroxe/approval-flow/internal/pkg/response"
	"github.com/aeroxe/approval-flow/internal/pkg/validation"
)

type WorkflowHandler struct {
	svc *workflow.Service
	cfg *config.Config
}

func NewWorkflowHandler(svc *workflow.Service, cfg *config.Config) *WorkflowHandler {
	return &WorkflowHandler{svc: svc, cfg: cfg}
}

func (h *WorkflowHandler) GetWorkflows(ctx context.Context, c *app.RequestContext) {
	var req validation.ListWorkflowsRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Limit <= 0 { req.Limit = 10 }
	workflows, err := h.svc.GetAllWorkflows(ctx, req.Limit, 0)
	if err != nil {
		h.cfg.Error("failed to get workflows", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get workflows")
		return
	}
	response.Success(c, workflows)
}

func (h *WorkflowHandler) GetWorkflow(ctx context.Context, c *app.RequestContext) {
	var req validation.GetWorkflowRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	wf, err := h.svc.GetWorkflow(ctx, req.WorkflowID)
	if err != nil {
		h.cfg.Error("failed to get workflow", zap.Error(err))
		response.Error(c, http.StatusNotFound, "workflow not found")
		return
	}
	response.Success(c, wf)
}

func (h *WorkflowHandler) CreateWorkflow(ctx context.Context, c *app.RequestContext) {
	var req validation.CreateWorkflowRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	var description *string
	if req.Description != "" { description = &req.Description }
	wf := &domain.Workflow{Name: req.Name, Description: description, Category: req.Category, IsActive: req.IsActive}
	if err := h.svc.CreateWorkflow(ctx, wf); err != nil {
		h.cfg.Error("failed to create workflow", zap.Error(err))
		response.Error(c, http.StatusConflict, err.Error())
		return
	}
	response.Success(c, wf)
}

func (h *WorkflowHandler) UpdateWorkflow(ctx context.Context, c *app.RequestContext) {
	var req validation.UpdateWorkflowRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	response.Success(c, map[string]string{"message": "workflow updated"})
}

func (h *WorkflowHandler) DeleteWorkflow(ctx context.Context, c *app.RequestContext) {
	var req validation.DeleteWorkflowRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	response.Success(c, map[string]string{"message": "workflow deleted"})
}
