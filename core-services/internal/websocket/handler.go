package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WebSocketHandler WebSocket处理器
type WebSocketHandler struct {
	hub *Hub
}

// NewWebSocketHandler 创建新的WebSocket处理器
func NewWebSocketHandler(hub *Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
	}
}

// HandleWebSocket 处理WebSocket连接升级
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// 从JWT中获取用户ID (假设已经通过中间件验证)
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "UNAUTHORIZED",
			"message": "用户未认证",
		})
		return
	}

	// 验证用户ID格式
	if _, err := strconv.ParseUint(userIDStr, 10, 32); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_USER_ID",
			"message": "无效的用户ID",
		})
		return
	}

	// 升级HTTP连接为WebSocket连接
	conn, err := h.hub.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// 创建客户端
	client := &Client{
		ID:     uuid.New().String(),
		UserID: userIDStr,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    h.hub,
	}

	// 注册客户端到Hub
	h.hub.register <- client

	// 启动客户端的读写goroutine
	client.StartClient()
}

// HandleWebSocketPublic 处理公共WebSocket连接（不需要认证）
func (h *WebSocketHandler) HandleWebSocketPublic(c *gin.Context) {
	// 从查询参数获取用户ID（用于测试）
	userID := c.Query("user_id")
	if userID == "" {
		userID = "anonymous_" + uuid.New().String()[:8]
	}

	// 升级HTTP连接为WebSocket连接
	conn, err := h.hub.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// 创建客户端
	client := &Client{
		ID:     uuid.New().String(),
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    h.hub,
	}

	// 注册客户端到Hub
	h.hub.register <- client

	// 启动客户端的读写goroutine
	client.StartClient()
}

// GetConnectedUsers 获取当前连接的用户列表
func (h *WebSocketHandler) GetConnectedUsers(c *gin.Context) {
	users := h.hub.GetConnectedUsers()
	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
	})
}

// GetStats 获取WebSocket统计信息
func (h *WebSocketHandler) GetStats(c *gin.Context) {
	stats := gin.H{
		"connected_clients": h.hub.GetClientCount(),
		"connected_users":   len(h.hub.GetConnectedUsers()),
	}
	c.JSON(http.StatusOK, stats)
}

// SendMessage 通过HTTP API发送消息
func (h *WebSocketHandler) SendMessage(c *gin.Context) {
	var req struct {
		To      string      `json:"to"`
		Type    string      `json:"type"`
		Content string      `json:"content"`
		Data    interface{} `json:"data,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "请求参数无效: " + err.Error(),
		})
		return
	}

	// 创建消息
	msg := Message{
		Type:      req.Type,
		To:        req.To,
		Content:   req.Content,
		Data:      req.Data,
		Timestamp: time.Now().Unix(),
	}

	// 发送消息
	if req.To != "" {
		// 发送给特定用户
		if data, err := json.Marshal(msg); err == nil {
			success := h.hub.SendToUser(req.To, data)
			c.JSON(http.StatusOK, gin.H{
				"success": success,
				"message": "Message sent",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "MARSHAL_ERROR",
				"message": "消息序列化失败",
			})
		}
	} else {
		// 广播给所有用户
		if data, err := json.Marshal(msg); err == nil {
			h.hub.broadcast <- data
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Message broadcasted",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "MARSHAL_ERROR",
				"message": "消息序列化失败",
			})
		}
	}
}

// CheckUserOnline 检查用户是否在线
func (h *WebSocketHandler) CheckUserOnline(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "MISSING_USER_ID",
			"message": "用户ID不能为空",
		})
		return
	}

	online := h.hub.IsUserConnected(userID)
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"online":  online,
	})
}