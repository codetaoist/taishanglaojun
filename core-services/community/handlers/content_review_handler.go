package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/core-services/community/services"
)

// ContentReviewHandler 内容审核处理
type ContentReviewHandler struct {
	reviewService *services.ContentReviewService
	logger        *zap.Logger
}

// NewContentReviewHandler 创建内容审核处理器实?
func NewContentReviewHandler(reviewService *services.ContentReviewService, logger *zap.Logger) *ContentReviewHandler {
	return &ContentReviewHandler{
		reviewService: reviewService,
		logger:        logger,
	}
}

// ReviewPost 审核帖子
func (h *ContentReviewHandler) ReviewPost(c *gin.Context) {
	// 从JWT中获取用户ID作为审核员ID
	reviewerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req services.ReviewPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// 设置审核员ID
	req.ReviewerID = reviewerID.(string)

	if err := h.reviewService.ReviewPost(&req); err != nil {
		h.logger.Error("Failed to review post", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "审核完成"})
}

// ReviewComment 审核评论
func (h *ContentReviewHandler) ReviewComment(c *gin.Context) {
	// 从JWT中获取用户ID作为审核员ID
	reviewerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req services.ReviewCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// 设置审核员ID
	req.ReviewerID = reviewerID.(string)

	if err := h.reviewService.ReviewComment(&req); err != nil {
		h.logger.Error("Failed to review comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "审核完成"})
}

// GetPendingPosts 获取待审核帖子列?
func (h *ContentReviewHandler) GetPendingPosts(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	posts, total, err := h.reviewService.GetPendingPosts(page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get pending posts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetPendingComments 获取待审核评论列?
func (h *ContentReviewHandler) GetPendingComments(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	comments, total, err := h.reviewService.GetPendingComments(page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get pending comments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": comments,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// BatchReviewPosts 批量审核帖子
func (h *ContentReviewHandler) BatchReviewPosts(c *gin.Context) {
	// 从JWT中获取用户ID作为审核员ID
	reviewerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req services.BatchReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// 设置审核员ID
	req.ReviewerID = reviewerID.(string)

	results, err := h.reviewService.BatchReviewPosts(&req)
	if err != nil {
		h.logger.Error("Failed to batch review posts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "批量审核完成",
		"results": results,
	})
}

// BatchReviewComments 批量审核评论
func (h *ContentReviewHandler) BatchReviewComments(c *gin.Context) {
	// 从JWT中获取用户ID作为审核员ID
	reviewerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req services.BatchReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// 设置审核员ID
	req.ReviewerID = reviewerID.(string)

	results, err := h.reviewService.BatchReviewComments(&req)
	if err != nil {
		h.logger.Error("Failed to batch review comments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "批量审核完成",
		"results": results,
	})
}

// GetContentStatistics 获取内容审核统计信息
func (h *ContentReviewHandler) GetContentStatistics(c *gin.Context) {
	stats, err := h.reviewService.GetContentStatistics()
	if err != nil {
		h.logger.Error("Failed to get content statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statistics": stats,
	})
}

