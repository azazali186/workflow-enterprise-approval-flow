package websocket

import (
	"context"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/stretchr/testify/assert"
)

// mockConn is a thread-safe in-memory Conn used to exercise the hub.
type mockConn struct {
	mu     sync.Mutex
	closed bool
	writes [][]byte
}

func (m *mockConn) Read(p []byte) (int, error) { return 0, io.EOF }

func (m *mockConn) Write(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, io.ErrClosedPipe
	}
	m.writes = append(m.writes, append([]byte(nil), p...))
	return len(p), nil
}

func (m *mockConn) Close(code int, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil
	}
	m.closed = true
	return nil
}

func (m *mockConn) Ping(ctx context.Context) error { return nil }
func (m *mockConn) Pong(ctx context.Context) error { return nil }

func newTestHub(maxConnections int) *Hub {
	cfg := &config.Config{
		WSPingInterval:   30,
		WSMaxConnections: maxConnections,
		Logger:           config.NewNopLogger(),
	}
	hub := NewHub(cfg)
	go hub.Run()
	return hub
}

// waitFor waits until cond() returns true or the timeout elapses.
func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for !cond() && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
}

func TestHubConcurrentOperations(t *testing.T) {
	hub := newTestHub(100)
	t.Cleanup(hub.Shutdown)

	const numClients = 40
	for i := 0; i < numClients; i++ {
		conn := &mockConn{}
		client := &Client{
			ID:     fmt.Sprintf("client-%d", i),
			UserID: fmt.Sprintf("user-%d", i%5),
			Conn:   conn,
			Send:   make(chan []byte, 16),
			Hub:    hub,
		}
		hub.Register <- client
		go client.ReadPump()
		go client.WritePump()
	}

	// Concurrently fan out broadcast + per-user messages while clients
	// register/unregister. Run under -race to catch data races and
	// send-on-closed-channel panics.
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			hub.SendToAll("test.event", map[string]interface{}{"n": n})
			hub.SendToUser("user-2", "test.user", map[string]interface{}{"n": n})
		}(i)
	}
	wg.Wait()

	// All mock conns return io.EOF immediately, so clients unregister on their
	// own. Wait for the hub to drain.
	waitFor(t, 2*time.Second, func() bool { return hub.Len() == 0 })
	assert.Equal(t, 0, hub.Len())
}

func TestHubMaxConnections(t *testing.T) {
	hub := newTestHub(2)
	t.Cleanup(hub.Shutdown)

	conns := make([]*mockConn, 3)
	for i := 0; i < 3; i++ {
		conn := &mockConn{}
		conns[i] = conn
		client := &Client{
			ID:     fmt.Sprintf("c-%d", i),
			UserID: fmt.Sprintf("u-%d", i),
			Conn:   conn,
			Send:   make(chan []byte, 4),
			Hub:    hub,
		}
		hub.Register <- client
	}

	waitFor(t, time.Second, func() bool { return hub.Len() == 2 })
	assert.Equal(t, 2, hub.Len(), "only maxConnections clients should be accepted")

	// The third (rejected) client's connection must have been closed.
	waitFor(t, time.Second, func() bool { return conns[2].isClosed() })
	assert.True(t, conns[2].isClosed(), "rejected client connection should be closed")
	assert.False(t, conns[0].isClosed(), "accepted client connection should remain open")
}

func (m *mockConn) isClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

// blockingConn blocks all Read/Write calls until released, simulating a very
// slow or stuck client connection.
type blockingConn struct {
	mu      sync.Mutex
	closed  bool
	release chan struct{}
	closeFn sync.Once
}

func (b *blockingConn) Read(p []byte) (int, error) {
	<-b.release
	return 0, io.EOF
}

func (b *blockingConn) Write(p []byte) (int, error) {
	<-b.release
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return 0, io.ErrClosedPipe
	}
	return len(p), nil
}

func (b *blockingConn) Close(code int, reason string) error {
	b.closeFn.Do(func() {
		b.mu.Lock()
		b.closed = true
		b.mu.Unlock()
		close(b.release)
	})
	return nil
}

func (b *blockingConn) Ping(ctx context.Context) error { return nil }
func (b *blockingConn) Pong(ctx context.Context) error { return nil }

// TestHubEvictsSlowClient exercises the full-buffer eviction path: a client
// whose writes are blocked must be evicted by closing its Send channel without
// panicking (send-on-closed-channel). Run under -race in CI.
func TestHubEvictsSlowClient(t *testing.T) {
	hub := newTestHub(10)
	conn := &blockingConn{release: make(chan struct{})}
	t.Cleanup(func() {
		conn.Close(1000, "test done")
		hub.Shutdown()
	})

	client := &Client{
		ID:     "slow-client",
		UserID: "slow-user",
		Conn:   conn,
		Send:   make(chan []byte, 2),
		Hub:    hub,
	}
	hub.Register <- client
	waitFor(t, time.Second, func() bool { return hub.Len() == 1 })

	// WritePump blocks on the stuck conn; with a buffer of 2, the third
	// message forces the hub to evict the client.
	go client.WritePump()

	for i := 0; i < 10; i++ {
		hub.SendToUser("slow-user", "evict.me", map[string]interface{}{"i": i})
	}

	// The client must be evicted even though its connection never drains.
	waitFor(t, 2*time.Second, func() bool { return hub.Len() == 0 })
	assert.Equal(t, 0, hub.Len(), "slow client should be evicted")
}

func TestHubShutdownClosesAllClients(t *testing.T) {
	hub := newTestHub(10)

	client := &Client{
		ID:     "c-1",
		UserID: "u-1",
		Conn:   &mockConn{},
		Send:   make(chan []byte, 4),
		Hub:    hub,
	}
	hub.Register <- client
	waitFor(t, time.Second, func() bool { return hub.Len() == 1 })

	hub.Shutdown()

	waitFor(t, time.Second, func() bool { return hub.Len() == 0 })
	assert.Equal(t, 0, hub.Len(), "shutdown should clear all clients")
}
