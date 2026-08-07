package handler

import (
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/modules/workflow"
)

type WorkflowHandler struct {
	service *workflow.Service
	logger  *config.Config
}

func NewWorkflowHandler(service *workflow.Service, cfg *config.Config) *WorkflowHandler {
	return &WorkflowHandler{service: service, logger: cfg}
}

func (h *WorkflowHandler) CreateWorkflow(ctx app.RequestContext) {
	var req struct {
		Name        string                 `json:"name"`
		Description *string                `json:"description"`
		Category    string                 `json:"category"`
		Steps       []domain.WorkflowStep  `json:"steps"`
	}

	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(consts.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	workflow := &domain.Workflow{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Version:     1,
		IsActive:    true,
	}

	if err := h.service.CreateWorkflow(ctx, workflow); err != nil {
		h.logger.Error("failed to create workflow", "error", err)
		ctx.JSON(consts.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	for i, step := range req.Steps {
		step.WorkflowID = workflow.ID
		step.StepOrder = i + 1
		if err := h.service.AddWorkflowStep(ctx, &step); err != nil {
			h.logger.Error("failed to add workflow step", "error", err)
		}
	}

	ctx.JSON(consts.StatusCreated, workflow)
}

func (h *WorkflowHandler) GetWorkflow(ctx app.RequestContext) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(consts.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}

	workflow, err := h.service.GetWorkflow(ctx, id)
	if err != nil {
		h.logger.Error("failed to get workflow", "error", err)
		ctx.JSON(consts.StatusNotFound, map[string]string{"error": "workflow not found"})
		return
	}

	ctx.JSON(consts.StatusOK, workflow)
}

func (h *WorkflowHandler) GetWorkflows(ctx app.RequestContext) {
	category := ctx.Query("category")
	limit, _ := strconv.Atoi(ctx.Query("limit"))
	if limit == 0 {
		limit = 10
	}
	offset, _ := strconv.Atoi(ctx.Query("offset"))

	var workflows []domain.Workflow
	var err error

	if category != "" {
		workflows, err = h.service.GetWorkflowsByCategory(ctx, category)
	} else {
		workflows, err = h.service.GetAllWorkflows(ctx, limit, offset)
	}

	if err != nil {
		h.logger.Error("failed to get workflows", "error", err)
		ctx.JSON(consts.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	ctx.JSON(consts.StatusOK, map[string]interface{}{
		"data": workflows,
		"count": len(workflows),
	})
}
