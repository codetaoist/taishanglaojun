package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/community/models"
	"github.com/codetaoist/taishanglaojun/core-services/internal/websocket"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ChatService 
type ChatService struct {
	db     *gorm.DB
	logger *zap.Logger
	hub    *websocket.Hub
}

// NewChatService 
func NewChatService(db *gorm.DB, logger *zap.Logger, hub *websocket.Hub) *ChatService {
	return &ChatService{
		db:     db,
		logger: logger,
		hub:    hub,
	}
}

// CreateChatRoom 
func (s *ChatService) CreateChatRoom(ctx context.Context, userID uuid.UUID, req *CreateChatRoomRequest) (*models.ChatRoom, error) {
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

	// 
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 
	if err := tx.Create(room).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create chat room: %w", err)
	}

	// 
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

	// 
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("Chat room created", zap.Uint("room_id", room.ID), zap.String("creator_id", userID.String()))
	return room, nil
}

// GetUserChatRooms 
func (s *ChatService) GetUserChatRooms(ctx context.Context, userID uuid.UUID, page, size int) ([]*models.ChatRoom, int64, error) {
	var rooms []*models.ChatRoom
	var total int64

	// 
	offset := (page - 1) * size

	// 
	query := s.db.WithContext(ctx).
		Table("chat_rooms").
		Joins("JOIN chat_room_members ON chat_rooms.id = chat_room_members.room_id").
		Where("chat_room_members.user_id = ? AND chat_rooms.is_active = ?", userID, true)

	// 
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count chat rooms: %w", err)
	}

	// 
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

// JoinChatRoom 
func (s *ChatService) JoinChatRoom(ctx context.Context, userID uuid.UUID, roomID uint) error {
	// 
	var room models.ChatRoom
	if err := s.db.WithContext(ctx).Where("id = ? AND is_active = ?", roomID, true).First(&room).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("")
		}
		return fmt.Errorf("failed to find chat room: %w", err)
	}

	// 
	var existingMember models.ChatRoomMember
	err := s.db.WithContext(ctx).Where("room_id = ? AND user_id = ?", roomID, userID).First(&existingMember).Error
	if err == nil {
		return errors.New("")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check existing membership: %w", err)
	}

	// 
	var memberCount int64
	if err := s.db.WithContext(ctx).Model(&models.ChatRoomMember{}).Where("room_id = ?", roomID).Count(&memberCount).Error; err != nil {
		return fmt.Errorf("failed to count room members: %w", err)
	}

	if int(memberCount) >= room.MaxMembers {
		return errors.New("")
	}

	// 
	member := &models.ChatRoomMember{
		RoomID:   roomID,
		UserID:   userID,
		Role:     "member",
		JoinedAt: time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(member).Error; err != nil {
		return fmt.Errorf("failed to join chat room: %w", err)
	}

	// 
	systemMessage := &models.ChatMessage{
		RoomID:    roomID,
		SenderID:  uuid.Nil, // 
		Content:   fmt.Sprintf("用户 %s 加入了聊天室", userID.String()),
		Type:      "system",
		CreatedAt: time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(systemMessage).Error; err != nil {
		s.logger.Warn("Failed to create system message", zap.Error(err))
	}

	// WebSocket㲥
	if s.hub != nil {
		message := &websocket.Message{
			Type:      "user_joined",
			RoomID:    fmt.Sprintf("room_%d", roomID),
			UserID:    userID.String(),
			Content:   fmt.Sprintf("用户 %s 加入了聊天室", userID.String()),
			Timestamp: time.Now().Unix(),
		}
		s.hub.BroadcastToRoom(fmt.Sprintf("room_%d", roomID), message)
	}

	s.logger.Info("User joined chat room", zap.String("user_id", userID.String()), zap.Uint("room_id", roomID))
	return nil
}

// LeaveChatRoom 
func (s *ChatService) LeaveChatRoom(ctx context.Context, userID uuid.UUID, roomID uint) error {
	// 
	var member models.ChatRoomMember
	if err := s.db.WithContext(ctx).Where("room_id = ? AND user_id = ?", roomID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("")
		}
		return fmt.Errorf("failed to find membership: %w", err)
	}

	// 
	if err := s.db.WithContext(ctx).Delete(&member).Error; err != nil {
		return fmt.Errorf("failed to leave chat room: %w", err)
	}

	// 
	systemMessage := &models.ChatMessage{
		RoomID:    roomID,
		SenderID:  uuid.Nil, // 
		Content:   fmt.Sprintf("用户 %s 离开了聊天室", userID.String()),
		Type:      "system",
		CreatedAt: time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(systemMessage).Error; err != nil {
		s.logger.Warn("Failed to create system message", zap.Error(err))
	}

	// WebSocket㲥
	if s.hub != nil {
		message := &websocket.Message{
			Type:      "user_left",
			RoomID:    fmt.Sprintf("room_%d", roomID),
			UserID:    userID.String(),
			Content:   fmt.Sprintf("用户 %s 离开了聊天室", userID.String()),
			Timestamp: time.Now().Unix(),
		}
		s.hub.BroadcastToRoom(fmt.Sprintf("room_%d", roomID), message)
	}

	s.logger.Info("User left chat room", zap.String("user_id", userID.String()), zap.Uint("room_id", roomID))
	return nil
}

// GetChatMessages 
func (s *ChatService) GetChatMessages(ctx context.Context, userID uuid.UUID, roomID uint, page, size int) ([]*models.ChatMessage, int64, error) {
	// 
	var member models.ChatRoomMember
	if err := s.db.WithContext(ctx).Where("room_id = ? AND user_id = ?", roomID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, errors.New("")
		}
		return nil, 0, fmt.Errorf("failed to check room access: %w", err)
	}

	var messages []*models.ChatMessage
	var total int64

	// 
	offset := (page - 1) * size

	// 
	query := s.db.WithContext(ctx).Where("room_id = ?", roomID)

	// 
	if err := query.Model(&models.ChatMessage{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count messages: %w", err)
	}

	// 䵹
	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(size).
		Find(&messages).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get messages: %w", err)
	}

	return messages, total, nil
}

// SendMessage 
func (s *ChatService) SendMessage(ctx context.Context, userID uuid.UUID, roomID uint, req *SendMessageRequest) (*models.ChatMessage, error) {
	// 
	var member models.ChatRoomMember
	if err := s.db.WithContext(ctx).Where("room_id = ? AND user_id = ?", roomID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("")
		}
		return nil, fmt.Errorf("failed to check room access: %w", err)
	}

	// 
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

	// 
	if err := s.db.WithContext(ctx).Create(message).Error; err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// WebSocket㲥
	if s.hub != nil {
		wsMessage := &websocket.Message{
			Type:      "chat_message",
			RoomID:    fmt.Sprintf("room_%d", roomID),
			UserID:    userID.String(),
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
		zap.String("user_id", userID.String()),
		zap.Uint("room_id", roomID),
		zap.String("content", message.Content),
		zap.String("type", message.Type))

	return message, nil
}

// GetOnlineUsers 
func (s *ChatService) GetOnlineUsers(ctx context.Context, roomID uint) ([]uint, error) {
	if s.hub == nil {
		return []uint{}, nil
	}

	roomKey := fmt.Sprintf("room_%d", roomID)
	users := s.hub.GetRoomUsers(roomKey)

	return users, nil
}

// 
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

