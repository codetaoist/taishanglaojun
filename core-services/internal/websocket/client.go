package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// 写入等待时间
	writeWait = 10 * time.Second

	// Pong等待时间
	pongWait = 60 * time.Second

	// Ping周期，必须小于pongWait
	pingPeriod = (pongWait * 9) / 10

	// 最大消息大小
	maxMessageSize = 512
)

// readPump 处理从WebSocket连接读取消息
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

		// 解析消息
		var msg Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		// 设置消息发送者
		msg.From = c.UserID
		msg.Timestamp = time.Now().Unix()

		// 处理不同类型的消息
		c.handleMessage(&msg)
	}
}

// writePump 处理向WebSocket连接写入消息
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
				// Hub关闭了通道
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 将队列中的其他消息一起发送
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

// handleMessage 处理接收到的消息
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

// handleChatMessage 处理聊天消息
func (c *Client) handleChatMessage(msg *Message) {
	// 设置消息时间戳
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().Unix()
	}

	// 设置发送者ID
	msg.UserID = c.UserID

	// 如果指定了房间ID，广播到房间
	if msg.RoomID != "" {
		if _, err := json.Marshal(msg); err == nil {
			c.Hub.BroadcastToRoom(msg.RoomID, msg)
		}
	} else if msg.To != "" {
		// 如果指定了接收者，发送给特定用户
		if data, err := json.Marshal(msg); err == nil {
			if !c.Hub.SendToUser(msg.To, data) {
				// 发送失败，可能用户不在线
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
		// 广播给所有用户
		if data, err := json.Marshal(msg); err == nil {
			c.Hub.broadcast <- data
		}
	}
}

// handleTypingMessage 处理打字状态消息
func (c *Client) handleTypingMessage(msg *Message) {
	// 转发打字状态给指定用户或广播
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

// handlePingMessage 处理ping消息
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
	// TODO: 实现房间功能
	log.Printf("User %s wants to join room: %s", c.UserID, msg.Content)
}

// handleLeaveRoomMessage 处理离开房间消息
func (c *Client) handleLeaveRoomMessage(msg *Message) {
	// TODO: 实现房间功能
	log.Printf("User %s wants to leave room: %s", c.UserID, msg.Content)
}

// StartClient 启动客户端的读写goroutine
func (c *Client) StartClient() {
	go c.writePump()
	go c.readPump()
}