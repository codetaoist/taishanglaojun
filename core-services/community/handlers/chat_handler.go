package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/core-services/community/models"
	"github.com/codetaoist/taishanglaojun/core-services/community/services"
	"go.uber.org/zap"
)

// ChatHandler 聊天处理�?
type ChatHandler struct {
	chatService *services.ChatService
	logger      *zap.Logger
}

// NewChatHandler 创建聊天处理�?
func NewChatHandler(chatService *services.ChatService, logger *zap.Logger) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		logger:      logger,
	}
}

// CreateChatRoom 创建聊天�?
// @Summary 创建聊天�?
// @Description 创建新的聊天�?
// @Tags 聊天
// @Accept json
// @Produce json
// @Param request body CreateChatRoomRequest true "创建聊天室请�?
// @Success 200 {object} ChatRoomResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/chat/rooms [post]
func (h *ChatHandler) CreateChatRoom(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		return
	}

	var req services.CreateChatRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	room, err := h.chatService.CreateChatRoom(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Failed to create chat room", zap.Error(err), zap.Uint("user_id", userID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CREATE_ROOM_ERROR",
			Message: "创建聊天室失�?,
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ChatRoomResponse{
		Code:    200,
		Message: "聊天室创建成�?,
		Data:    room,
	})
}

// GetChatRooms 获取聊天室列�?
// @Summary 获取聊天室列�?
// @Description 获取用户参与的聊天室列表
// @Tags 聊天
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(20)
// @Success 200 {object} ChatRoomListResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/chat/rooms [get]
func (h *ChatHandler) GetChatRooms(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}

	rooms, total, err := h.chatService.GetUserChatRooms(c.Request.Context(), userID, page, size)
	if err != nil {
		h.logger.Error("Failed to get chat rooms", zap.Error(err), zap.Uint("user_id", userID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_ROOMS_ERROR",
			Message: "获取聊天室列表失�?,
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ChatRoomListResponse{
		Code:    200,
		Message: "获取成功",
		Data:    rooms,
		Total:   total,
		Page:    page,
		Size:    size,
	})
}

// JoinChatRoom 加入聊天�?
// @Summary 加入聊天�?
// @Description 用户加入指定聊天�?
// @Tags 聊天
// @Accept json
// @Produce json
// @Param room_id path int true "聊天室ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/chat/rooms/{room_id}/join [post]
func (h *ChatHandler) JoinChatRoom(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		return
	}

	roomIDStr := c.Param("room_id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ROOM_ID",
			Message: "聊天室ID无效",
		})
		return
	}

	err = h.chatService.JoinChatRoom(c.Request.Context(), userID, uint(roomID))
	if err != nil {
		h.logger.Error("Failed to join chat room", zap.Error(err), zap.Uint("user_id", userID), zap.Uint("room_id", uint(roomID)))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "JOIN_ROOM_ERROR",
			Message: "加入聊天室失�?,
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    200,
		Message: "成功加入聊天�?,
	})
}

// LeaveChatRoom 离开聊天�?
// @Summary 离开聊天�?
// @Description 用户离开指定聊天�?
// @Tags 聊天
// @Accept json
// @Produce json
// @Param room_id path int true "聊天室ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/chat/rooms/{room_id}/leave [post]
func (h *ChatHandler) LeaveChatRoom(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		return
	}

	roomIDStr := c.Param("room_id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ROOM_ID",
			Message: "聊天室ID无效",
		})
		return
	}

	err = h.chatService.LeaveChatRoom(c.Request.Context(), userID, uint(roomID))
	if err != nil {
		h.logger.Error("Failed to leave chat room", zap.Error(err), zap.Uint("user_id", userID), zap.Uint("room_id", uint(roomID)))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "LEAVE_ROOM_ERROR",
			Message: "离开聊天室失�?,
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    200,
		Message: "成功离开聊天�?,
	})
}

// GetChatMessages 获取聊天消息
// @Summary 获取聊天消息
// @Description 获取指定聊天室的消息列表
// @Tags 聊天
// @Accept json
// @Produce json
// @Param room_id path int true "聊天室ID"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(50)
// @Success 200 {object} ChatMessageListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/chat/rooms/{room_id}/messages [get]
func (h *ChatHandler) GetChatMessages(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		return
	}

	roomIDStr := c.Param("room_id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ROOM_ID",
			Message: "聊天室ID无效",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "50"))

	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 50
	}

	messages, total, err := h.chatService.GetChatMessages(c.Request.Context(), userID, uint(roomID), page, size)
	if err != nil {
		h.logger.Error("Failed to get chat messages", zap.Error(err), zap.Uint("user_id", userID), zap.Uint("room_id", uint(roomID)))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_MESSAGES_ERROR",
			Message: "获取聊天消息失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ChatMessageListResponse{
		Code:    200,
		Message: "获取成功",
		Data:    messages,
		Total:   total,
		Page:    page,
		Size:    size,
	})
}

// SendMessage 发送消�?
// @Summary 发送消�?
// @Description 向聊天室发送消�?
// @Tags 聊天
// @Accept json
// @Produce json
// @Param room_id path int true "聊天室ID"
// @Param request body SendMessageRequest true "发送消息请�?
// @Success 200 {object} ChatMessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/chat/rooms/{room_id}/messages [post]
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		return
	}

	roomIDStr := c.Param("room_id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ROOM_ID",
			Message: "聊天室ID无效",
		})
		return
	}

	var req services.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	message, err := h.chatService.SendMessage(c.Request.Context(), userID, uint(roomID), &req)
	if err != nil {
		h.logger.Error("Failed to send message", zap.Error(err), zap.Uint("user_id", userID), zap.Uint("room_id", uint(roomID)))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "SEND_MESSAGE_ERROR",
			Message: "发送消息失�?,
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ChatMessageResponse{
		Code:    200,
		Message: "消息发送成�?,
		Data:    message,
	})
}

// getUserID 从上下文获取用户ID
func (h *ChatHandler) getUserID(c *gin.Context) uint {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "用户未认�?,
		})
		return 0
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "无效的用户ID",
		})
		return 0
	}

	return uint(userID)
}

// 请求和响应结构体
type CreateChatRoomRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
	Type        string `json:"type" binding:"required,oneof=public private group"`
	MaxMembers  int    `json:"max_members" binding:"min=2,max=1000"`
}

type SendMessageRequest struct {
	Content   string `json:"content" binding:"required,min=1"`
	Type      string `json:"type" binding:"oneof=text image file"`
	ReplyToID *uint  `json:"reply_to_id,omitempty"`
}

type ChatRoomResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    *models.ChatRoom  `json:"data"`
}

type ChatRoomListResponse struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    []*models.ChatRoom `json:"data"`
	Total   int64              `json:"total"`
	Page    int                `json:"page"`
	Size    int                `json:"size"`
}

type ChatMessageResponse struct {
	Code    int                  `json:"code"`
	Message string               `json:"message"`
	Data    *models.ChatMessage  `json:"data"`
}

type ChatMessageListResponse struct {
	Code    int                   `json:"code"`
	Message string                `json:"message"`
	Data    []*models.ChatMessage `json:"data"`
	Total   int64                 `json:"total"`
	Page    int                   `json:"page"`
	Size    int                   `json:"size"`
}

type SuccessResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}
