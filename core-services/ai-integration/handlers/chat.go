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

// ChatHandler 
type ChatHandler struct {
	chatService *services.ChatService
	logger      *zap.Logger
}

// NewChatHandler 
func NewChatHandler(chatService *services.ChatService, logger *zap.Logger) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		logger:      logger,
	}
}

// Chat 
// @Summary 
// @Description AI
// @Tags AI
// @Accept json
// @Produce json
// @Param request body models.ChatRequest true ""
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
			Message: ": " + err.Error(),
		})
		return
	}

	// JWTID ()
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "",
		})
		return
	}

	// uint
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "ID",
		})
		return
	}

	req.UserID = uint(userID)

	// 
	resp, err := h.chatService.Chat(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Chat service error", zap.Error(err), zap.Uint("user_id", req.UserID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CHAT_ERROR",
			Message: ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetSessions 
// @Summary 
// @Description 
// @Tags AI
// @Accept json
// @Produce json
// @Param page query int false "" default(1)
// @Param page_size query int false "" default(20)
// @Param status query string false "" default(active)
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
			Message: ": " + err.Error(),
		})
		return
	}

	// 
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// JWTID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "",
		})
		return
	}

	// interface{}uint
	var userID uint
	switch v := userIDInterface.(type) {
	case uint:
		userID = v
	case string:
		userIDUint64, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    "INVALID_USER_ID",
				Message: "ID",
			})
			return
		}
		userID = uint(userIDUint64)
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "ID",
		})
		return
	}

	// 
	resp, err := h.chatService.GetSessions(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Get sessions error", zap.Error(err), zap.Uint("user_id", userID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_SESSIONS_ERROR",
			Message: ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetMessages 
// @Summary 
// @Description 
// @Tags AI
// @Accept json
// @Produce json
// @Param session_id path int true "ID"
// @Param page query int false "" default(1)
// @Param page_size query int false "" default(50)
// @Success 200 {object} models.MessageListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ai/sessions/{session_id}/messages [get]
func (h *ChatHandler) GetMessages(c *gin.Context) {
	// 
	sessionIDStr := c.Param("session_id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_SESSION_ID",
			Message: "ID",
		})
		return
	}

	var req models.MessageListRequest
	req.SessionID = uint(sessionID)

	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("Invalid query parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_PARAMS",
			Message: ": " + err.Error(),
		})
		return
	}

	// 
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}

	// JWTID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "",
		})
		return
	}

	// interface{}uint
	var userID uint
	switch v := userIDInterface.(type) {
	case uint:
		userID = v
	case string:
		userIDUint64, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    "INVALID_USER_ID",
				Message: "ID",
			})
			return
		}
		userID = uint(userIDUint64)
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "ID",
		})
		return
	}

	// 
	resp, err := h.chatService.GetMessages(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Get messages error", zap.Error(err),
			zap.Uint("user_id", userID),
			zap.Uint("session_id", req.SessionID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_MESSAGES_ERROR",
			Message: ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteSession 
// @Summary 
// @Description 
// @Tags AI
// @Accept json
// @Produce json
// @Param session_id path int true "ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ai/sessions/{session_id} [delete]
func (h *ChatHandler) DeleteSession(c *gin.Context) {
	// 
	sessionIDStr := c.Param("session_id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_SESSION_ID",
			Message: "ID",
		})
		return
	}

	// JWTID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "",
		})
		return
	}

	// interface{}uint
	var userID uint
	switch v := userIDInterface.(type) {
	case uint:
		userID = v
	case string:
		userIDUint64, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    "INVALID_USER_ID",
				Message: "ID",
			})
			return
		}
		userID = uint(userIDUint64)
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "ID",
		})
		return
	}

	// 
	err = h.chatService.DeleteSession(c.Request.Context(), userID, uint(sessionID))
	if err != nil {
		h.logger.Error("Delete session error", zap.Error(err),
			zap.Uint("user_id", userID),
			zap.Uint64("session_id", sessionID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "DELETE_SESSION_ERROR",
			Message: ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    "SUCCESS",
		Message: "",
	})
}

// ClearSession 
// @Summary 
// @Description 
// @Tags AI
// @Accept json
// @Produce json
// @Param session_id path int true "ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ai/sessions/{session_id}/clear [post]
func (h *ChatHandler) ClearSession(c *gin.Context) {
	sessionIDStr := c.Param("sessionId")
	if sessionIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "",
		})
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "ID",
		})
		return
	}

	//  - uintstring
	if err := h.chatService.ClearSession(c.Request.Context(), sessionIDStr, fmt.Sprintf("%d", userID)); err != nil {
		h.logger.Error("Failed to clear session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CLEAR_SESSION_ERROR",
			Message: ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    "SUCCESS",
		Message: "",
	})
}

