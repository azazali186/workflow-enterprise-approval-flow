package messaging

import (
	"errors"
	"fmt"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// EventStream is the JetStream stream that durably captures every event the
// services and the saga orchestrator publish. Publishing through JetStream
// (with an ack) means a subscriber that is briefly down — or the publisher
// itself restarting — never silently loses a workflow event.
const EventStream = "APPROVAL_FLOW_EVENTS"

// EventSubjects is the set of subjects captured by the durable stream.
var EventSubjects = []string{
	"approval.>",
	"application.>",
	"workflow.>",
	"escalation.>",
	"notification.>",
	"template.>",
	"approval_needed",
}

type NATS struct {
	Conn   *nats.Conn
	Jet    nats.JetStreamContext
	logger *config.Config
}

func New(cfg *config.Config) (*NATS, error) {
	conn, err := nats.Connect(
		cfg.NATSURL,
		nats.MaxReconnects(10),
		nats.ReconnectWait(1*time.Second),
		nats.Timeout(10*time.Second),
		nats.PingInterval(20*time.Second),
		nats.MaxPingsOutstanding(2),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	jet, err := conn.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to init jetstream: %w", err)
	}

	// Ensure the durable event stream exists before anything publishes, so a
	// JetStream publish (which requires a matching stream) never fails at
	// request time. NATS must be started with JetStream enabled (-js).
	if err := ensureEventStream(jet); err != nil {
		return nil, fmt.Errorf("failed to ensure JetStream event stream: %w", err)
	}

	cfg.Info("nats connection established")
	return &NATS{Conn: conn, Jet: jet, logger: cfg}, nil
}

// ensureEventStream creates the event stream if it does not exist, or updates
// it to match the desired configuration. Idempotent, safe on every startup.
func ensureEventStream(jet nats.JetStreamContext) error {
	cfg := &nats.StreamConfig{
		Name:     EventStream,
		Subjects: EventSubjects,
		// Retain messages long enough for consumers to catch up after restarts.
		MaxAge: 24 * time.Hour,
	}

	if _, err := jet.StreamInfo(EventStream); err != nil {
		if errors.Is(err, nats.ErrStreamNotFound) {
			_, createErr := jet.AddStream(cfg)
			return createErr
		}
		return err
	}

	_, err := jet.UpdateStream(cfg)
	return err
}

func (n *NATS) Close() {
	if n.Conn != nil {
		n.Conn.Close()
	}
}

// Publish durably publishes an event via JetStream. The publish is
// asynchronous: the request path never blocks on a stream ack, so a degraded
// NATS cannot stall API responses (it only delays delivery). JetStream still
// retains the message for the durable stream, and core-NATS subscribers (the
// saga orchestrator) receive it live. A failed stream ack is logged.
//
// Note: PublishAsync must NOT be given a nats.AckWait option — it conflicts
// with the internal context PublishAsync uses and nats.go rejects the call
// with ErrContextAndTimeout ("nats: context and timeout can not both be set"),
// which silently broke every event publish. The async future carries its own
// ack wait, and the server-side JetStream ack wait defaults to 30s.
func (n *NATS) Publish(subject string, data []byte) error {
	future, err := n.Jet.PublishAsync(subject, data)
	if err != nil {
		return err
	}
	go func() {
		select {
		case <-future.Ok():
			// Acked by the stream.
		case err := <-future.Err():
			if n.logger != nil {
				n.logger.Error("nats publish ack failed",
					zap.String("subject", subject),
					zap.Error(err),
				)
			}
		}
	}()
	return nil
}

func (n *NATS) Request(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
	return n.Conn.Request(subject, data, timeout)
}

func (n *NATS) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return n.Conn.Subscribe(subject, handler)
}

func (n *NATS) QueueSubscribe(subject, group string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return n.Conn.QueueSubscribe(subject, group, handler)
}

func (n *NATS) PublishAsync(subject string, data []byte) (nats.PubAckFuture, error) {
	return n.Jet.PublishAsync(subject, data)
}

func (n *NATS) CreateStream(name string, subjects []string) error {
	_, err := n.Jet.AddStream(&nats.StreamConfig{
		Name:     name,
		Subjects: subjects,
	})
	return err
}

// EnsureStream creates the stream if it does not exist, or updates it to match
// the desired configuration. AddStream alone fails when the stream already
// exists, so this is the idempotent variant safe to call on every startup.
func (n *NATS) EnsureStream(name string, subjects []string) error {
	cfg := &nats.StreamConfig{
		Name:     name,
		Subjects: subjects,
	}

	if _, err := n.Jet.StreamInfo(name); err != nil {
		if errors.Is(err, nats.ErrStreamNotFound) {
			_, createErr := n.Jet.AddStream(cfg)
			return createErr
		}
		return err
	}

	_, err := n.Jet.UpdateStream(cfg)
	return err
}

func (n *NATS) CreateConsumer(stream, consumer string) error {
	_, err := n.Jet.AddConsumer(stream, &nats.ConsumerConfig{
		Durable:    consumer,
		AckPolicy:  nats.AckExplicitPolicy,
		MaxDeliver: 3,
	})
	return err
}

// Ping checks if NATS connection is alive
func (n *NATS) Ping() error {
	if n.Conn == nil {
		return fmt.Errorf("nats connection is nil")
	}
	if !n.Conn.IsConnected() {
		return fmt.Errorf("nats not connected")
	}
	return nil
}
