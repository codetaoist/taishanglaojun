package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client 表示一个WebSocket客户端连接
type Client struct {
	ID     string          `json:"id"`
	UserID string          `json:"user_id"`
	RoomID string          `json:"room_id"`
	Conn   *websocket.Conn `json:"-"`
	Send   chan []byte     `json:"-"`
	Hub    *Hub            `json:"-"`
}

// Message 表示WebSocket消息
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

// Hub 管理WebSocket连接和消息广播
type Hub struct {
	// 注册的客户端连接
	clients map[*Client]bool

	// 用户ID到客户端的映射
	userClients map[string]*Client

	// 从客户端接收的消息
	broadcast chan []byte

	// 注册新客户端
	register chan *Client

	// 注销客户端
	unregister chan *Client

	// 互斥锁保护并发访问
	mutex sync.RWMutex

	// WebSocket升级器
	upgrader websocket.Upgrader
}

// NewHub 创建新的Hub实例
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
				// 在生产环境中应该检查Origin
				return true
			},
		},
	}
}

// Run 启动Hub，处理客户端注册、注销和消息广播
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

// registerClient 注册新客户端
func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// 如果用户已经有连接，先关闭旧连接
	if existingClient, exists := h.userClients[client.UserID]; exists {
		h.unregisterClientUnsafe(existingClient)
	}

	h.clients[client] = true
	h.userClients[client.UserID] = client

	log.Printf("Client registered: %s (User: %s)", client.ID, client.UserID)

	// 发送连接成功消息
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

// unregisterClient 注销客户端
func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.unregisterClientUnsafe(client)
}

// unregisterClientUnsafe 注销客户端（不加锁版本）
func (h *Hub) unregisterClientUnsafe(client *Client) {
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		delete(h.userClients, client.UserID)
		close(client.Send)
		client.Conn.Close()
		log.Printf("Client unregistered: %s (User: %s)", client.ID, client.UserID)
	}
}

// broadcastMessage 广播消息给所有客户端
func (h *Hub) broadcastMessage(message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for client := range h.clients {
		select {
		case client.Send <- message:
		default:
			// 如果发送失败，关闭连接
			close(client.Send)
			delete(h.clients, client)
			delete(h.userClients, client.UserID)
		}
	}
}

// SendToUser 发送消息给特定用户
func (h *Hub) SendToUser(userID string, message []byte) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if client, exists := h.userClients[userID]; exists {
		select {
		case client.Send <- message:
			return true
		default:
			// 发送失败，清理连接
			close(client.Send)
			delete(h.clients, client)
			delete(h.userClients, client.UserID)
			return false
		}
	}
	return false
}

// GetConnectedUsers 获取当前连接的用户列表
func (h *Hub) GetConnectedUsers() []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	users := make([]string, 0, len(h.userClients))
	for userID := range h.userClients {
		users = append(users, userID)
	}
	return users
}

// GetClientCount 获取当前连接的客户端数量
func (h *Hub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// IsUserConnected 检查用户是否在线
func (h *Hub) IsUserConnected(userID string) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	_, exists := h.userClients[userID]
	return exists
}

// BroadcastToRoom 向指定房间广播消息
func (h *Hub) BroadcastToRoom(roomID string, message *Message) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	// 将消息序列化为JSON
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

// GetRoomUsers 获取房间内的用户列表
func (h *Hub) GetRoomUsers(roomID string) []uint {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	var users []uint
	for client := range h.clients {
		if client.RoomID == roomID {
			// 实际使用中可能需要根据具体情况调整
			users = append(users, uint(len(client.UserID))) // 临时实现，需要根据实际需求调整
		}
	}
	return users
}
