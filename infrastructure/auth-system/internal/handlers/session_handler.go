package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/models"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/repository"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/service"
)

// SessionHandler 会话管理处理器
type SessionHandler struct {
	sessionRepo   repository.SessionRepository
	authService   service.AuthService
	logger        *zap.Logger
}

// NewSessionHandler 创建会话管理处理器
func NewSessionHandler(sessionRepo repository.SessionRepository, authService service.AuthService, logger *zap.Logger) *SessionHandler {
	return &SessionHandler{
		sessionRepo: sessionRepo,
		authService: authService,
		logger:      logger,
	}
}

// GetSessionStats 获取会话统计信息
// @Summary 获取会话统计信息
// @Description 获取系统会话统计数据，包括总数、活跃数、过期数等
// @Tags 会话管理
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /admin/stats/sessions [get]
func (h *SessionHandler) GetSessionStats(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取总会话数
	totalSessions, err := h.sessionRepo.Count(ctx)
	if err != nil {
		h.logger.Error("Failed to get total sessions count", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to get session statistics",
		})
		return
	}

	// 获取活跃会话数
	activeSessions, err := h.sessionRepo.CountByStatus(ctx, models.SessionStatusActive)
	if err != nil {
		h.logger.Error("Failed to get active sessions count", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to get session statistics",
		})
		return
	}

	// 获取过期会话数
	expiredSessions, err := h.sessionRepo.CountByStatus(ctx, models.SessionStatusExpired)
	if err != nil {
		h.logger.Error("Failed to get expired sessions count", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to get session statistics",
		})
		return
	}

	// 获取已撤销会话数
	revokedSessions, err := h.sessionRepo.CountByStatus(ctx, models.SessionStatusRevoked)
	if err != nil {
		h.logger.Error("Failed to get revoked sessions count", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to get session statistics",
		})
		return
	}

	// 获取会话状态分布
	statusDistribution := map[string]int64{
		"active":  activeSessions,
		"expired": expiredSessions,
		"revoked": revokedSessions,
	}

	// 构建统计响应
	stats := map[string]interface{}{
		"total_sessions":       totalSessions,
		"active_sessions":      activeSessions,
		"expired_sessions":     expiredSessions,
		"revoked_sessions":     revokedSessions,
		"status_distribution":  statusDistribution,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// ListSessions 获取会话列表
// @Summary 获取会话列表
// @Description 获取系统中的会话列表，支持分页和筛选
// @Tags 会话管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Param status query string false "会话状态"
// @Param user_id query string false "用户ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /admin/sessions [get]
func (h *SessionHandler) ListSessions(c *gin.Context) {
	ctx := c.Request.Context()

	// 解析查询参数
	query := &models.SessionQuery{
		Status: models.SessionStatus(c.Query("status")),
	}

	// 解析用户ID
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			query.UserID = userID
		}
	}

	// 解析分页参数
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			query.Page = page
		} else {
			query.Page = 1
		}
	} else {
		query.Page = 1
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			query.PageSize = limit
		} else {
			query.PageSize = 20
		}
	} else {
		query.PageSize = 20
	}

	// 获取会话列表
	sessions, total, err := h.sessionRepo.List(ctx, query)
	if err != nil {
		h.logger.Error("Failed to get sessions list", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to get sessions list",
		})
		return
	}

	// 计算分页信息
	totalPages := (total + int64(query.PageSize) - 1) / int64(query.PageSize)

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.Header("X-Page-Count", strconv.FormatInt(totalPages, 10))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"sessions":     sessions,
			"total":        total,
			"page":         query.Page,
			"limit":        query.PageSize,
			"total_pages":  totalPages,
		},
	})
}

// RevokeSession 撤销指定会话
// @Summary 撤销指定会话
// @Description 管理员撤销指定的用户会话
// @Tags 会话管理
// @Param session_id path string true "会话ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /admin/sessions/{session_id} [delete]
func (h *SessionHandler) RevokeSession(c *gin.Context) {
	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid session ID format",
		})
		return
	}

	if err := h.authService.RevokeSession(c.Request.Context(), sessionID); err != nil {
		if err == service.ErrSessionNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "session_not_found",
				"message": "Session not found",
			})
			return
		}

		h.logger.Error("Failed to revoke session", 
			zap.String("session_id", sessionID.String()),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to revoke session",
		})
		return
	}

	h.logger.Info("Session revoked by admin", 
		zap.String("session_id", sessionID.String()),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Session revoked successfully",
	})
}

// RevokeUserSessions 撤销用户的所有会话
// @Summary 撤销用户的所有会话
// @Description 管理员撤销指定用户的所有活跃会话
// @Tags 会话管理
// @Param user_id path string true "用户ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /admin/users/{user_id}/sessions [delete]
func (h *SessionHandler) RevokeUserSessions(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid user ID format",
		})
		return
	}

	if err := h.authService.RevokeAllSessions(c.Request.Context(), userID); err != nil {
		h.logger.Error("Failed to revoke user sessions", 
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to revoke user sessions",
		})
		return
	}

	h.logger.Info("All user sessions revoked by admin", 
		zap.String("user_id", userID.String()),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "All user sessions revoked successfully",
	})
}

// CleanupExpiredSessions 清理过期会话
// @Summary 清理过期会话
// @Description 清理系统中的过期和已撤销会话
// @Tags 会话管理
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /admin/sessions/cleanup [post]
func (h *SessionHandler) CleanupExpiredSessions(c *gin.Context) {
	ctx := c.Request.Context()

	// 清理过期会话
	expiredCount, err := h.sessionRepo.CleanupExpiredSessions(ctx)
	if err != nil {
		h.logger.Error("Failed to cleanup expired sessions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to cleanup expired sessions",
		})
		return
	}

	// 清理已撤销会话
	revokedCount, err := h.sessionRepo.CleanupRevokedSessions(ctx)
	if err != nil {
		h.logger.Error("Failed to cleanup revoked sessions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to cleanup revoked sessions",
		})
		return
	}

	totalCleaned := expiredCount + revokedCount

	h.logger.Info("Sessions cleanup completed", 
		zap.Int64("expired_cleaned", expiredCount),
		zap.Int64("revoked_cleaned", revokedCount),
		zap.Int64("total_cleaned", totalCleaned),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"expired_sessions_cleaned": expiredCount,
			"revoked_sessions_cleaned": revokedCount,
			"total_sessions_cleaned":   totalCleaned,
		},
		"message": "Sessions cleanup completed successfully",
	})
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// SuccessResponse 成功响应结构
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}