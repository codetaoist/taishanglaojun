package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/services"
)

// ChatHandler 对话API处理�?type ChatHandler struct {
	chatService *services.ChatService
}

// NewChatHandler 创建对话处理器实�?func NewChatHandler(chatService *services.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

// Chat 处理对话请求
// @Summary 智能对话
// @Description 与AI进行智能对话
// @Tags AI对话
// @Accept json
// @Produce json
// @Param request body services.ChatRequest true "对话请求"
// @Success 200 {object} services.ChatResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ai/chat [post]
func (h *ChatHandler) Chat(c *gin.Context) {
	var req services.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// 从上下文获取用户ID（假设已通过中间件设置）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "用户未认�?,
		})
		return
	}
	req.UserID = userID.(string)

	// 设置默认�?	if req.Temperature == 0 {
		req.Temperature = 0.7
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 2000
	}

	response, err := h.chatService.Chat(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CHAT_ERROR",
			Message: "对话处理失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetConversationHistory 获取对话历史
// @Summary 获取对话历史
// @Description 获取指定会话的对话历�?// @Tags AI对话
// @Produce json
// @Param session_id path string true "会话ID"
// @Param limit query int false "消息数量限制" default(50)
// @Success 200 {object} models.Conversation
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/ai/sessions/{session_id} [get]
func (h *ChatHandler) GetConversationHistory(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_SESSION_ID",
			Message: "会话ID不能为空",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "用户未认�?,
		})
		return
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	conversation, err := h.chatService.GetConversationHistory(
		c.Request.Context(),
		userID.(string),
		sessionID,
		limit,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "CONVERSATION_NOT_FOUND",
			Message: "对话不存�?,
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, conversation)
}

// ListConversations 获取对话列表
// @Summary 获取对话列表
// @Description 获取用户的对话列�?// @Tags AI对话
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(20)
// @Success 200 {object} ConversationListResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/ai/sessions [get]
func (h *ChatHandler) ListConversations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "用户未认�?,
		})
		return
	}

	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	size := 20
	if sizeStr := c.Query("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			size = s
		}
	}

	conversations, err := h.chatService.ListConversations(
		c.Request.Context(),
		userID.(string),
		page,
		size,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "LIST_ERROR",
			Message: "获取对话列表失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ConversationListResponse{
		Conversations: conversations,
		Page:          page,
		Size:          size,
		Total:         len(conversations),
	})
}

// DeleteConversation 删除对话
// @Summary 删除对话
// @Description 删除指定的对话会�?// @Tags AI对话
// @Param session_id path string true "会话ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/ai/sessions/{session_id} [delete]
func (h *ChatHandler) DeleteConversation(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_SESSION_ID",
			Message: "会话ID不能为空",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "用户未认�?,
		})
		return
	}

	err := h.chatService.DeleteConversation(
		c.Request.Context(),
		userID.(string),
		sessionID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "DELETE_ERROR",
			Message: "删除对话失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "对话删除成功",
	})
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Message string `json:"message"`
}

// ConversationListResponse 对话列表响应
type ConversationListResponse struct {
	Conversations interface{} `json:"conversations"`
	Page          int         `json:"page"`
	Size          int         `json:"size"`
	Total         int         `json:"total"`
}
