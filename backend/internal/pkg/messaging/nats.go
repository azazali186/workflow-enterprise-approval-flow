package messaging

import (
	"errors"
	"fmt"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/nats-io/nats.go"
)

type NATS struct {
	Conn *nats.Conn
	Jet  nats.JetStreamContext
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

	cfg.Info("nats connection established")
	return &NATS{Conn: conn, Jet: jet}, nil
}

func (n *NATS) Close() {
	if n.Conn != nil {
		n.Conn.Close()
	}
}

func (n *NATS) Publish(subject string, data []byte) error {
	return n.Conn.Publish(subject, data)
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
