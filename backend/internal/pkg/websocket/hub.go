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
				log.Printf("read error: %v", err)
			}
			break
		}
		// Echo back for now (basic WebSocket)
		if n > 0 {
			_ = buf[:n]
		}
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
				c.Conn.Write([]byte("close"))
				return
			}
			// Write raw message for basic WebSocket
			c.Conn.Write(msg)
		case <-ticker.C:
			if err := c.Conn.Ping(context.Background()); err != nil {
				return
			}
		}
	}
}

type Hub struct {
	Clients      map[*Client]bool
	Register     chan *Client
	Unregister   chan *Client
	Broadcast    chan []byte
	mu           sync.RWMutex
	PingInterval time.Duration
	Logger       *config.Config
}

func NewHub(cfg *config.Config) *Hub {
	pingInterval := time.Duration(cfg.WSPingInterval) * time.Second
	return &Hub{
		Clients:      make(map[*Client]bool),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		Broadcast:    make(chan []byte),
		PingInterval: pingInterval,
		Logger:       cfg,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()
			h.Logger.Info("client registered", zap.String("user_id", client.UserID))
		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			h.Logger.Info("client unregistered", zap.String("user_id", client.UserID))
		case msg := <-h.Broadcast:
			h.mu.RLock()
			for client := range h.Clients {
				select {
				case client.Send <- msg:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) SendToUser(userID string, event string, payload interface{}) {
	data, err := json.Marshal(map[string]interface{}{
		"event":   event,
		"payload": payload,
	})
	if err != nil {
		log.Printf("failed to marshal websocket message: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.Clients {
		if client.UserID == userID {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.Clients, client)
			}
		}
	}
}

func (h *Hub) SendToAll(event string, payload interface{}) {
	data, err := json.Marshal(map[string]interface{}{
		"event":   event,
		"payload": payload,
	})
	if err != nil {
		log.Printf("failed to marshal websocket message: %v", err)
		return
	}

	h.Broadcast <- data
}
