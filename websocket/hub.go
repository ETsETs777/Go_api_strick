package websocket

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

type Client struct {
	ID   string
	Conn *websocket.Conn
	Send chan Message
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			fmt.Printf("WebSocket: Клиент подключен (ID: %s). Всего клиентов: %d\n", 
				client.ID, len(h.clients))
			
			h.SendToClient(client, Message{
				Type: "welcome",
				Data: map[string]interface{}{
					"message": "Добро пожаловать в Go Showcase WebSocket!",
					"id":      client.ID,
				},
				Timestamp: time.Now(),
			})
			
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.mu.Lock()
				delete(h.clients, client)
				close(client.Send)
				h.mu.Unlock()
				fmt.Printf("WebSocket: Клиент отключен (ID: %s). Всего клиентов: %d\n", 
					client.ID, len(h.clients))
			}
			
		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
			
		case <-ticker.C:
			h.BroadcastMessage(Message{
				Type: "heartbeat",
				Data: map[string]interface{}{
					"active_clients": len(h.clients),
					"server_time":    time.Now().Format(time.RFC3339),
				},
				Timestamp: time.Now(),
			})
		}
	}
}

func (h *Hub) BroadcastMessage(msg Message) {
	h.broadcast <- msg
}

func (h *Hub) SendToClient(client *Client, msg Message) {
	select {
	case client.Send <- msg:
	default:
		close(client.Send)
		h.mu.Lock()
		delete(h.clients, client)
		h.mu.Unlock()
	}
}

func (h *Hub) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	return map[string]interface{}{
		"total_clients": len(h.clients),
		"timestamp":     time.Now(),
	}
}

func (h *Hub) Shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	shutdownMsg := Message{
		Type: "shutdown",
		Data: map[string]interface{}{
			"message": "Server is shutting down gracefully",
		},
		Timestamp: time.Now(),
	}
	
	for client := range h.clients {
		client.Send <- shutdownMsg
		close(client.Send)
		client.Conn.Close()
	}
	
	h.clients = make(map[*Client]bool)
	fmt.Printf("WebSocket: All clients disconnected\n")
}

func (c *Client) ReadPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.Conn.Close()
	}()
	
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	for {
		var msg Message
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("WebSocket error: %v\n", err)
			}
			break
		}
		
		msg.Timestamp = time.Now()
		
		response := Message{
			Type: "echo",
			Data: map[string]interface{}{
				"received": msg,
				"from":     c.ID,
			},
			Timestamp: time.Now(),
		}
		
		hub.BroadcastMessage(response)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			data, err := json.Marshal(message)
			if err != nil {
				return
			}
			
			if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
			
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

