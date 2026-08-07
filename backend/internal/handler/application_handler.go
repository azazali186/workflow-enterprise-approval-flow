package handler

import (
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/modules/application"
)

type ApplicationHandler struct {
	service *application.Service
	logger  *config.Config
}

func NewApplicationHandler(service *application.Service, cfg *config.Config) *ApplicationHandler {
	return &ApplicationHandler{service: service, logger: cfg}
}

func (h *ApplicationHandler) SubmitApplication(ctx app.RequestContext) {
	var req struct {
		ApplicantID string                 `json:"applicant_id"`
		WorkflowID  string                 `json:"workflow_id"`
		TemplateID  string                 `json:"template_id"`
		Priority    string                 `json:"priority"`
		Data        map[string]interface{} `json:"data"`
	}

	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(consts.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	app := &domain.Application{
		ApplicantID: req.ApplicantID,
		WorkflowID:  req.WorkflowID,
		TemplateID:  req.TemplateID,
		Priority:    req.Priority,
		Status:      "submitted",
		Data:        req.Data,
	}

	if err := h.service.SubmitApplication(ctx, app); err != nil {
		h.logger.Error("failed to submit application", "error", err)
		ctx.JSON(consts.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	ctx.JSON(consts.StatusCreated, app)
}

func (h *ApplicationHandler) GetApplication(ctx app.RequestContext) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(consts.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}

	app, err := h.service.GetApplication(ctx, id)
	if err != nil {
		h.logger.Error("failed to get application", "error", err)
		ctx.JSON(consts.StatusNotFound, map[string]string{"error": "application not found"})
		return
	}

	ctx.JSON(consts.StatusOK, app)
}

func (h *ApplicationHandler) GetApplicationsByApplicant(ctx app.RequestContext) {
	applicantID := ctx.Param("applicant_id")
	if applicantID == "" {
		ctx.JSON(consts.StatusBadRequest, map[string]string{"error": "applicant_id is required"})
		return
	}

	limit, _ := strconv.Atoi(ctx.Query("limit"))
	if limit == 0 {
		limit = 10
	}
	offset, _ := strconv.Atoi(ctx.Query("offset"))

	apps, err := h.service.GetApplicationsByApplicant(ctx, applicantID, limit, offset)
	if err != nil {
		h.logger.Error("failed to get applications", "error", err)
		ctx.JSON(consts.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	ctx.JSON(consts.StatusOK, map[string]interface{}{
		"data": apps,
		"count": len(apps),
	})
}
