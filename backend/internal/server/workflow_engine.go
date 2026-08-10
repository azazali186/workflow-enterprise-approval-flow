package server

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/modules/application"
	"github.com/aeroxe/approval-flow/internal/modules/approval"
	"github.com/aeroxe/approval-flow/internal/modules/escalation"
	"github.com/aeroxe/approval-flow/internal/modules/notification"
	"github.com/aeroxe/approval-flow/internal/modules/rbac"
	"github.com/aeroxe/approval-flow/internal/modules/workflow"
	"github.com/aeroxe/approval-flow/internal/pkg/websocket"
)

// workflowEngine wires the domain modules into a real approval workflow:
//   - submitting an application creates the first approval task from the
//     workflow's ordered steps;
//   - approving advances to the next step, and the application completes when
//     all steps are decided;
//   - rejecting terminates the application;
//   - escalating marks the application as escalated (the SLA monitor and the
//     escalation module handle the rest).
//
// It lives in the server package (composition root) to avoid circular module
// imports, and is attached to the application/approval services as callbacks.
type workflowEngine struct {
	appSvc          *application.Service
	approvalSvc     *approval.Service
	workflowSvc     *workflow.Service
	escalationSvc   *escalation.Service
	notificationSvc *notification.Service
	rbacRepo        *rbac.Repository
	hub             *websocket.Hub
	cfg             *config.Config
}

func newWorkflowEngine(
	appSvc *application.Service,
	approvalSvc *approval.Service,
	workflowSvc *workflow.Service,
	escalationSvc *escalation.Service,
	notificationSvc *notification.Service,
	rbacRepo *rbac.Repository,
	hub *websocket.Hub,
	cfg *config.Config,
) *workflowEngine {
	return &workflowEngine{
		appSvc:          appSvc,
		approvalSvc:     approvalSvc,
		workflowSvc:     workflowSvc,
		escalationSvc:   escalationSvc,
		notificationSvc: notificationSvc,
		rbacRepo:        rbacRepo,
		hub:             hub,
		cfg:             cfg,
	}
}

// onSubmitted creates the first approval task for a submitted application.
func (e *workflowEngine) onSubmitted(ctx context.Context, applicationID string) error {
	app, err := e.appSvc.GetApplication(ctx, applicationID)
	if err != nil {
		return fmt.Errorf("load application: %w", err)
	}

	steps, err := e.workflowSvc.GetWorkflowSteps(ctx, app.WorkflowID.String())
	if err != nil {
		return fmt.Errorf("load workflow steps: %w", err)
	}
	if len(steps) == 0 {
		e.cfg.Info("application has no workflow steps; nothing to route",
			zap.String("application_id", applicationID))
		return nil
	}

	if _, err := e.createApprovalForStep(ctx, app, steps[0]); err != nil {
		return err
	}
	return nil
}

// onDecided advances the workflow after an approval decision. It runs
// synchronously so the API response reflects the new application state.
func (e *workflowEngine) onDecided(ctx context.Context, applicationID, decision, decidedApprovalID string) error {
	app, err := e.appSvc.GetApplication(ctx, applicationID)
	if err != nil {
		return fmt.Errorf("load application: %w", err)
	}

	// Realtime: notify the applicant the moment their application is decided.
	if e.hub != nil {
		e.hub.SendToUser(app.ApplicantID.String(), "decision_made", map[string]interface{}{
			"application_id": app.ID,
			"decision":       decision,
		})
	}

	now := time.Now()
	switch decision {
	case "rejected":
		app.Status = "rejected"
		app.CompletedAt = &now
		if err := e.appSvc.UpdateApplication(ctx, app); err != nil {
			return fmt.Errorf("reject application: %w", err)
		}
		e.notifyUser(ctx, app.ApplicantID.String(), "application",
			"Application rejected",
			fmt.Sprintf("Your application %q was rejected.", app.Title))
		return nil

	case "approved":
		decided, err := e.approvalSvc.GetApproval(ctx, decidedApprovalID)
		if err != nil {
			return fmt.Errorf("load decided approval: %w", err)
		}
		steps, err := e.workflowSvc.GetWorkflowSteps(ctx, app.WorkflowID.String())
		if err != nil {
			return fmt.Errorf("load workflow steps: %w", err)
		}
		if next := nextStep(steps, decided.WorkflowStepID.String()); next != nil {
			// Route to the next approval step.
			if _, err := e.createApprovalForStep(ctx, app, *next); err != nil {
				return err
			}
			return nil
		}
		// All steps decided — the application is approved.
		app.Status = "approved"
		app.CompletedAt = &now
		if err := e.appSvc.UpdateApplication(ctx, app); err != nil {
			return fmt.Errorf("complete application: %w", err)
		}
		e.notifyUser(ctx, app.ApplicantID.String(), "application",
			"Application approved",
			fmt.Sprintf("Your application %q was approved.", app.Title))
		return nil

	case "escalated":
		app.Status = "escalated"
		if err := e.appSvc.UpdateApplication(ctx, app); err != nil {
			return fmt.Errorf("escalate application: %w", err)
		}
		return nil
	}
	return nil
}

// nextStep returns the step that follows the decided step, or nil when the
// decided step is the last one.
func nextStep(steps []domain.WorkflowStep, decidedStepID string) *domain.WorkflowStep {
	for i := range steps {
		if steps[i].ID.String() == decidedStepID && i+1 < len(steps) {
			next := steps[i+1]
			return &next
		}
	}
	return nil
}

// createApprovalForStep resolves the step's approver and creates a pending
// approval task, notifying the approver. Returns (nil, nil) when no approver
// can be resolved — the step is logged and skipped.
func (e *workflowEngine) createApprovalForStep(ctx context.Context, app *domain.Application, step domain.WorkflowStep) (*domain.Approval, error) {
	approverID, err := e.resolveApprover(ctx, step)
	if err != nil {
		return nil, err
	}
	if approverID == nil {
		e.cfg.Warn("no approver resolvable for workflow step; skipping task",
			zap.String("step_id", step.ID.String()),
			zap.String("workflow_id", app.WorkflowID.String()),
			zap.String("application_id", app.ID.String()),
		)
		return nil, nil
	}

	approval := &domain.Approval{
		ApplicationID:  app.ID,
		WorkflowStepID: step.ID,
		ApproverID:     *approverID,
		Status:         "pending",
	}
	if err := e.approvalSvc.CreateApproval(ctx, approval); err != nil {
		return nil, fmt.Errorf("create approval task: %w", err)
	}

	e.notifyUser(ctx, approverID.String(), "approval",
		"Approval needed",
		fmt.Sprintf("An approval for %q is waiting on you (step: %s).", app.Title, step.Name))
	return approval, nil
}

// resolveApprover prefers the step's explicit approver; otherwise it resolves
// the first active user holding the step's approver role.
func (e *workflowEngine) resolveApprover(ctx context.Context, step domain.WorkflowStep) (*uuid.UUID, error) {
	if step.ApproverID != nil {
		return step.ApproverID, nil
	}
	if step.ApproverRole != "" {
		users, err := e.rbacRepo.GetUsersByRole(ctx, step.ApproverRole)
		if err != nil {
			return nil, fmt.Errorf("resolve approver by role: %w", err)
		}
		if len(users) > 0 {
			id := users[0].ID
			return &id, nil
		}
	}
	return nil, nil
}

func (e *workflowEngine) notifyUser(ctx context.Context, userID, ntype, title, body string) {
	uid, err := uuid.Parse(userID)
	if err != nil || e.notificationSvc == nil {
		return
	}
	now := time.Now()
	n := &domain.Notification{
		UserID:  uid,
		Type:    ntype,
		Channel: "in_app",
		Title:   title,
		Body:    body,
		SentAt:  &now,
	}
	if err := e.notificationSvc.SendNotification(ctx, n); err != nil {
		e.cfg.Error("failed to send workflow notification", zap.Error(err))
	}
}
