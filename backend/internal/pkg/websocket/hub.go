package websocket

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"sync"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"go.uber.org/zap"
)

// Conn defines the interface for WebSocket connections
type Conn interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close(code int, reason string) error
	Ping(ctx context.Context) error
	Pong(ctx context.Context) error
}

// Client represents a connected WebSocket client.
type Client struct {
	ID     string
	UserID string
	Conn   Conn
	Send   chan []byte
	Hub    *Hub
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close(1000, "")
	}()

	buf := make([]byte, 4096)
	for {
		n, err := c.Conn.Read(buf)
		if err != nil {
			if err != io.EOF && err != io.ErrClosedPipe {
				log.Printf("websocket read error: %v", err)
			}
			break
		}
		// Inbound frames are currently ignored (notifications are server-push only).
		_ = n
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(c.Hub.PingInterval)
	defer func() {
		ticker.Stop()
		c.Conn.Close(1000, "")
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			if !ok {
				// Channel closed: the hub evicted us or is shutting down.
				return
			}
			if _, err := c.Conn.Write(msg); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.Conn.Ping(context.Background()); err != nil {
				return
			}
		}
	}
}

// directMsg is a message targeted at a specific user (or all users when userID == "").
type directMsg struct {
	userID string
	data   []byte
}

// Hub manages all WebSocket clients. All mutations of the client map and all
// closes of client.Send channels happen inside Run, which is the single owner
// goroutine. This prevents the send-on-closed-channel races that a naive
// multi-goroutine implementation suffers from.
type Hub struct {
	Register       chan *Client
	Unregister     chan *Client
	Broadcast      chan []byte
	direct         chan directMsg
	clients        map[*Client]bool
	mu             sync.RWMutex
	maxConnections int
	PingInterval   time.Duration
	closed         chan struct{}
	closeOnce      sync.Once
	Logger         *config.Config
}

func NewHub(cfg *config.Config) *Hub {
	pingInterval := time.Duration(cfg.WSPingInterval) * time.Second
	if pingInterval <= 0 {
		pingInterval = 30 * time.Second
	}
	return &Hub{
		Register:       make(chan *Client, 256),
		Unregister:     make(chan *Client, 256),
		Broadcast:      make(chan []byte, 256),
		direct:         make(chan directMsg, 256),
		clients:        make(map[*Client]bool),
		maxConnections: cfg.WSMaxConnections,
		PingInterval:   pingInterval,
		closed:         make(chan struct{}),
		Logger:         cfg,
	}
}

// Run is the hub's single-owner event loop. It must be started once via `go hub.Run()`.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			if h.maxConnections > 0 && len(h.clients) >= h.maxConnections {
				h.Logger.Warn("websocket max connections reached, rejecting client",
					zap.String("user_id", client.UserID),
					zap.Int("max", h.maxConnections),
				)
				client.Conn.Close(1013, "too many connections")
				continue
			}
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.Logger.Info("client registered", zap.String("user_id", client.UserID))
		case client := <-h.Unregister:
			if h.remove(client) {
				h.Logger.Info("client unregistered", zap.String("user_id", client.UserID))
			}
		case msg := <-h.Broadcast:
			for client := range h.snapshot() {
				if !h.trySend(client, msg) {
					h.remove(client)
				}
			}
		case dm := <-h.direct:
			for client := range h.snapshot() {
				if dm.userID == "" || client.UserID == dm.userID {
					if !h.trySend(client, dm.data) {
						h.remove(client)
					}
				}
			}
		case <-h.closed:
			h.Logger.Info("websocket hub shutting down")
			h.mu.Lock()
			for client := range h.clients {
				client.Conn.Close(1001, "server shutting down")
				close(client.Send)
			}
			h.clients = make(map[*Client]bool)
			h.mu.Unlock()
			return
		}
	}
}

// snapshot returns a copy of the client map for iteration. Safe to call from Run only.
func (h *Hub) snapshot() map[*Client]bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients := make(map[*Client]bool, len(h.clients))
	for client := range h.clients {
		clients[client] = true
	}
	return clients
}

// remove deletes a client from the map and closes its Send channel.
// Must only be called from Run (single owner).
func (h *Hub) remove(client *Client) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[client]; !ok {
		return false
	}
	delete(h.clients, client)
	close(client.Send)
	return true
}

// trySend performs a non-blocking send to a client's queue.
// Must only be called from Run (single owner).
func (h *Hub) trySend(client *Client, data []byte) bool {
	select {
	case client.Send <- data:
		return true
	default:
		// Client is too slow to keep up — evict it.
		return false
	}
}

// Shutdown gracefully closes all client connections and stops the event loop.
func (h *Hub) Shutdown() {
	h.closeOnce.Do(func() {
		close(h.closed)
	})
}

// Len returns the number of currently connected clients.
func (h *Hub) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// SendToUser delivers a JSON event to all connections of a specific user.
func (h *Hub) SendToUser(userID string, event string, payload interface{}) {
	data, err := json.Marshal(map[string]interface{}{
		"event":   event,
		"payload": payload,
	})
	if err != nil {
		log.Printf("failed to marshal websocket message: %v", err)
		return
	}

	select {
	case h.direct <- directMsg{userID: userID, data: data}:
	case <-h.closed:
	}
}

// SendToAll delivers a JSON event to every connected client.
func (h *Hub) SendToAll(event string, payload interface{}) {
	data, err := json.Marshal(map[string]interface{}{
		"event":   event,
		"payload": payload,
	})
	if err != nil {
		log.Printf("failed to marshal websocket message: %v", err)
		return
	}

	select {
	case h.Broadcast <- data:
	case <-h.closed:
	}
}
