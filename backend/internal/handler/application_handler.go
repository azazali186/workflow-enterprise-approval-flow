package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/modules/application"
	"github.com/aeroxe/approval-flow/internal/pkg/pagination"
	"github.com/aeroxe/approval-flow/internal/pkg/response"
	"github.com/aeroxe/approval-flow/internal/pkg/validation"
)

type ApplicationHandler struct {
	svc *application.Service
	cfg *config.Config
}

func NewApplicationHandler(svc *application.Service, cfg *config.Config) *ApplicationHandler {
	return &ApplicationHandler{svc: svc, cfg: cfg}
}

// GetApplications godoc
// @Summary      List applications
// @Description  Get paginated list of applications
// @Tags         Applications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.ListApplicationsRequest true  "List parameters"
// @Success      200  {object}  response.Response
// @Router       /api/v1/applications [post]
func (h *ApplicationHandler) GetApplications(ctx context.Context, c *app.RequestContext) {
	var req validation.ListApplicationsRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	filters := &pagination.ApplicationFilters{}
	if req.Status != "" {
		filters.Status = &pagination.StatusFilter{Statuses: []string{req.Status}}
	}
	filters.Priority = req.Priority
	filters.ApplicantID = req.ApplicantID
	filters.WorkflowID = req.WorkflowID

	result, err := h.svc.ListApplications(ctx, filters, nil, req.Limit)
	if err != nil {
		h.cfg.Error("failed to get applications", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get applications")
		return
	}

	response.Success(c, result)
}

// GetApplication godoc
// @Summary      Get application
// @Description  Get application by ID
// @Tags         Applications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.GetApplicationRequest true  "Application ID"
// @Success      200  {object}  response.Response{data=domain.Application}
// @Failure      404  {object}  response.Response
// @Router       /api/v1/applications/get [post]
func (h *ApplicationHandler) GetApplication(ctx context.Context, c *app.RequestContext) {
	var req validation.GetApplicationRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	app, err := h.svc.GetApplication(ctx, req.ApplicationID)
	if err != nil {
		h.cfg.Error("failed to get application", zap.Error(err))
		response.Error(c, http.StatusNotFound, "application not found")
		return
	}

	response.Success(c, app)
}

// SubmitApplication godoc
// @Summary      Submit application
// @Description  Submit a new approval application
// @Tags         Applications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.SubmitApplicationRequest true  "Application details"
// @Success      201  {object}  response.Response{data=domain.Application}
// @Failure      400  {object}  response.Response
// @Router       /api/v1/applications/submit [post]
func (h *ApplicationHandler) SubmitApplication(ctx context.Context, c *app.RequestContext) {
	var req validation.SubmitApplicationRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	applicantID, err := uuid.Parse(req.ApplicantID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid applicant ID")
		return
	}
	workflowID, err := uuid.Parse(req.WorkflowID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid workflow ID")
		return
	}
	templateID, err := uuid.Parse(req.TemplateID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid template ID")
		return
	}

	now := time.Now()
	app := &domain.Application{
		ApplicantID: applicantID,
		WorkflowID:  workflowID,
		TemplateID:  templateID,
		Title:       req.Title,
		Status:      "submitted",
		Priority:    req.Priority,
		SubmittedAt: &now,
		Data:        domain.JSONMap(req.Data),
	}
	// The description is not a column yet; preserve it in the data payload.
	if req.Description != "" {
		if app.Data == nil {
			app.Data = domain.JSONMap{}
		}
		app.Data["description"] = req.Description
	}

	if err := h.svc.SubmitApplication(ctx, app); err != nil {
		h.cfg.Error("failed to submit application", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, app)
}

// UpdateApplication godoc
// @Summary      Update application
// @Description  Update an application
// @Tags         Applications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.UpdateApplicationRequest true  "Update details"
// @Success      200  {object}  response.Response{data=domain.Application}
// @Failure      400  {object}  response.Response
// @Router       /api/v1/applications/update [post]
func (h *ApplicationHandler) UpdateApplication(ctx context.Context, c *app.RequestContext) {
	var req validation.UpdateApplicationRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	app, err := h.svc.GetApplication(ctx, req.ApplicationID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "application not found")
		return
	}

	if req.Status != "" {
		app.Status = req.Status
	}
	if req.Priority != "" {
		app.Priority = req.Priority
	}
	if req.Data != nil {
		app.Data = domain.JSONMap(req.Data)
	}

	if err := h.svc.UpdateApplication(ctx, app); err != nil {
		h.cfg.Error("failed to update application", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to update application")
		return
	}

	response.Success(c, app)
}

// DeleteApplication godoc
// @Summary      Delete application
// @Description  Delete an application
// @Tags         Applications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.DeleteApplicationRequest true  "Application ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /api/v1/applications/delete [post]
func (h *ApplicationHandler) DeleteApplication(ctx context.Context, c *app.RequestContext) {
	var req validation.DeleteApplicationRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.svc.Repo.Delete(ctx, req.ApplicationID); err != nil {
		h.cfg.Error("failed to delete application", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to delete application")
		return
	}
	response.Success(c, map[string]string{"message": "application deleted"})
}
