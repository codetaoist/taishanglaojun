package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// 
	writeWait = 10 * time.Second

	// Pong
	pongWait = 60 * time.Second

	// PingpongWait
	pingPeriod = (pongWait * 9) / 10

	// 
	maxMessageSize = 512
)

// readPump WebSocket
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// 
		var msg Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		// ID
		msg.From = c.UserID
		msg.Timestamp = time.Now().Unix()

		// 
		c.handleMessage(&msg)
	}
}

// writePump WebSocket
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 
func (c *Client) handleMessage(msg *Message) {
	switch msg.Type {
	case "chat":
		c.handleChatMessage(msg)
	case "typing":
		c.handleTypingMessage(msg)
	case "ping":
		c.handlePingMessage(msg)
	case "join_room":
		c.handleJoinRoomMessage(msg)
	case "leave_room":
		c.handleLeaveRoomMessage(msg)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// handleChatMessage 
func (c *Client) handleChatMessage(msg *Message) {
	// 
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().Unix()
	}

	// ID
	msg.UserID = c.UserID

	// ID?
	if msg.RoomID != "" {
		if _, err := json.Marshal(msg); err == nil {
			c.Hub.BroadcastToRoom(msg.RoomID, msg)
		}
	} else if msg.To != "" {
		// 
		if data, err := json.Marshal(msg); err == nil {
			if !c.Hub.SendToUser(msg.To, data) {
				// 
				errorMsg := Message{
					Type:      "error",
					Content:   "User not online",
					Timestamp: time.Now().Unix(),
				}
				if errorData, err := json.Marshal(errorMsg); err == nil {
					select {
					case c.Send <- errorData:
					default:
					}
				}
			}
		}
	} else {
		// ?
		if data, err := json.Marshal(msg); err == nil {
			c.Hub.broadcast <- data
		}
	}
}

// handleTypingMessage 
func (c *Client) handleTypingMessage(msg *Message) {
	// ?
	if msg.To != "" {
		if data, err := json.Marshal(msg); err == nil {
			c.Hub.SendToUser(msg.To, data)
		}
	} else {
		if data, err := json.Marshal(msg); err == nil {
			c.Hub.broadcast <- data
		}
	}
}

// handlePingMessage ping
func (c *Client) handlePingMessage(msg *Message) {
	pongMsg := Message{
		Type:      "pong",
		Content:   "pong",
		Timestamp: time.Now().Unix(),
	}

	if data, err := json.Marshal(pongMsg); err == nil {
		select {
		case c.Send <- data:
		default:
		}
	}
}

// handleJoinRoomMessage 处理加入房间消息
func (c *Client) handleJoinRoomMessage(msg *Message) {
	// TODO: 实现加入房间逻辑
	log.Printf("User %s wants to join room: %s", c.UserID, msg.Content)
}

// handleLeaveRoomMessage 处理离开房间消息
func (c *Client) handleLeaveRoomMessage(msg *Message) {
	// TODO: 实现离开房间逻辑
	log.Printf("User %s wants to leave room: %s", c.UserID, msg.Content)
}

// StartClient goroutine
func (c *Client) StartClient() {
	go c.writePump()
	go c.readPump()
}

