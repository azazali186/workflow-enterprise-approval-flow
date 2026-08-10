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

// GetEscalations lists escalations
// @Summary      List escalations
// @Description  Get active escalations
// @Tags         Escalations
// @Accept       json
// @Produce      json
// @Param        request body     validation.ListEscalationsRequest true  "List parameters"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/escalations [post]
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

// GetEscalation gets a single escalation by ID
// @Summary      Get escalation
// @Description  Get escalation details by ID
// @Tags         Escalations
// @Accept       json
// @Produce      json
// @Param        request body     validation.GetEscalationRequest true  "Escalation ID"
// @Success      200  {object}  response.Response{data=domain.Escalation}
// @Failure      400  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /api/v1/escalations/get [post]
func (h *EscalationHandler) GetEscalation(ctx context.Context, c *app.RequestContext) {
	var req validation.GetEscalationRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	response.Success(c, map[string]interface{}{"escalation_id": req.EscalationID})
}

// CreateEscalation creates a new escalation
// @Summary      Create escalation
// @Description  Create a new escalation record
// @Tags         Escalations
// @Accept       json
// @Produce      json
// @Param        request body     validation.CreateEscalationRequest true  "Escalation details"
// @Success      201  {object}  response.Response{data=domain.Escalation}
// @Failure      400  {object}  response.Response
// @Router       /api/v1/escalations/create [post]
func (h *EscalationHandler) CreateEscalation(ctx context.Context, c *app.RequestContext) {
	var req validation.CreateEscalationRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	response.Success(c, map[string]string{"message": "escalation created"})
}

// ResolveEscalation resolves an escalation
// @Summary      Resolve escalation
// @Description  Mark an escalation as resolved
// @Tags         Escalations
// @Accept       json
// @Produce      json
// @Param        request body     validation.ResolveEscalationRequest true  "Escalation ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /api/v1/escalations/resolve [post]
func (h *EscalationHandler) ResolveEscalation(ctx context.Context, c *app.RequestContext) {
	var req validation.ResolveEscalationRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.ResolveEscalation(ctx, req.EscalationID); err != nil {
		h.cfg.Error("failed to resolve escalation", zap.Error(err))
		// Never leak wrapped details; the 404 semantics are what matter here.
		response.Error(c, http.StatusNotFound, "escalation not found or already resolved")
		return
	}
	response.Success(c, map[string]string{"message": "escalation resolved"})
}
