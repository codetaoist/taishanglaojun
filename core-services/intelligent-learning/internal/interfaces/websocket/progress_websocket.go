package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	analyticsServices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/analytics"
)

// ProgressWebSocketHandler WebSocket?
type ProgressWebSocketHandler struct {
	progressService analyticsServices.ProgressTrackingService
	upgrader        websocket.Upgrader
	clients         map[uuid.UUID]*Client
	clientsMutex    sync.RWMutex
	broadcast       chan *ProgressBroadcast
	register        chan *Client
	unregister      chan *Client
}

// Client WebSocket
type Client struct {
	ID            uuid.UUID
	LearnerID     uuid.UUID
	Conn          *websocket.Conn
	Send          chan []byte
	Handler       *ProgressWebSocketHandler
	LastSeen      time.Time
	Subscriptions map[string]bool // ?
}

// ProgressBroadcast 㲥
type ProgressBroadcast struct {
	Type      string      `json:"type"`
	LearnerID uuid.UUID   `json:"learner_id"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// WebSocketMessage WebSocket
type WebSocketMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
	ID   string          `json:"id,omitempty"`
}

// ProgressUpdateMessage 
type ProgressUpdateMessage struct {
	ContentID       uuid.UUID              `json:"content_id"`
	Progress        float64                `json:"progress"`
	TimeSpent       int                    `json:"time_spent"`
	LastPosition    int                    `json:"last_position"`
	InteractionData map[string]interface{} `json:"interaction_data"`
}

// SubscriptionMessage 
type SubscriptionMessage struct {
	Events []string `json:"events"`
}

// NewProgressWebSocketHandler WebSocket
// ?
func NewProgressWebSocketHandler(progressService analyticsServices.ProgressTrackingService) *ProgressWebSocketHandler {
	handler := &ProgressWebSocketHandler{
		progressService: progressService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Origin
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		clients:    make(map[uuid.UUID]*Client),
		broadcast:  make(chan *ProgressBroadcast, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}

	// 
	go handler.run()

	return handler
}

// HandleWebSocket WebSocket
func (h *ProgressWebSocketHandler) HandleWebSocket(c *gin.Context) {
	// ID
	learnerIDStr := c.Query("learner_id")
	if learnerIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "learner_id is required"})
		return
	}

	learnerID, err := uuid.Parse(learnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid learner_id"})
		return
	}

	// WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// 
	client := &Client{
		ID:            uuid.New(),
		LearnerID:     learnerID,
		Conn:          conn,
		Send:          make(chan []byte, 256),
		Handler:       h,
		LastSeen:      time.Now(),
		Subscriptions: make(map[string]bool),
	}

	// 
	h.register <- client

	// ?
	go client.writePump()
	go client.readPump()
}

// run 
func (h *ProgressWebSocketHandler) run() {
	ticker := time.NewTicker(54 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.clientsMutex.Lock()
			h.clients[client.ID] = client
			h.clientsMutex.Unlock()
			log.Printf("Client %s connected for learner %s", client.ID, client.LearnerID)

			// ?
			welcome := map[string]interface{}{
				"type":      "welcome",
				"message":   "Connected to progress tracking",
				"client_id": client.ID,
			}
			client.sendMessage(welcome)

		case client := <-h.unregister:
			h.clientsMutex.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
			}
			h.clientsMutex.Unlock()
			log.Printf("Client %s disconnected", client.ID)

		case broadcast := <-h.broadcast:
			h.clientsMutex.RLock()
			for _, client := range h.clients {
				if client.LearnerID == broadcast.LearnerID {
					select {
					case client.Send <- h.marshalBroadcast(broadcast):
					default:
						close(client.Send)
						delete(h.clients, client.ID)
					}
				}
			}
			h.clientsMutex.RUnlock()

		case <-ticker.C:
			// 
			h.cleanupClients()
		}
	}
}

// readPump ?
// ?
func (c *Client) readPump() {
	defer func() {
		c.Handler.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		c.LastSeen = time.Now()
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		c.LastSeen = time.Now()
		c.handleMessage(message)
	}
}

// writePump 
func (c *Client) writePump() {
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

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// ?
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
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

// handleMessage ?
// ?
func (c *Client) handleMessage(message []byte) {
	var msg WebSocketMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Invalid message format: %v", err)
		return
	}

	switch msg.Type {
	case "progress_update":
		c.handleProgressUpdate(msg.Data)
	case "subscribe":
		c.handleSubscription(msg.Data)
	case "ping":
		c.sendMessage(map[string]interface{}{
			"type":      "pong",
			"timestamp": time.Now(),
		})
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// handleProgressUpdate 
func (c *Client) handleProgressUpdate(data json.RawMessage) {
	var updateMsg ProgressUpdateMessage
	if err := json.Unmarshal(data, &updateMsg); err != nil {
		log.Printf("Invalid progress update format: %v", err)
		return
	}

	// 
	req := &analyticsServices.ProgressUpdateRequest{
		LearnerID:       c.LearnerID,
		ContentID:       updateMsg.ContentID,
		Progress:        updateMsg.Progress,
		TimeSpent:       int64(updateMsg.TimeSpent),
		LastPosition:    strconv.Itoa(updateMsg.LastPosition),
		InteractionData: updateMsg.InteractionData,
	}

	// 
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := c.Handler.progressService.UpdateProgress(ctx, req)
	if err != nil {
		c.sendError("progress_update_failed", err.Error())
		return
	}

	// ?
	c.sendMessage(map[string]interface{}{
		"type": "progress_updated",
		"data": response,
	})

	// 㲥
	c.Handler.broadcast <- &ProgressBroadcast{
		Type:      "progress_update",
		LearnerID: c.LearnerID,
		Data:      response,
		Timestamp: time.Now(),
	}
}

// handleSubscription 
func (c *Client) handleSubscription(data json.RawMessage) {
	var subMsg SubscriptionMessage
	if err := json.Unmarshal(data, &subMsg); err != nil {
		log.Printf("Invalid subscription format: %v", err)
		return
	}

	// 
	for _, event := range subMsg.Events {
		c.Subscriptions[event] = true
	}

	c.sendMessage(map[string]interface{}{
		"type":   "subscription_updated",
		"events": subMsg.Events,
	})
}

// sendMessage ?
//  JSON ?
func (c *Client) sendMessage(data interface{}) {
	message, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	select {
	case c.Send <- message:
	default:
		close(c.Send)
	}
}

// sendError ?
//  JSON ?
func (c *Client) sendError(errorType, message string) {
	c.sendMessage(map[string]interface{}{
		"type":       "error",
		"error_type": errorType,
		"message":    message,
		"timestamp":  time.Now(),
	})
}

// marshalBroadcast 㲥?
// 㲥 JSON ?
func (h *ProgressWebSocketHandler) marshalBroadcast(broadcast *ProgressBroadcast) []byte {
	data, err := json.Marshal(broadcast)
	if err != nil {
		log.Printf("Failed to marshal broadcast: %v", err)
		return nil
	}
	return data
}

// cleanupClients 
func (h *ProgressWebSocketHandler) cleanupClients() {
	h.clientsMutex.Lock()
	defer h.clientsMutex.Unlock()

	timeout := time.Now().Add(-5 * time.Minute)
	for id, client := range h.clients {
		if client.LastSeen.Before(timeout) {
			close(client.Send)
			delete(h.clients, id)
			log.Printf("Cleaned up timeout client %s", id)
		}
	}
}

// BroadcastProgressUpdate 㲥
func (h *ProgressWebSocketHandler) BroadcastProgressUpdate(learnerID uuid.UUID, data interface{}) {
	select {
	case h.broadcast <- &ProgressBroadcast{
		Type:      "progress_update",
		LearnerID: learnerID,
		Data:      data,
		Timestamp: time.Now(),
	}:
	default:
		log.Printf("Broadcast channel full, dropping message for learner %s", learnerID)
	}
}

// BroadcastAchievement 㲥
func (h *ProgressWebSocketHandler) BroadcastAchievement(learnerID uuid.UUID, achievement interface{}) {
	select {
	case h.broadcast <- &ProgressBroadcast{
		Type:      "achievement_unlocked",
		LearnerID: learnerID,
		Data:      achievement,
		Timestamp: time.Now(),
	}:
	default:
		log.Printf("Broadcast channel full, dropping achievement for learner %s", learnerID)
	}
}

// BroadcastRecommendation 㲥
func (h *ProgressWebSocketHandler) BroadcastRecommendation(learnerID uuid.UUID, recommendations interface{}) {
	select {
	case h.broadcast <- &ProgressBroadcast{
		Type:      "recommendations_updated",
		LearnerID: learnerID,
		Data:      recommendations,
		Timestamp: time.Now(),
	}:
	default:
		log.Printf("Broadcast channel full, dropping recommendations for learner %s", learnerID)
	}
}

// GetConnectedClients 
func (h *ProgressWebSocketHandler) GetConnectedClients() int {
	h.clientsMutex.RLock()
	defer h.clientsMutex.RUnlock()
	return len(h.clients)
}

// GetClientsByLearner ?
// ?
func (h *ProgressWebSocketHandler) GetClientsByLearner(learnerID uuid.UUID) int {
	h.clientsMutex.RLock()
	defer h.clientsMutex.RUnlock()

	count := 0
	for _, client := range h.clients {
		if client.LearnerID == learnerID {
			count++
		}
	}
	return count
}

