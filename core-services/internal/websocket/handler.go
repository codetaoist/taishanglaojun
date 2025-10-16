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

// WebSocketHandler WebSocket
type WebSocketHandler struct {
	hub *Hub
}

// NewWebSocketHandler WebSocket
func NewWebSocketHandler(hub *Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
	}
}

// HandleWebSocket WebSocket
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// JWTID ()
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "UNAUTHORIZED",
			"message": "",
		})
		return
	}

	// ID
	if _, err := strconv.ParseUint(userIDStr, 10, 32); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_USER_ID",
			"message": "ID",
		})
		return
	}

	// HTTPWebSocket
	conn, err := h.hub.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// 
	client := &Client{
		ID:     uuid.New().String(),
		UserID: userIDStr,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    h.hub,
	}

	// Hub
	h.hub.register <- client

	// goroutine
	client.StartClient()
}

// HandleWebSocketPublic WebSocket
func (h *WebSocketHandler) HandleWebSocketPublic(c *gin.Context) {
	// ID
	userID := c.Query("user_id")
	if userID == "" {
		userID = "anonymous_" + uuid.New().String()[:8]
	}

	// HTTPWebSocket
	conn, err := h.hub.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// 
	client := &Client{
		ID:     uuid.New().String(),
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    h.hub,
	}

	// Hub
	h.hub.register <- client

	// goroutine
	client.StartClient()
}

// GetConnectedUsers 
func (h *WebSocketHandler) GetConnectedUsers(c *gin.Context) {
	users := h.hub.GetConnectedUsers()
	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
	})
}

// GetStats WebSocket
func (h *WebSocketHandler) GetStats(c *gin.Context) {
	stats := gin.H{
		"connected_clients": h.hub.GetClientCount(),
		"connected_users":   len(h.hub.GetConnectedUsers()),
	}
	c.JSON(http.StatusOK, stats)
}

// SendMessage HTTP API
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
			"message": ": " + err.Error(),
		})
		return
	}

	// 
	msg := Message{
		Type:      req.Type,
		To:        req.To,
		Content:   req.Content,
		Data:      req.Data,
		Timestamp: time.Now().Unix(),
	}

	// 
	if req.To != "" {
		// 
		if data, err := json.Marshal(msg); err == nil {
			success := h.hub.SendToUser(req.To, data)
			c.JSON(http.StatusOK, gin.H{
				"success": success,
				"message": "Message sent",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "MARSHAL_ERROR",
				"message": ": " + err.Error(),
			})
		}
	} else {
		// ?
		if data, err := json.Marshal(msg); err == nil {
			h.hub.broadcast <- data
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Message broadcasted",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "MARSHAL_ERROR",
				"message": ": " + err.Error(),
			})
		}
	}
}

// CheckUserOnline 
func (h *WebSocketHandler) CheckUserOnline(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "MISSING_USER_ID",
			"message": "ID",
		})
		return
	}

	online := h.hub.IsUserConnected(userID)
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"online":  online,
	})
}

