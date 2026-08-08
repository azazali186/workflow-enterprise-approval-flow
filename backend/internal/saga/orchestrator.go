package saga

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/modules/approval"
	"github.com/aeroxe/approval-flow/internal/modules/escalation"
	"github.com/aeroxe/approval-flow/internal/modules/notification"
	"github.com/aeroxe/approval-flow/internal/modules/rbac"
	"github.com/aeroxe/approval-flow/internal/modules/workflow"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
	"github.com/aeroxe/approval-flow/internal/pkg/messaging"
	pkguuid "github.com/aeroxe/approval-flow/internal/pkg/uuid"
	"github.com/aeroxe/approval-flow/internal/pkg/websocket"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// Saga types matching README
const (
	SagaTypeApplicationSubmission = "application_submission"
	SagaTypeApprovalRouting       = "approval_routing"
	SagaTypeEscalationProcess     = "escalation_process"
	SagaTypeDecisionNotification  = "decision_notification"
)

// Saga statuses
const (
	SagaStatusStarted     = "started"
	SagaStatusRunning     = "running"
	SagaStatusCompleted   = "completed"
	SagaStatusFailed      = "failed"
	SagaStatusCompensated = "compensated"
)

// SagaState represents the state of a saga
type SagaState struct {
	SagaID     string                 `json:"saga_id"`
	SagaType   string                 `json:"saga_type"`
	Status     string                 `json:"status"`
	Step       int                    `json:"step"`
	TotalSteps int                    `json:"total_steps"`
	Data       map[string]interface{} `json:"data"`
	Error      string                 `json:"error,omitempty"`
	CreatedAt  string                 `json:"created_at"`
	UpdatedAt  string                 `json:"updated_at"`
}

// Orchestrator manages saga execution and the SLA escalation monitor.
type Orchestrator struct {
	nats            *messaging.NATS
	cache           *cache.Redis
	hub             *websocket.Hub
	logger          *config.Config
	escalationSvc   *escalation.Service
	notificationSvc *notification.Service
	approvalSvc     *approval.Service
	workflowSvc     *workflow.Service
	rbacRepo        *rbac.Repository
}

// NewOrchestrator creates a new saga orchestrator. The extra services power
// the SLA escalation monitor, which turns overdue approvals into real
// escalations.
func NewOrchestrator(
	nats *messaging.NATS,
	cache *cache.Redis,
	hub *websocket.Hub,
	cfg *config.Config,
	escalationSvc *escalation.Service,
	notificationSvc *notification.Service,
	approvalSvc *approval.Service,
	workflowSvc *workflow.Service,
	rbacRepo *rbac.Repository,
) *Orchestrator {
	return &Orchestrator{
		nats:            nats,
		cache:           cache,
		hub:             hub,
		logger:          cfg,
		escalationSvc:   escalationSvc,
		notificationSvc: notificationSvc,
		approvalSvc:     approvalSvc,
		workflowSvc:     workflowSvc,
		rbacRepo:        rbacRepo,
	}
}

// Start begins the saga orchestrator, subscribing to NATS events
func (o *Orchestrator) Start(ctx context.Context) error {
	// Subscribe to all saga-related events
	if _, err := o.nats.Subscribe("application.submitted", o.handleApplicationSubmitted); err != nil {
		o.logger.Error("failed to subscribe to application.submitted", zap.Error(err))
	}
	if _, err := o.nats.Subscribe("approval.created", o.handleApprovalCreated); err != nil {
		o.logger.Error("failed to subscribe to approval.created", zap.Error(err))
	}
	if _, err := o.nats.Subscribe("approval.decided", o.handleApprovalDecided); err != nil {
		o.logger.Error("failed to subscribe to approval.decided", zap.Error(err))
	}
	if _, err := o.nats.Subscribe("escalation.created", o.handleEscalationCreated); err != nil {
		o.logger.Error("failed to subscribe to escalation.created", zap.Error(err))
	}
	if _, err := o.nats.Subscribe("notification.created", o.handleNotificationCreated); err != nil {
		o.logger.Error("failed to subscribe to notification.created", zap.Error(err))
	}

	// Start background SLA monitor for escalation
	go o.startSLAMonitor(ctx)

	o.logger.Info("saga orchestrator started",
		zap.String("subscriptions", "application.submitted, approval.created, approval.decided, escalation.created, notification.created"),
	)
	return nil
}

// ==================== Saga State Management ====================

// StartSaga initializes a new saga
func (o *Orchestrator) StartSaga(ctx context.Context, sagaID, sagaType string, data map[string]interface{}, totalSteps int) error {
	state := &SagaState{
		SagaID:     sagaID,
		SagaType:   sagaType,
		Status:     SagaStatusStarted,
		Step:       0,
		TotalSteps: totalSteps,
		Data:       data,
		CreatedAt:  time.Now().Format(time.RFC3339),
		UpdatedAt:  time.Now().Format(time.RFC3339),
	}

	stateJSON, _ := json.Marshal(state)
	if err := o.cache.Set(ctx, fmt.Sprintf("saga:%s", sagaID), string(stateJSON), 24*time.Hour); err != nil {
		return fmt.Errorf("failed to save saga state: %w", err)
	}

	o.logger.Info("saga started",
		zap.String("saga_id", sagaID),
		zap.String("saga_type", sagaType),
		zap.Int("total_steps", totalSteps),
	)
	return nil
}

// UpdateSagaStep updates the current step of a saga
func (o *Orchestrator) UpdateSagaStep(ctx context.Context, sagaID string, step int, data map[string]interface{}) error {
	state, err := o.getSagaState(ctx, sagaID)
	if err != nil {
		return err
	}

	state.Step = step
	state.Status = SagaStatusRunning
	if data != nil {
		// Merge data
		for k, v := range data {
			state.Data[k] = v
		}
	}
	state.UpdatedAt = time.Now().Format(time.RFC3339)

	return o.saveSagaState(ctx, state)
}

// CompleteSaga marks a saga as completed
func (o *Orchestrator) CompleteSaga(ctx context.Context, sagaID string) error {
	state, err := o.getSagaState(ctx, sagaID)
	if err != nil {
		return err
	}

	state.Status = SagaStatusCompleted
	state.UpdatedAt = time.Now().Format(time.RFC3339)

	o.logger.Info("saga completed",
		zap.String("saga_id", sagaID),
		zap.String("saga_type", state.SagaType),
		zap.Int("steps_completed", state.Step),
	)

	return o.saveSagaState(ctx, state)
}

// FailSaga marks a saga as failed
func (o *Orchestrator) FailSaga(ctx context.Context, sagaID, errMsg string) error {
	state, err := o.getSagaState(ctx, sagaID)
	if err != nil {
		return err
	}

	state.Status = SagaStatusFailed
	state.Error = errMsg
	state.UpdatedAt = time.Now().Format(time.RFC3339)

	o.logger.Error("saga failed",
		zap.String("saga_id", sagaID),
		zap.String("saga_type", state.SagaType),
		zap.String("error", errMsg),
	)

	return o.saveSagaState(ctx, state)
}

// getSagaState retrieves saga state from Redis
func (o *Orchestrator) getSagaState(ctx context.Context, sagaID string) (*SagaState, error) {
	cacheKey := fmt.Sprintf("saga:%s", sagaID)
	stateJSON, err := o.cache.Get(ctx, cacheKey)
	if err != nil {
		return nil, fmt.Errorf("saga not found: %w", err)
	}

	var state SagaState
	if err := json.Unmarshal([]byte(stateJSON), &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal saga state: %w", err)
	}

	return &state, nil
}

// saveSagaState saves saga state to Redis
func (o *Orchestrator) saveSagaState(ctx context.Context, state *SagaState) error {
	stateJSON, _ := json.Marshal(state)
	cacheKey := fmt.Sprintf("saga:%s", state.SagaID)
	return o.cache.Set(ctx, cacheKey, string(stateJSON), 24*time.Hour)
}

// ==================== Saga 1: ApplicationSubmission ====================
// Flow: submission → validation → routing → approval/rejection → notification

func (o *Orchestrator) handleApplicationSubmitted(msg *nats.Msg) {
	var data map[string]interface{}
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		o.logger.Error("failed to unmarshal application.submitted", zap.Error(err))
		return
	}

	applicationID, _ := data["application_id"].(string)
	sagaID := fmt.Sprintf("application_submission:%s", applicationID)

	ctx := context.Background()

	// Step 1: Start Saga
	if err := o.StartSaga(ctx, sagaID, SagaTypeApplicationSubmission, data, 4); err != nil {
		o.logger.Error("failed to start application_submission saga", zap.Error(err))
		return
	}

	// Step 2: Validate Application
	o.logger.Info("application_submission: validating", zap.String("application_id", applicationID))
	o.UpdateSagaStep(ctx, sagaID, 1, map[string]interface{}{
		"step_name": "validation",
		"status":    "completed",
	})

	// Step 3: Route to Approval (publish approval_needed event)
	o.logger.Info("application_submission: routing to approval", zap.String("application_id", applicationID))
	approvalNeededData, _ := json.Marshal(map[string]interface{}{
		"application_id": applicationID,
		"saga_id":        sagaID,
	})
	o.nats.Publish("approval_needed", approvalNeededData)

	// Broadcast WebSocket event
	if o.hub != nil {
		o.hub.SendToAll("application_submitted", map[string]interface{}{
			"application_id": applicationID,
			"saga_id":        sagaID,
			"status":         "submitted",
		})
	}

	o.UpdateSagaStep(ctx, sagaID, 2, map[string]interface{}{
		"step_name": "routing",
		"status":    "completed",
	})

	o.logger.Info("application_submission: saga progressing", zap.String("saga_id", sagaID))
}

// ==================== Saga 2: ApprovalRouting ====================
// Flow: receive approval_needed → assign approver → wait for decision

func (o *Orchestrator) handleApprovalCreated(msg *nats.Msg) {
	var data map[string]interface{}
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		o.logger.Error("failed to unmarshal approval.created", zap.Error(err))
		return
	}

	approvalID, _ := data["approval_id"].(string)
	applicationID, _ := data["application_id"].(string)

	ctx := context.Background()
	sagaID := fmt.Sprintf("approval_routing:%s", approvalID)

	// Start ApprovalRouting saga
	if err := o.StartSaga(ctx, sagaID, SagaTypeApprovalRouting, data, 3); err != nil {
		o.logger.Error("failed to start approval_routing saga", zap.Error(err))
		return
	}

	// Step 1: Assign approver
	o.logger.Info("approval_routing: assigning approver",
		zap.String("approval_id", approvalID),
		zap.String("application_id", applicationID),
	)
	o.UpdateSagaStep(ctx, sagaID, 1, map[string]interface{}{
		"step_name": "assign_approver",
		"status":    "completed",
	})

	// Step 2: Notify approver via WebSocket
	if o.hub != nil {
		approverID, _ := data["approver_id"].(string)
		if approverID != "" {
			o.hub.SendToUser(approverID, "approval_needed", map[string]interface{}{
				"approval_id":    approvalID,
				"application_id": applicationID,
				"status":         "pending",
			})
		}
	}

	o.UpdateSagaStep(ctx, sagaID, 2, map[string]interface{}{
		"step_name": "notify_approver",
		"status":    "completed",
	})

	o.logger.Info("approval_routing: saga completed", zap.String("saga_id", sagaID))
	o.CompleteSaga(ctx, sagaID)
}

// ==================== Saga 3: DecisionNotification ====================
// Flow: receive approval.decided → update application → notify applicant

func (o *Orchestrator) handleApprovalDecided(msg *nats.Msg) {
	var data map[string]interface{}
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		o.logger.Error("failed to unmarshal approval.decided", zap.Error(err))
		return
	}

	approvalID, _ := data["approval_id"].(string)
	decision, _ := data["decision"].(string)
	applicationID, _ := data["application_id"].(string)

	ctx := context.Background()
	sagaID := fmt.Sprintf("decision_notification:%s", approvalID)

	// Start DecisionNotification saga
	if err := o.StartSaga(ctx, sagaID, SagaTypeDecisionNotification, data, 3); err != nil {
		o.logger.Error("failed to start decision_notification saga", zap.Error(err))
		return
	}

	// Step 1: Update application status based on decision
	o.logger.Info("decision_notification: updating application",
		zap.String("approval_id", approvalID),
		zap.String("decision", decision),
	)
	o.UpdateSagaStep(ctx, sagaID, 1, map[string]interface{}{
		"step_name": "update_application",
		"status":    "completed",
		"decision":  decision,
	})

	// Step 2: Send notification to applicant
	notificationData, _ := json.Marshal(map[string]interface{}{
		"application_id": applicationID,
		"approval_id":    approvalID,
		"decision":       decision,
		"type":           "decision",
	})
	o.nats.Publish("notification.created", notificationData)

	o.UpdateSagaStep(ctx, sagaID, 2, map[string]interface{}{
		"step_name": "send_notification",
		"status":    "completed",
	})

	// Step 3: Broadcast WebSocket event
	if o.hub != nil {
		o.hub.SendToAll("decision_made", map[string]interface{}{
			"approval_id":    approvalID,
			"application_id": applicationID,
			"decision":       decision,
		})
	}

	o.UpdateSagaStep(ctx, sagaID, 3, map[string]interface{}{
		"step_name": "broadcast_event",
		"status":    "completed",
	})

	// If rejected, complete the parent application_submission saga
	if decision == "rejected" {
		parentSagaID := fmt.Sprintf("application_submission:%s", applicationID)
		o.CompleteSaga(ctx, parentSagaID)
	}

	o.CompleteSaga(ctx, sagaID)
	o.logger.Info("decision_notification: saga completed",
		zap.String("saga_id", sagaID),
		zap.String("decision", decision),
	)
}

// ==================== Saga 4: EscalationProcess ====================
// Flow: receive escalation.created → check SLA → escalate → notify

func (o *Orchestrator) handleEscalationCreated(msg *nats.Msg) {
	var data map[string]interface{}
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		o.logger.Error("failed to unmarshal escalation.created", zap.Error(err))
		return
	}

	escalationID, _ := data["escalation_id"].(string)
	approvalID, _ := data["approval_id"].(string)

	ctx := context.Background()
	sagaID := fmt.Sprintf("escalation_process:%s", escalationID)

	// Start EscalationProcess saga
	if err := o.StartSaga(ctx, sagaID, SagaTypeEscalationProcess, data, 3); err != nil {
		o.logger.Error("failed to start escalation_process saga", zap.Error(err))
		return
	}

	// Step 1: Check SLA breach
	o.logger.Info("escalation_process: checking SLA breach",
		zap.String("escalation_id", escalationID),
		zap.String("approval_id", approvalID),
	)
	o.UpdateSagaStep(ctx, sagaID, 1, map[string]interface{}{
		"step_name": "check_sla",
		"status":    "completed",
	})

	// Step 2: Route to higher authority
	o.UpdateSagaStep(ctx, sagaID, 2, map[string]interface{}{
		"step_name": "route_to_authority",
		"status":    "completed",
	})

	// Step 3: Notify via WebSocket
	if o.hub != nil {
		escalatedTo, _ := data["escalated_to"].(string)
		if escalatedTo != "" {
			o.hub.SendToUser(escalatedTo, "escalation_trigger", map[string]interface{}{
				"escalation_id": escalationID,
				"approval_id":   approvalID,
				"level":         data["level"],
				"reason":        data["reason"],
			})
		}
		// Also broadcast to all connected clients
		o.hub.SendToAll("escalation_trigger", map[string]interface{}{
			"escalation_id": escalationID,
			"approval_id":   approvalID,
			"status":        "escalated",
		})
	}

	o.UpdateSagaStep(ctx, sagaID, 3, map[string]interface{}{
		"step_name": "notify_stakeholders",
		"status":    "completed",
	})

	o.CompleteSaga(ctx, sagaID)
	o.logger.Info("escalation_process: saga completed", zap.String("saga_id", sagaID))
}

// ==================== Notification Handler ====================

func (o *Orchestrator) handleNotificationCreated(msg *nats.Msg) {
	var data map[string]interface{}
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		o.logger.Error("failed to unmarshal notification.created", zap.Error(err))
		return
	}

	notificationID, _ := data["notification_id"].(string)
	userID, _ := data["user_id"].(string)

	o.logger.Info("notification.created: broadcasting",
		zap.String("notification_id", notificationID),
		zap.String("user_id", userID),
	)

	// Send to specific user via WebSocket
	if o.hub != nil && userID != "" {
		o.hub.SendToUser(userID, "notification", data)
	}
}

// ==================== SLA Monitor ====================

// startSLAMonitor runs a background goroutine that checks for overdue approvals
func (o *Orchestrator) startSLAMonitor(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			o.checkOverdueApprovals(ctx)
		}
	}
}

// checkOverdueApprovals escalates every pending approval whose step SLA has
// expired. Each breach creates an Escalation record, bumps the approval's
// escalation level, and notifies the escalation target.
func (o *Orchestrator) checkOverdueApprovals(ctx context.Context) {
	if o.approvalSvc == nil || o.escalationSvc == nil {
		return
	}
	overdue, err := o.approvalSvc.Repo.ListOverduePending(ctx, time.Now())
	if err != nil {
		o.logger.Error("SLA monitor: failed to query overdue approvals", zap.Error(err))
		return
	}
	for i := range overdue {
		o.escalateOverdue(ctx, &overdue[i])
	}
}

// escalateOverdue creates the escalation for a single overdue approval.
func (o *Orchestrator) escalateOverdue(ctx context.Context, a *domain.Approval) {
	target, err := o.resolveEscalationTarget(ctx, a)
	if err != nil {
		o.logger.Error("SLA monitor: failed to resolve escalation target",
			zap.String("approval_id", a.ID.String()), zap.Error(err))
		return
	}

	level := a.EscalationLevel + 1
	reason := "approval exceeded its SLA (no decision within the step timeout)"
	if err := o.escalationSvc.Escalate(ctx, a.ID.String(), level, target, reason); err != nil {
		o.logger.Error("SLA monitor: escalation failed",
			zap.String("approval_id", a.ID.String()), zap.Error(err))
		return
	}

	a.EscalationLevel = level
	if err := o.approvalSvc.Repo.Update(ctx, a); err != nil {
		o.logger.Error("SLA monitor: failed to update escalation level", zap.Error(err))
	}

	if o.notificationSvc != nil {
		now := time.Now()
		n := &domain.Notification{
			UserID:  target,
			Type:    "escalation",
			Channel: "in_app",
			Title:   "Approval escalated (SLA)",
			Body:    fmt.Sprintf("Approval %s was escalated to level %d.", a.ID.String(), level),
			SentAt:  &now,
		}
		if err := o.notificationSvc.SendNotification(ctx, n); err != nil {
			o.logger.Error("SLA monitor: failed to notify escalation target", zap.Error(err))
		}
	}

	o.logger.Warn("SLA monitor: approval escalated",
		zap.String("approval_id", a.ID.String()),
		zap.Int("level", level),
	)
}

// resolveEscalationTarget picks an active admin as the escalation owner,
// falling back to the approval's own approver.
func (o *Orchestrator) resolveEscalationTarget(ctx context.Context, a *domain.Approval) (uuid.UUID, error) {
	if o.rbacRepo != nil {
		admins, err := o.rbacRepo.GetUsersByRole(ctx, "admin")
		if err == nil && len(admins) > 0 {
			return admins[0].ID, nil
		}
	}
	return a.ApproverID, nil
}

// ==================== Helper: Generate Saga ID ====================

// GenerateSagaID creates a unique saga ID
func GenerateSagaID(sagaType string, entityID string) string {
	if entityID != "" {
		return fmt.Sprintf("%s:%s", sagaType, entityID)
	}
	return fmt.Sprintf("%s:%s", sagaType, pkguuid.GenerateID())
}
