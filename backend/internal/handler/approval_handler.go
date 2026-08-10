package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/modules/approval"
	"github.com/aeroxe/approval-flow/internal/pkg/middleware"
	"github.com/aeroxe/approval-flow/internal/pkg/pagination"
	"github.com/aeroxe/approval-flow/internal/pkg/response"
	"github.com/aeroxe/approval-flow/internal/pkg/validation"
)

type ApprovalHandler struct {
	svc *approval.Service
	cfg *config.Config
}

func NewApprovalHandler(svc *approval.Service, cfg *config.Config) *ApprovalHandler {
	return &ApprovalHandler{svc: svc, cfg: cfg}
}

// GetApprovals lists approvals with filters
// @Summary      List approvals
// @Description  Get approvals with filters and pagination
// @Tags         Approvals
// @Accept       json
// @Produce      json
// @Param        request body     validation.ListApprovalsRequest true  "List parameters"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/approvals [post]
func (h *ApprovalHandler) GetApprovals(ctx context.Context, c *app.RequestContext) {
	var req validation.ListApprovalsRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	filters := &pagination.ApprovalFilters{}
	if req.Status != "" {
		filters.Status = &pagination.StatusFilter{Statuses: []string{req.Status}}
	}
	filters.Decision = req.Decision
	filters.ApproverID = req.ApproverID
	filters.ApplicationID = req.ApplicationID

	result, err := h.svc.ListApprovals(ctx, filters, nil, req.Limit)
	if err != nil {
		h.cfg.Error("failed to get approvals", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get approvals")
		return
	}
	response.Success(c, result)
}

// GetApproval gets a single approval by ID
// @Summary      Get approval
// @Description  Get approval details by ID
// @Tags         Approvals
// @Accept       json
// @Produce      json
// @Param        request body     validation.GetApprovalRequest true  "Approval ID"
// @Success      200  {object}  response.Response{data=domain.Approval}
// @Failure      400  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /api/v1/approvals/get [post]
func (h *ApprovalHandler) GetApproval(ctx context.Context, c *app.RequestContext) {
	var req validation.GetApprovalRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	approval, err := h.svc.GetApproval(ctx, req.ApprovalID)
	if err != nil {
		h.cfg.Error("failed to get approval", zap.Error(err))
		response.Error(c, http.StatusNotFound, "approval not found")
		return
	}
	response.Success(c, approval)
}

// CreateApproval creates a new approval
// @Summary      Create approval
// @Description  Create a new approval record
// @Tags         Approvals
// @Accept       json
// @Produce      json
// @Param        request body     validation.CreateApprovalRequest true  "Approval details"
// @Success      201  {object}  response.Response{data=domain.Approval}
// @Failure      400  {object}  response.Response
// @Router       /api/v1/approvals/create [post]
func (h *ApprovalHandler) CreateApproval(ctx context.Context, c *app.RequestContext) {
	var req validation.CreateApprovalRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	applicationID, err := uuid.Parse(req.ApplicationID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid application ID")
		return
	}
	stepID, err := uuid.Parse(req.WorkflowStepID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid workflow step ID")
		return
	}
	approverID, err := uuid.Parse(req.ApproverID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid approver ID")
		return
	}

	approval := &domain.Approval{
		ApplicationID:  applicationID,
		WorkflowStepID: stepID,
		ApproverID:     approverID,
		Status:         "pending",
	}
	if err := h.svc.CreateApproval(ctx, approval); err != nil {
		h.cfg.Error("failed to create approval", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to create approval")
		return
	}
	response.Success(c, approval)
}

// DecideApproval decides on an approval
// @Summary      Decide approval
// @Description  Approve or reject an approval
// @Tags         Approvals
// @Accept       json
// @Produce      json
// @Param        request body     validation.DecideApprovalRequest true  "Decision details"
// @Success      200  {object}  response.Response{data=domain.Approval}
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/approvals/decide [post]
func (h *ApprovalHandler) DecideApproval(ctx context.Context, c *app.RequestContext) {
	var req validation.DecideApprovalRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	approval, err := h.svc.DecideApproval(ctx, req.ApprovalID, req.Decision, req.Comment)
	if err != nil {
		h.cfg.Error("failed to decide approval", zap.Error(err))
		// Internal error: never leak wrapped details (e.g. DB drivers) to the client.
		response.Error(c, http.StatusInternalServerError, "failed to process approval decision")
		return
	}
	response.Success(c, approval)
}

// UpdateApproval updates an approval
// @Summary      Update approval
// @Description  Update approval details
// @Tags         Approvals
// @Accept       json
// @Produce      json
// @Param        request body     validation.UpdateApprovalRequest true  "Update details"
// @Success      200  {object}  response.Response{data=domain.Approval}
// @Failure      400  {object}  response.Response
// @Router       /api/v1/approvals/update [post]
func (h *ApprovalHandler) UpdateApproval(ctx context.Context, c *app.RequestContext) {
	var req validation.UpdateApprovalRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	approval, err := h.svc.GetApproval(ctx, req.ApprovalID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "approval not found")
		return
	}
	if req.Status != "" {
		approval.Status = req.Status
	}
	if req.Comment != "" {
		approval.Comment = &req.Comment
	}
	if err := h.svc.Repo.Update(ctx, approval); err != nil {
		h.cfg.Error("failed to update approval", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to update approval")
		return
	}
	response.Success(c, approval)
}

// DeleteApproval deletes an approval
// @Summary      Delete approval
// @Description  Delete an approval record
// @Tags         Approvals
// @Accept       json
// @Produce      json
// @Param        request body     validation.DeleteApprovalRequest true  "Approval ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /api/v1/approvals/delete [post]
func (h *ApprovalHandler) DeleteApproval(ctx context.Context, c *app.RequestContext) {
	var req validation.DeleteApprovalRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.DeleteApproval(ctx, req.ApprovalID); err != nil {
		h.cfg.Error("failed to delete approval", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to delete approval")
		return
	}
	response.Success(c, map[string]string{"message": "approval deleted"})
}

// GetPendingApprovals gets the authenticated user's pending approvals.
// The approver identity always comes from the token, never the request body
// (IDOR prevention).
// @Summary      Get my pending approvals
// @Description  Get the authenticated user's pending approvals. Identity is
// @Description  taken from the token; the legacy body is tolerated but ignored.
// @Tags         Approvals
// @Accept       json
// @Produce      json
// @Param        request body     validation.GetPendingApprovalsRequest false  "Legacy body (ignored)"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/approvals/pending [post]
func (h *ApprovalHandler) GetPendingApprovals(ctx context.Context, c *app.RequestContext) {
	// Identity always comes from the authenticated token, never the body, so a
	// user can never read another approver's inbox (IDOR prevention). The body
	// is tolerated for backward compatibility but ignored.
	var req validation.GetPendingApprovalsRequest
	_ = c.BindAndValidate(&req)

	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	approvals, err := h.svc.GetPendingApprovals(ctx, userIDStr)
	if err != nil {
		h.cfg.Error("failed to get pending approvals", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get pending approvals")
		return
	}
	response.Success(c, approvals)
}

// Ensure pagination is imported
var _ = pagination.PaginationResponse{}
