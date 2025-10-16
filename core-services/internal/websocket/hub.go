package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client WebSocket
type Client struct {
	ID     string          `json:"id"`
	UserID string          `json:"user_id"`
	RoomID string          `json:"room_id"`
	Conn   *websocket.Conn `json:"-"`
	Send   chan []byte     `json:"-"`
	Hub    *Hub            `json:"-"`
}

// Message WebSocket
type Message struct {
	Type      string      `json:"type"`
	From      string      `json:"from"`
	To        string      `json:"to,omitempty"`
	Content   string      `json:"content"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
	UserID    string      `json:"user_id,omitempty"`
	RoomID    string      `json:"room_id,omitempty"`
}

// Hub WebSocket?
type Hub struct {
	// 
	clients map[*Client]bool

	// ID
	userClients map[string]*Client

	// ?
	broadcast chan []byte

	// 
	register chan *Client

	// 
	unregister chan *Client

	// 
	mutex sync.RWMutex

	// WebSocket
	upgrader websocket.Upgrader
}

// NewHub Hub
func NewHub() *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		userClients: make(map[string]*Client),
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Origin
				return true
			},
		},
	}
}

// Run Hub?
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient 
func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// 
	if existingClient, exists := h.userClients[client.UserID]; exists {
		h.unregisterClientUnsafe(existingClient)
	}

	h.clients[client] = true
	h.userClients[client.UserID] = client

	log.Printf("Client registered: %s (User: %s)", client.ID, client.UserID)

	// 
	welcomeMsg := Message{
		Type:      "connection",
		Content:   "Connected successfully",
		Timestamp: time.Now().Unix(),
	}

	if data, err := json.Marshal(welcomeMsg); err == nil {
		select {
		case client.Send <- data:
		default:
			close(client.Send)
			delete(h.clients, client)
			delete(h.userClients, client.UserID)
		}
	}
}

// unregisterClient 
func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.unregisterClientUnsafe(client)
}

// unregisterClientUnsafe 取消注册客户端（非线程安全）
func (h *Hub) unregisterClientUnsafe(client *Client) {
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		delete(h.userClients, client.UserID)
		close(client.Send)
		client.Conn.Close()
		log.Printf("Client unregistered: %s (User: %s)", client.ID, client.UserID)
	}
}

// broadcastMessage ?
func (h *Hub) broadcastMessage(message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for client := range h.clients {
		select {
		case client.Send <- message:
		default:
			// 
			close(client.Send)
			delete(h.clients, client)
			delete(h.userClients, client.UserID)
		}
	}
}

// SendToUser 
func (h *Hub) SendToUser(userID string, message []byte) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if client, exists := h.userClients[userID]; exists {
		select {
		case client.Send <- message:
			return true
		default:
			// 
			close(client.Send)
			delete(h.clients, client)
			delete(h.userClients, client.UserID)
			return false
		}
	}
	return false
}

// GetConnectedUsers 
func (h *Hub) GetConnectedUsers() []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	users := make([]string, 0, len(h.userClients))
	for userID := range h.userClients {
		users = append(users, userID)
	}
	return users
}

// GetClientCount 
func (h *Hub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// IsUserConnected 
func (h *Hub) IsUserConnected(userID string) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	_, exists := h.userClients[userID]
	return exists
}

// BroadcastToRoom ?
func (h *Hub) BroadcastToRoom(roomID string, message *Message) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	// JSON
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	for client := range h.clients {
		if client.RoomID == roomID {
			select {
			case client.Send <- messageBytes:
			default:
				close(client.Send)
				delete(h.clients, client)
			}
		}
	}
}

// GetRoomUsers 
func (h *Hub) GetRoomUsers(roomID string) []uint {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	var users []uint
	for client := range h.clients {
		if client.RoomID == roomID {
			// 
			users = append(users, uint(len(client.UserID))) // 
		}
	}
	return users
}

