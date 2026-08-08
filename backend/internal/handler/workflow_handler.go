package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
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

// GetWorkflows lists workflows
// @Summary      List workflows
// @Description  Get all workflows with pagination
// @Tags         Workflows
// @Accept       json
// @Produce      json
// @Param        request body     validation.ListWorkflowsRequest true  "List parameters"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/workflows [post]
func (h *WorkflowHandler) GetWorkflows(ctx context.Context, c *app.RequestContext) {
	var req validation.ListWorkflowsRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	workflows, err := h.svc.GetAllWorkflows(ctx, req.Limit, 0)
	if err != nil {
		h.cfg.Error("failed to get workflows", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get workflows")
		return
	}
	response.Success(c, workflows)
}

// buildWorkflowSteps converts validated request steps into domain steps.
func buildWorkflowSteps(input []validation.WorkflowStepInput) []domain.WorkflowStep {
	steps := make([]domain.WorkflowStep, 0, len(input))
	for _, s := range input {
		var approverID *uuid.UUID
		if s.ApproverID != "" {
			if id, err := uuid.Parse(s.ApproverID); err == nil {
				approverID = &id
			}
		}
		steps = append(steps, domain.WorkflowStep{
			Name:         s.Name,
			StepOrder:    s.StepOrder,
			ApproverRole: s.ApproverRole,
			ApproverID:   approverID,
			Action:       s.Action,
			TimeoutHours: s.TimeoutHours,
			IsRequired:   s.IsRequired,
		})
	}
	return steps
}

// GetWorkflow gets a single workflow by ID
// @Summary      Get workflow
// @Description  Get workflow details by ID
// @Tags         Workflows
// @Accept       json
// @Produce      json
// @Param        request body     validation.GetWorkflowRequest true  "Workflow ID"
// @Success      200  {object}  response.Response{data=domain.Workflow}
// @Failure      400  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /api/v1/workflows/get [post]
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

// CreateWorkflow creates a new workflow
// @Summary      Create workflow
// @Description  Create a new workflow record
// @Tags         Workflows
// @Accept       json
// @Produce      json
// @Param        request body     validation.CreateWorkflowRequest true  "Workflow details"
// @Success      201  {object}  response.Response{data=domain.Workflow}
// @Failure      400  {object}  response.Response
// @Failure      409  {object}  response.Response
// @Router       /api/v1/workflows/create [post]
func (h *WorkflowHandler) CreateWorkflow(ctx context.Context, c *app.RequestContext) {
	var req validation.CreateWorkflowRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	var description *string
	if req.Description != "" {
		description = &req.Description
	}
	wf := &domain.Workflow{Name: req.Name, Description: description, Category: req.Category, IsActive: req.IsActive}

	steps := buildWorkflowSteps(req.Steps)
	if err := h.svc.CreateWorkflowWithSteps(ctx, wf, steps); err != nil {
		h.cfg.Error("failed to create workflow", zap.Error(err))
		response.Error(c, http.StatusConflict, err.Error())
		return
	}

	// Mirror the steps into the JSON column so list/get responses expose them
	// without an extra join.
	if len(steps) > 0 {
		if raw, err := json.Marshal(steps); err == nil {
			wf.Steps = raw
		}
	}

	response.Success(c, wf)
}

// UpdateWorkflow updates a workflow
// @Summary      Update workflow
// @Description  Update workflow details
// @Tags         Workflows
// @Accept       json
// @Produce      json
// @Param        request body     validation.UpdateWorkflowRequest true  "Update details"
// @Success      200  {object}  response.Response{data=domain.Workflow}
// @Failure      400  {object}  response.Response
// @Router       /api/v1/workflows/update [post]
func (h *WorkflowHandler) UpdateWorkflow(ctx context.Context, c *app.RequestContext) {
	var req validation.UpdateWorkflowRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	wf, err := h.svc.GetWorkflow(ctx, req.WorkflowID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "workflow not found")
		return
	}

	if req.Name != "" {
		wf.Name = req.Name
	}
	if req.Description != "" {
		wf.Description = &req.Description
	}
	if req.Category != "" {
		wf.Category = req.Category
	}
	if req.IsActive != nil {
		wf.IsActive = *req.IsActive
	}

	steps := buildWorkflowSteps(req.Steps)
	if err := h.svc.UpdateWorkflow(ctx, wf, steps); err != nil {
		h.cfg.Error("failed to update workflow", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to update workflow")
		return
	}
	if len(steps) > 0 {
		if raw, err := json.Marshal(steps); err == nil {
			wf.Steps = raw
		}
	}

	response.Success(c, wf)
}

// DeleteWorkflow deletes a workflow
// @Summary      Delete workflow
// @Description  Delete a workflow record
// @Tags         Workflows
// @Accept       json
// @Produce      json
// @Param        request body     validation.DeleteWorkflowRequest true  "Workflow ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /api/v1/workflows/delete [post]
func (h *WorkflowHandler) DeleteWorkflow(ctx context.Context, c *app.RequestContext) {
	var req validation.DeleteWorkflowRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.DeleteWorkflow(ctx, req.WorkflowID); err != nil {
		h.cfg.Error("failed to delete workflow", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to delete workflow")
		return
	}
	response.Success(c, map[string]string{"message": "workflow deleted"})
}
