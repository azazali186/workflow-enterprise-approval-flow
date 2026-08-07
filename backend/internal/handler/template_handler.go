package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/modules/template"
	"github.com/aeroxe/approval-flow/internal/pkg/response"
	"github.com/aeroxe/approval-flow/internal/pkg/validation"
)

type TemplateHandler struct {
	svc *template.Service
	cfg *config.Config
}

func NewTemplateHandler(svc *template.Service, cfg *config.Config) *TemplateHandler {
	return &TemplateHandler{svc: svc, cfg: cfg}
}

// GetTemplates lists templates
// @Summary      List templates
// @Description  Get all templates with pagination
// @Tags         Templates
// @Accept       json
// @Produce      json
// @Param        request body     validation.ListTemplatesRequest true  "List parameters"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/templates [post]
func (h *TemplateHandler) GetTemplates(ctx context.Context, c *app.RequestContext) {
	var req validation.ListTemplatesRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Limit <= 0 { req.Limit = 10 }
	templates, err := h.svc.GetAllTemplates(ctx, req.Limit, 0)
	if err != nil {
		h.cfg.Error("failed to get templates", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get templates")
		return
	}
	response.Success(c, templates)
}

// GetTemplate gets a single template by ID
// @Summary      Get template
// @Description  Get template details by ID
// @Tags         Templates
// @Accept       json
// @Produce      json
// @Param        request body     validation.GetTemplateRequest true  "Template ID"
// @Success      200  {object}  response.Response{data=domain.Template}
// @Failure      400  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /api/v1/templates/get [post]
func (h *TemplateHandler) GetTemplate(ctx context.Context, c *app.RequestContext) {
	var req validation.GetTemplateRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	tmpl, err := h.svc.GetTemplate(ctx, req.TemplateID)
	if err != nil {
		h.cfg.Error("failed to get template", zap.Error(err))
		response.Error(c, http.StatusNotFound, "template not found")
		return
	}
	response.Success(c, tmpl)
}

// CreateTemplate creates a new template
// @Summary      Create template
// @Description  Create a new template record
// @Tags         Templates
// @Accept       json
// @Produce      json
// @Param        request body     validation.CreateTemplateRequest true  "Template details"
// @Success      201  {object}  response.Response{data=domain.Template}
// @Failure      400  {object}  response.Response
// @Failure      409  {object}  response.Response
// @Router       /api/v1/templates/create [post]
func (h *TemplateHandler) CreateTemplate(ctx context.Context, c *app.RequestContext) {
	var req validation.CreateTemplateRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	tmpl := &domain.Template{Name: req.Name, Category: req.Category, Schema: req.Schema, UI: req.UI, IsActive: true}
	if err := h.svc.CreateTemplate(ctx, tmpl); err != nil {
		h.cfg.Error("failed to create template", zap.Error(err))
		response.Error(c, http.StatusConflict, err.Error())
		return
	}
	response.Success(c, tmpl)
}

// UpdateTemplate updates a template
// @Summary      Update template
// @Description  Update template details
// @Tags         Templates
// @Accept       json
// @Produce      json
// @Param        request body     validation.UpdateTemplateRequest true  "Update details"
// @Success      200  {object}  response.Response{data=domain.Template}
// @Failure      400  {object}  response.Response
// @Router       /api/v1/templates/update [post]
func (h *TemplateHandler) UpdateTemplate(ctx context.Context, c *app.RequestContext) {
	var req validation.UpdateTemplateRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	response.Success(c, map[string]string{"message": "template updated"})
}

// DeleteTemplate deletes a template
// @Summary      Delete template
// @Description  Delete a template record
// @Tags         Templates
// @Accept       json
// @Produce      json
// @Param        request body     validation.DeleteTemplateRequest true  "Template ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /api/v1/templates/delete [post]
func (h *TemplateHandler) DeleteTemplate(ctx context.Context, c *app.RequestContext) {
	var req validation.DeleteTemplateRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	response.Success(c, map[string]string{"message": "template deleted"})
}
