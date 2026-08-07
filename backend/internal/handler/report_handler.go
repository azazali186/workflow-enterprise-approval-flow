package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/modules/report"
	"github.com/aeroxe/approval-flow/internal/pkg/response"
	"github.com/aeroxe/approval-flow/internal/pkg/validation"
)

type ReportHandler struct {
	svc *report.Service
	cfg *config.Config
}

func NewReportHandler(svc *report.Service, cfg *config.Config) *ReportHandler {
	return &ReportHandler{svc: svc, cfg: cfg}
}

func (h *ReportHandler) GetStatuses(ctx context.Context, c *app.RequestContext) {
	var req validation.GetStatusesRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	statuses, err := h.svc.GetStatusesByEntityType(ctx, req.EntityType)
	if err != nil {
		h.cfg.Error("failed to get statuses", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get statuses")
		return
	}
	response.Success(c, statuses)
}

func (h *ReportHandler) GetComments(ctx context.Context, c *app.RequestContext) {
	var req validation.GetCommentsRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	comments, err := h.svc.GetComments(ctx, req.ApprovalID)
	if err != nil {
		h.cfg.Error("failed to get comments", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get comments")
		return
	}
	response.Success(c, comments)
}

func (h *ReportHandler) GetDocuments(ctx context.Context, c *app.RequestContext) {
	var req validation.GetDocumentsRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	documents, err := h.svc.GetDocuments(ctx, req.ApplicationID)
	if err != nil {
		h.cfg.Error("failed to get documents", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get documents")
		return
	}
	response.Success(c, documents)
}
