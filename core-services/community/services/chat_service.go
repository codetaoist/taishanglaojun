package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/community/models"
	"github.com/codetaoist/taishanglaojun/core-services/internal/websocket"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ChatService 聊天服务
type ChatService struct {
	db     *gorm.DB
	logger *zap.Logger
	hub    *websocket.Hub
}

// NewChatService 创建聊天服务
func NewChatService(db *gorm.DB, logger *zap.Logger, hub *websocket.Hub) *ChatService {
	return &ChatService{
		db:     db,
		logger: logger,
		hub:    hub,
	}
}

// CreateChatRoom 创建聊天室
func (s *ChatService) CreateChatRoom(ctx context.Context, userID uint, req *CreateChatRoomRequest) (*models.ChatRoom, error) {
	room := &models.ChatRoom{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		CreatorID:   userID,
		MaxMembers:  req.MaxMembers,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 开始事务
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建聊天室
	if err := tx.Create(room).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create chat room: %w", err)
	}

	// 创建者自动加入聊天室
	member := &models.ChatRoomMember{
		RoomID:   room.ID,
		UserID:   userID,
		Role:     "admin",
		JoinedAt: time.Now(),
	}

	if err := tx.Create(member).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to add creator to room: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("Chat room created", zap.Uint("room_id", room.ID), zap.Uint("creator_id", userID))
	return room, nil
}

// GetUserChatRooms 获取用户参与的聊天室列表
func (s *ChatService) GetUserChatRooms(ctx context.Context, userID uint, page, size int) ([]*models.ChatRoom, int64, error) {
	var rooms []*models.ChatRoom
	var total int64

	// 计算偏移量
	offset := (page - 1) * size

	// 查询用户参与的聊天室
	query := s.db.WithContext(ctx).
		Table("chat_rooms").
		Joins("JOIN chat_room_members ON chat_rooms.id = chat_room_members.room_id").
		Where("chat_room_members.user_id = ? AND chat_rooms.is_active = ?", userID, true)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count chat rooms: %w", err)
	}

	// 获取分页数据
	if err := query.
		Select("chat_rooms.*").
		Order("chat_rooms.updated_at DESC").
		Offset(offset).
		Limit(size).
		Find(&rooms).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get chat rooms: %w", err)
	}

	return rooms, total, nil
}

// JoinChatRoom 加入聊天室
func (s *ChatService) JoinChatRoom(ctx context.Context, userID, roomID uint) error {
	// 检查聊天室是否存在
	var room models.ChatRoom
	if err := s.db.WithContext(ctx).Where("id = ? AND is_active = ?", roomID, true).First(&room).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("聊天室不存在")
		}
		return fmt.Errorf("failed to find chat room: %w", err)
	}

	// 检查用户是否已经在聊天室中
	var existingMember models.ChatRoomMember
	err := s.db.WithContext(ctx).Where("room_id = ? AND user_id = ?", roomID, userID).First(&existingMember).Error
	if err == nil {
		return errors.New("用户已在聊天室中")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check existing membership: %w", err)
	}

	// 检查聊天室成员数量限制
	var memberCount int64
	if err := s.db.WithContext(ctx).Model(&models.ChatRoomMember{}).Where("room_id = ?", roomID).Count(&memberCount).Error; err != nil {
		return fmt.Errorf("failed to count room members: %w", err)
	}

	if int(memberCount) >= room.MaxMembers {
		return errors.New("聊天室已满")
	}

	// 添加成员
	member := &models.ChatRoomMember{
		RoomID:   roomID,
		UserID:   userID,
		Role:     "member",
		JoinedAt: time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(member).Error; err != nil {
		return fmt.Errorf("failed to join chat room: %w", err)
	}

	// 发送系统消息通知其他成员
	systemMessage := &models.ChatMessage{
		RoomID:    roomID,
		SenderID:  0, // 系统消息
		Content:   fmt.Sprintf("用户 %d 加入了聊天室", userID),
		Type:      "system",
		CreatedAt: time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(systemMessage).Error; err != nil {
		s.logger.Warn("Failed to create system message", zap.Error(err))
	}

	// 通过WebSocket广播加入消息
	if s.hub != nil {
		message := &websocket.Message{
			Type:      "user_joined",
			RoomID:    fmt.Sprintf("room_%d", roomID),
			UserID:    fmt.Sprintf("%d", userID),
			Content:   fmt.Sprintf("用户 %d 加入了聊天室", userID),
			Timestamp: time.Now().Unix(),
		}
		s.hub.BroadcastToRoom(fmt.Sprintf("room_%d", roomID), message)
	}

	s.logger.Info("User joined chat room", zap.Uint("user_id", userID), zap.Uint("room_id", roomID))
	return nil
}

// LeaveChatRoom 离开聊天室
func (s *ChatService) LeaveChatRoom(ctx context.Context, userID, roomID uint) error {
	// 检查用户是否在聊天室中
	var member models.ChatRoomMember
	if err := s.db.WithContext(ctx).Where("room_id = ? AND user_id = ?", roomID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不在聊天室中")
		}
		return fmt.Errorf("failed to find membership: %w", err)
	}

	// 删除成员记录
	if err := s.db.WithContext(ctx).Delete(&member).Error; err != nil {
		return fmt.Errorf("failed to leave chat room: %w", err)
	}

	// 发送系统消息通知其他成员
	systemMessage := &models.ChatMessage{
		RoomID:    roomID,
		SenderID:  0, // 系统消息
		Content:   fmt.Sprintf("用户 %d 离开了聊天室", userID),
		Type:      "system",
		CreatedAt: time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(systemMessage).Error; err != nil {
		s.logger.Warn("Failed to create system message", zap.Error(err))
	}

	// 通过WebSocket广播离开消息
	if s.hub != nil {
		message := &websocket.Message{
			Type:      "user_left",
			RoomID:    fmt.Sprintf("room_%d", roomID),
			UserID:    fmt.Sprintf("%d", userID),
			Content:   fmt.Sprintf("用户 %d 离开了聊天室", userID),
			Timestamp: time.Now().Unix(),
		}
		s.hub.BroadcastToRoom(fmt.Sprintf("room_%d", roomID), message)
	}

	s.logger.Info("User left chat room", zap.Uint("user_id", userID), zap.Uint("room_id", roomID))
	return nil
}

// GetChatMessages 获取聊天消息
func (s *ChatService) GetChatMessages(ctx context.Context, userID, roomID uint, page, size int) ([]*models.ChatMessage, int64, error) {
	// 检查用户是否有权限访问聊天室
	var member models.ChatRoomMember
	if err := s.db.WithContext(ctx).Where("room_id = ? AND user_id = ?", roomID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, errors.New("无权限访问该聊天室")
		}
		return nil, 0, fmt.Errorf("failed to check room access: %w", err)
	}

	var messages []*models.ChatMessage
	var total int64

	// 计算偏移量
	offset := (page - 1) * size

	// 查询消息
	query := s.db.WithContext(ctx).Where("room_id = ?", roomID)

	// 获取总数
	if err := query.Model(&models.ChatMessage{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count messages: %w", err)
	}

	// 获取分页数据（按时间倒序）
	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(size).
		Find(&messages).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get messages: %w", err)
	}

	return messages, total, nil
}

// SendMessage 发送消息
func (s *ChatService) SendMessage(ctx context.Context, userID, roomID uint, req *SendMessageRequest) (*models.ChatMessage, error) {
	// 检查用户是否有权限发送消息
	var member models.ChatRoomMember
	if err := s.db.WithContext(ctx).Where("room_id = ? AND user_id = ?", roomID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("无权限在该聊天室发送消息")
		}
		return nil, fmt.Errorf("failed to check room access: %w", err)
	}

	// 创建消息
	message := &models.ChatMessage{
		RoomID:    roomID,
		SenderID:  userID,
		Content:   req.Content,
		Type:      req.Type,
		ReplyToID: req.ReplyToID,
		CreatedAt: time.Now(),
	}

	if message.Type == "" {
		message.Type = "text"
	}

	// 保存消息到数据库
	if err := s.db.WithContext(ctx).Create(message).Error; err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// 通过WebSocket广播消息
	if s.hub != nil {
		wsMessage := &websocket.Message{
			Type:      "chat_message",
			RoomID:    fmt.Sprintf("room_%d", roomID),
			UserID:    fmt.Sprintf("%d", userID),
			Content:   message.Content,
			Timestamp: message.CreatedAt.Unix(),
			Data: map[string]interface{}{
				"message_id":   message.ID,
				"message_type": message.Type,
				"reply_to_id":  message.ReplyToID,
				"sender_id":    message.SenderID,
				"room_id":      message.RoomID,
			},
		}
		s.hub.BroadcastToRoom(fmt.Sprintf("room_%d", roomID), wsMessage)
	}

	s.logger.Info("Message sent and persisted", 
		zap.Uint("message_id", message.ID), 
		zap.Uint("user_id", userID), 
		zap.Uint("room_id", roomID),
		zap.String("content", message.Content),
		zap.String("type", message.Type))
	
	return message, nil
}

// GetOnlineUsers 获取在线用户列表
func (s *ChatService) GetOnlineUsers(ctx context.Context, roomID uint) ([]uint, error) {
	if s.hub == nil {
		return []uint{}, nil
	}

	roomKey := fmt.Sprintf("room_%d", roomID)
	users := s.hub.GetRoomUsers(roomKey)
	
	return users, nil
}

// 请求结构体
type CreateChatRoomRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	MaxMembers  int    `json:"max_members"`
}

type SendMessageRequest struct {
	Content   string `json:"content"`
	Type      string `json:"type"`
	ReplyToID *uint  `json:"reply_to_id"`
}