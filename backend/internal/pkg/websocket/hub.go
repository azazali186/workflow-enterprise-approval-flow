package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/coder/websocket"
)

type Client struct {
	ID     string
	UserID string
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *Hub
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close(websocket.StatusNormalClosure, "")
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetPingHandler(func(appData string) error {
		c.Conn.Pong(appData)
		return nil
	})

	for {
		_, _, err := c.Conn.Read(context.Background())
		if err != nil {
			break
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(c.Hub.PingInterval)
	defer func() {
		ticker.Stop()
		c.Conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			if !ok {
				c.Conn.Write(context.Background(), websocket.MessageText, []byte("close"))
				return
			}
			c.Conn.Write(context.Background(), websocket.MessageText, msg)
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
			h.Logger.Info("client registered", "user_id", client.UserID)
		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			h.Logger.Info("client unregistered", "user_id", client.UserID)
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
