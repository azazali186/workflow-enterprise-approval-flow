package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
	"github.com/aeroxe/approval-flow/internal/pkg/messaging"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// OutboxEvent represents an event to be published
type OutboxEvent struct {
	ID        string                 `json:"id"`
	Aggregate string                 `json:"aggregate"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
	Status    string                 `json:"status"` // pending, published, failed
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Retries   int                    `json:"retries"`
}

// Outbox implements the transactional outbox pattern
type Outbox struct {
	redis  *cache.Redis
	nats   *messaging.NATS
	cfg    *config.Config
}

// NewOutbox creates a new outbox instance
func NewOutbox(redis *cache.Redis, nats *messaging.NATS, cfg *config.Config) *Outbox {
	return &Outbox{
		redis: redis,
		nats:  nats,
		cfg:   cfg,
	}
}

// PublishEvent stores an event in the outbox for reliable publishing
func (o *Outbox) PublishEvent(ctx context.Context, aggregate, eventType string, payload map[string]interface{}) error {
	event := OutboxEvent{
		ID:        uuid.New().String(),
		Aggregate: aggregate,
		Type:      eventType,
		Payload:   payload,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Retries:   0,
	}

	// Store in Redis as pending
	eventKey := fmt.Sprintf("outbox:event:%s", event.ID)
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := o.redis.Set(ctx, eventKey, string(data), 24*time.Hour); err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}

	// Add to pending queue
	queueKey := "outbox:pending"
	if err := o.redis.Client.LPush(ctx, queueKey, event.ID).Err(); err != nil {
		return fmt.Errorf("failed to add to queue: %w", err)
	}

	o.cfg.Debug("event queued", zap.String("event_id", event.ID), zap.String("type", eventType))
	return nil
}

// ProcessPending processes all pending events in the outbox
func (o *Outbox) ProcessPending(ctx context.Context) error {
	queueKey := "outbox:pending"

	for {
		// Get next pending event
		eventID, err := o.redis.Client.RPop(ctx, queueKey).Result()
		if err != nil {
			if err.Error() == "redis: nil" {
				break // Queue empty
			}
			return fmt.Errorf("failed to get pending event: %w", err)
		}

		// Get event data
		eventKey := fmt.Sprintf("outbox:event:%s", eventID)
		eventData, err := o.redis.Get(ctx, eventKey)
		if err != nil {
			o.cfg.Warn("event not found", zap.String("event_id", eventID))
			continue
		}

		var event OutboxEvent
		if err := json.Unmarshal([]byte(eventData), &event); err != nil {
			o.cfg.Error("failed to unmarshal event", zap.Error(err))
			continue
		}

		// Try to publish
		subject := fmt.Sprintf("%s.%s", event.Aggregate, event.Type)
		payload, _ := json.Marshal(event.Payload)

		if err := o.nats.Publish(subject, payload); err != nil {
			o.cfg.Error("failed to publish event", zap.Error(err))

			// Retry logic
			event.Retries++
			event.UpdatedAt = time.Now()

			if event.Retries >= 3 {
				event.Status = "failed"
				o.cfg.Error("event failed after retries", zap.String("event_id", eventID))
			} else {
				// Re-queue for retry
				event.Status = "pending"
				o.redis.Client.LPush(ctx, queueKey, eventID)
			}

			// Update event
			updatedData, _ := json.Marshal(event)
			o.redis.Set(ctx, eventKey, string(updatedData), 24*time.Hour)
			continue
		}

		// Mark as published
		event.Status = "published"
		event.UpdatedAt = time.Now()
		updatedData, _ := json.Marshal(event)
		o.redis.Set(ctx, eventKey, string(updatedData), 1*time.Hour)

		o.cfg.Debug("event published", zap.String("event_id", eventID), zap.String("subject", subject))
	}

	return nil
}

// StartProcessor starts the outbox processor in background
func (o *Outbox) StartProcessor(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := o.ProcessPending(ctx); err != nil {
					o.cfg.Error("outbox processor error", zap.Error(err))
				}
			}
		}
	}()
}
