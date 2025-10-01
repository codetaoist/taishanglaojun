package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/taishanglaojun/core-services/ai-integration/models"
	"github.com/taishanglaojun/core-services/ai-integration/services"
	"go.uber.org/zap"
)

// ChatHandler 对话处理器
type ChatHandler struct {
	chatService *services.ChatService
	logger      *zap.Logger
}

// NewChatHandler 创建对话处理器
func NewChatHandler(chatService *services.ChatService, logger *zap.Logger) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		logger:      logger,
	}
}

// Chat 发送对话消息
// @Summary 发送对话消息
// @Description 发送消息到AI进行对话
// @Tags AI对话
// @Accept json
// @Produce json
// @Param request body models.ChatRequest true "对话请求"
// @Success 200 {object} models.ChatResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ai/chat [post]
func (h *ChatHandler) Chat(c *gin.Context) {
	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效: " + err.Error(),
		})
		return
	}

	// 从JWT中获取用户ID (这里假设已经通过中间件验证)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "用户未认证",
		})
		return
	}

	req.UserID = userID.(uint)

	// 调用服务
	resp, err := h.chatService.Chat(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Chat service error", zap.Error(err), zap.Uint("user_id", req.UserID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CHAT_ERROR",
			Message: "对话处理失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetSessions 获取用户会话列表
// @Summary 获取会话列表
// @Description 获取用户的对话会话列表
// @Tags AI对话
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Param status query string false "会话状态"
// @Success 200 {object} models.SessionListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ai/sessions [get]
func (h *ChatHandler) GetSessions(c *gin.Context) {
	var req models.SessionListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("Invalid query parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_PARAMS",
			Message: "查询参数无效: " + err.Error(),
		})
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "用户未认证",
		})
		return
	}

	// 调用服务
	resp, err := h.chatService.GetSessions(c.Request.Context(), userID.(uint), &req)
	if err != nil {
		h.logger.Error("Get sessions error", zap.Error(err), zap.Uint("user_id", userID.(uint)))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_SESSIONS_ERROR",
			Message: "获取会话列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetMessages 获取会话消息列表
// @Summary 获取消息列表
// @Description 获取指定会话的消息列表
// @Tags AI对话
// @Accept json
// @Produce json
// @Param session_id path int true "会话ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(50)
// @Success 200 {object} models.MessageListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ai/sessions/{session_id}/messages [get]
func (h *ChatHandler) GetMessages(c *gin.Context) {
	// 解析路径参数
	sessionIDStr := c.Param("session_id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_SESSION_ID",
			Message: "会话ID无效",
		})
		return
	}

	var req models.MessageListRequest
	req.SessionID = uint(sessionID)

	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("Invalid query parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_PARAMS",
			Message: "查询参数无效: " + err.Error(),
		})
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "用户未认证",
		})
		return
	}

	// 调用服务
	resp, err := h.chatService.GetMessages(c.Request.Context(), userID.(uint), &req)
	if err != nil {
		h.logger.Error("Get messages error", zap.Error(err),
			zap.Uint("user_id", userID.(uint)),
			zap.Uint("session_id", req.SessionID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_MESSAGES_ERROR",
			Message: "获取消息列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteSession 删除会话
// @Summary 删除会话
// @Description 删除指定的对话会话
// @Tags AI对话
// @Accept json
// @Produce json
// @Param session_id path int true "会话ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ai/sessions/{session_id} [delete]
func (h *ChatHandler) DeleteSession(c *gin.Context) {
	// 解析路径参数
	sessionIDStr := c.Param("session_id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_SESSION_ID",
			Message: "会话ID无效",
		})
		return
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "用户未认证",
		})
		return
	}

	// 调用服务
	err = h.chatService.DeleteSession(c.Request.Context(), userID.(uint), uint(sessionID))
	if err != nil {
		h.logger.Error("Delete session error", zap.Error(err),
			zap.Uint("user_id", userID.(uint)),
			zap.Uint64("session_id", sessionID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "DELETE_SESSION_ERROR",
			Message: "删除会话失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    "SUCCESS",
		Message: "会话删除成功",
	})
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
