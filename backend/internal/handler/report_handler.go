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

// GetStatuses gets statuses by entity type
// @Summary      Get statuses
// @Description  Get statuses filtered by entity type
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Param        request body     validation.GetStatusesRequest true  "Entity type"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/reports/statuses [post]
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

// GetComments gets comments for an approval
// @Summary      Get comments
// @Description  Get comments for a specific approval
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Param        request body     validation.GetCommentsRequest true  "Approval ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/reports/comments [post]
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

// GetDocuments gets documents for an application
// @Summary      Get documents
// @Description  Get documents for a specific application
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Param        request body     validation.GetDocumentsRequest true  "Application ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/reports/documents [post]
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
