package saga

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
	"github.com/aeroxe/approval-flow/internal/pkg/messaging"
	"github.com/nats-io/nats.go"
)

type SagaState struct {
	SagaID    string                 `json:"saga_id"`
	Status    string                 `json:"status"`
	Step      int                    `json:"step"`
	Data      map[string]interface{} `json:"data"`
	Error     string                 `json:"error,omitempty"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

type Orchestrator struct {
	nats   *messaging.NATS
	cache  *cache.Redis
	logger *config.Config
}

func NewOrchestrator(nats *messaging.NATS, cache *cache.Redis, cfg *config.Config) *Orchestrator {
	return &Orchestrator{nats: nats, cache: cache, logger: cfg}
}

func (o *Orchestrator) Start(ctx context.Context) error {
	o.nats.Subscribe("application.submitted", o.handleApplicationSubmitted)
	o.nats.Subscribe("approval.decided", o.handleApprovalDecided)
	o.nats.Subscribe("escalation.created", o.handleEscalationCreated)
	o.logger.Info("saga orchestrator started")
	return nil
}

func (o *Orchestrator) StartSaga(ctx context.Context, sagaID, sagaType string, data map[string]interface{}) error {
	state := &SagaState{
		SagaID:    sagaID,
		Status:    "started",
		Step:      0,
		Data:      data,
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	stateJSON, _ := json.Marshal(state)
	if err := o.cache.Set(ctx, fmt.Sprintf("saga:%s", sagaID), string(stateJSON), 0); err != nil {
		return fmt.Errorf("failed to save saga state: %w", err)
	}

	o.logger.Info("saga started", "saga_id", sagaID, "saga_type", sagaType)
	return nil
}

func (o *Orchestrator) UpdateSagaStep(ctx context.Context, sagaID string, step int, data map[string]interface{}) error {
	cacheKey := fmt.Sprintf("saga:%s", sagaID)
	stateJSON, err := o.cache.Get(ctx, cacheKey)
	if err != nil {
		return fmt.Errorf("saga not found: %w", err)
	}

	var state SagaState
	if err := json.Unmarshal([]byte(stateJSON), &state); err != nil {
		return fmt.Errorf("failed to unmarshal saga state: %w", err)
	}

	state.Step = step
	state.Data = data
	state.UpdatedAt = time.Now().Format(time.RFC3339)

	newStateJSON, _ := json.Marshal(state)
	return o.cache.Set(ctx, cacheKey, string(newStateJSON), 0)
}

func (o *Orchestrator) CompleteSaga(ctx context.Context, sagaID string) error {
	cacheKey := fmt.Sprintf("saga:%s", sagaID)
	stateJSON, err := o.cache.Get(ctx, cacheKey)
	if err != nil {
		return fmt.Errorf("saga not found: %w", err)
	}

	var state SagaState
	if err := json.Unmarshal([]byte(stateJSON), &state); err != nil {
		return fmt.Errorf("failed to unmarshal saga state: %w", err)
	}

	state.Status = "completed"
	state.UpdatedAt = time.Now().Format(time.RFC3339)

	newStateJSON, _ := json.Marshal(state)
	return o.cache.Set(ctx, cacheKey, string(newStateJSON), 0)
}

func (o *Orchestrator) FailSaga(ctx context.Context, sagaID, errMsg string) error {
	cacheKey := fmt.Sprintf("saga:%s", sagaID)
	stateJSON, err := o.cache.Get(ctx, cacheKey)
	if err != nil {
		return fmt.Errorf("saga not found: %w", err)
	}

	var state SagaState
	if err := json.Unmarshal([]byte(stateJSON), &state); err != nil {
		return fmt.Errorf("failed to unmarshal saga state: %w", err)
	}

	state.Status = "failed"
	state.Error = errMsg
	state.UpdatedAt = time.Now().Format(time.RFC3339)

	newStateJSON, _ := json.Marshal(state)
	return o.cache.Set(ctx, cacheKey, string(newStateJSON), 0)
}

func (o *Orchestrator) handleApplicationSubmitted(msg *nats.Msg) {
	var data map[string]interface{}
	json.Unmarshal(msg.Data, &data)
	applicationID := data["application_id"].(string)
	o.StartSaga(context.Background(), fmt.Sprintf("application_submission:%s", applicationID), "application_submission", data)
}

func (o *Orchestrator) handleApprovalDecided(msg *nats.Msg) {
	var data map[string]interface{}
	json.Unmarshal(msg.Data, &data)
	approvalID := data["approval_id"].(string)
	decision := data["decision"].(string)

	if decision == "rejected" {
		o.CompleteSaga(context.Background(), fmt.Sprintf("application_submission:%s", data["application_id"]))
	}
}

func (o *Orchestrator) handleEscalationCreated(msg *nats.Msg) {
	var data map[string]interface{}
	json.Unmarshal(msg.Data, &data)
	escalationID := data["escalation_id"].(string)
	o.logger.Info("escalation event received", "escalation_id", escalationID)
}
