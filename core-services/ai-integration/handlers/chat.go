package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/services"
)

// ChatHandler 对话处理�?
type ChatHandler struct {
	chatService *services.ChatService
	logger      *zap.Logger
}

// NewChatHandler 创建对话处理�?
func NewChatHandler(chatService *services.ChatService, logger *zap.Logger) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		logger:      logger,
	}
}

// Chat 发送对话消�?
// @Summary 发送对话消�?
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

	// 从JWT中获取用户ID (这里假设已经通过中间件验�?
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "UNAUTHORIZED",
			"message": "用户未认�?,
		})
		return
	}

	// 将字符串转换为uint
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_USER_ID",
			"message": "无效的用户ID",
		})
		return
	}

	req.UserID = uint(userID)

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
// @Description 获取用户的对话会话列�?
// @Tags AI对话
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Param status query string false "会话状�?
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

	// 设置默认�?
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 从JWT中获取用户ID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "用户未认�?,
		})
		return
	}

	// 将interface{}转换为uint
	var userID uint
	switch v := userIDInterface.(type) {
	case uint:
		userID = v
	case string:
		userIDUint64, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    "INVALID_USER_ID",
				Message: "无效的用户ID",
			})
			return
		}
		userID = uint(userIDUint64)
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "无效的用户ID类型",
		})
		return
	}

	// 调用服务
	resp, err := h.chatService.GetSessions(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Get sessions error", zap.Error(err), zap.Uint("user_id", userID))
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
// @Description 获取指定会话的消息列�?
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

	// 设置默认�?
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}

	// 从JWT中获取用户ID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "用户未认�?,
		})
		return
	}

	// 将interface{}转换为uint
	var userID uint
	switch v := userIDInterface.(type) {
	case uint:
		userID = v
	case string:
		userIDUint64, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    "INVALID_USER_ID",
				Message: "无效的用户ID",
			})
			return
		}
		userID = uint(userIDUint64)
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "无效的用户ID类型",
		})
		return
	}

	// 调用服务
	resp, err := h.chatService.GetMessages(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Get messages error", zap.Error(err),
			zap.Uint("user_id", userID),
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
// @Description 删除指定的对话会�?
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
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "用户未认�?,
		})
		return
	}

	// 将interface{}转换为uint
	var userID uint
	switch v := userIDInterface.(type) {
	case uint:
		userID = v
	case string:
		userIDUint64, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    "INVALID_USER_ID",
				Message: "无效的用户ID",
			})
			return
		}
		userID = uint(userIDUint64)
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "无效的用户ID类型",
		})
		return
	}

	// 调用服务
	err = h.chatService.DeleteSession(c.Request.Context(), userID, uint(sessionID))
	if err != nil {
		h.logger.Error("Delete session error", zap.Error(err),
			zap.Uint("user_id", userID),
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

// ClearSession 清空会话消息
// @Summary 清空会话消息
// @Description 清空指定会话的所有消息，但保留会�?
// @Tags AI对话
// @Accept json
// @Produce json
// @Param session_id path int true "会话ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ai/sessions/{session_id}/clear [post]
func (h *ChatHandler) ClearSession(c *gin.Context) {
	sessionIDStr := c.Param("sessionId")
	if sessionIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话ID不能为空"})
		return
	}

	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授�?})
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}

	// 修复类型错误 - 将uint转换为string
	if err := h.chatService.ClearSession(c.Request.Context(), sessionIDStr, fmt.Sprintf("%d", userID)); err != nil {
		h.logger.Error("Failed to clear session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清空会话失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "会话已清�?})
}
