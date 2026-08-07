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

func (h *TemplateHandler) UpdateTemplate(ctx context.Context, c *app.RequestContext) {
	var req validation.UpdateTemplateRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	response.Success(c, map[string]string{"message": "template updated"})
}

func (h *TemplateHandler) DeleteTemplate(ctx context.Context, c *app.RequestContext) {
	var req validation.DeleteTemplateRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	response.Success(c, map[string]string{"message": "template deleted"})
}
