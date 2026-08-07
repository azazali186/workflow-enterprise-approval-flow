package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/modules/escalation"
	"github.com/aeroxe/approval-flow/internal/pkg/response"
	"github.com/aeroxe/approval-flow/internal/pkg/validation"
)

type EscalationHandler struct {
	svc *escalation.Service
	cfg *config.Config
}

func NewEscalationHandler(svc *escalation.Service, cfg *config.Config) *EscalationHandler {
	return &EscalationHandler{svc: svc, cfg: cfg}
}

func (h *EscalationHandler) GetEscalations(ctx context.Context, c *app.RequestContext) {
	var req validation.ListEscalationsRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	escalations, err := h.svc.GetActiveEscalations(ctx)
	if err != nil {
		h.cfg.Error("failed to get escalations", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get escalations")
		return
	}
	response.Success(c, escalations)
}

func (h *EscalationHandler) GetEscalation(ctx context.Context, c *app.RequestContext) {
	var req validation.GetEscalationRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	response.Success(c, map[string]interface{}{"escalation_id": req.EscalationID})
}

func (h *EscalationHandler) CreateEscalation(ctx context.Context, c *app.RequestContext) {
	var req validation.CreateEscalationRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	response.Success(c, map[string]string{"message": "escalation created"})
}

func (h *EscalationHandler) ResolveEscalation(ctx context.Context, c *app.RequestContext) {
	var req validation.ResolveEscalationRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.ResolveEscalation(ctx, req.EscalationID); err != nil {
		h.cfg.Error("failed to resolve escalation", zap.Error(err))
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	response.Success(c, map[string]string{"message": "escalation resolved"})
}
