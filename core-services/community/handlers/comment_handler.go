package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/taishanglaojun/core-services/community/models"
	"github.com/codetaoist/taishanglaojun/core-services/community/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CommentHandler 
type CommentHandler struct {
	commentService *services.CommentService
	logger         *zap.Logger
}

// NewCommentHandler 
func NewCommentHandler(commentService *services.CommentService, logger *zap.Logger) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
		logger:         logger,
	}
}

// CreateComment 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param request body models.CreateCommentRequest true ""
// @Success 201 {object} models.CommentResponse ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 401 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/comments [post]
func (h *CommentHandler) CreateComment(c *gin.Context) {
	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
		return
	}

	var req models.CommentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "", "details": err.Error()})
		return
	}

	// 
	if req.PostID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}
	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": ""})
		return
	}

	// 
	comment, err := h.commentService.CreateComment(userID.(string), &req)
	if err != nil {
		h.logger.Error("Failed to create comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, comment.ToResponse())
}

// GetComment 
// @Summary 
// @Description ID
// @Tags 
// @Produce json
// @Param id path string true "ID"
// @Success 200 {object} models.CommentResponse ""
// @Failure 404 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/comments/{id} [get]
func (h *CommentHandler) GetComment(c *gin.Context) {
	commentID := c.Param("id")
	if commentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	comment, err := h.commentService.GetComment(commentID)
	if err != nil {
		h.logger.Error("Failed to get comment", zap.String("comment_id", commentID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": ""})
		return
	}

	c.JSON(http.StatusOK, comment.ToResponse())
}

// GetPostComments 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param post_id path string true "ID"
// @Param page query int false "" default(1)
// @Param page_size query int false "" default(20)
// @Param sort query string false "" Enums(latest,oldest,hot)
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/posts/{post_id}/comments [get]
func (h *CommentHandler) GetPostComments(c *gin.Context) {
	postID := c.Param("id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	// 
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	sort := c.DefaultQuery("sort", "latest")

	// 
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 
	req := &models.CommentListRequest{
		PostID:   postID,
		Page:     page,
		PageSize: pageSize,
		SortBy:   sort,
	}

	// ID)
	var userID *string
	if uid, exists := c.Get("user_id"); exists {
		if uidStr, ok := uid.(string); ok {
			userID = &uidStr
		}
	}

	// 
	response, err := h.commentService.GetComments(req, userID)
	if err != nil {
		h.logger.Error("Failed to get post comments", zap.String("post_id", postID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	// 
	totalPages := (int(response.Total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"comments": response.Comments,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       response.Total,
			"total_pages": totalPages,
		},
	})
}

// GetCommentReplies 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param comment_id path string true "ID"
// @Param page query int false "" default(1)
// @Param page_size query int false "" default(20)
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/comments/{comment_id}/replies [get]
func (h *CommentHandler) GetCommentReplies(c *gin.Context) {
	commentID := c.Param("comment_id")
	if commentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	// 
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 
	response, err := h.commentService.GetReplies(commentID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get comment replies", zap.String("comment_id", commentID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	// 
	totalPages := (int(response.Total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"replies": response.Comments,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       response.Total,
			"total_pages": totalPages,
		},
	})
}

// UpdateComment 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param request body models.UpdateCommentRequest true ""
// @Success 200 {object} models.CommentResponse ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 401 {object} map[string]interface{} ""
// @Failure 403 {object} map[string]interface{} ""
// @Failure 404 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/comments/{id} [put]
func (h *CommentHandler) UpdateComment(c *gin.Context) {
	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
		return
	}

	commentID := c.Param("id")
	if commentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	var req models.CommentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "", "details": err.Error()})
		return
	}

	// 
	if req.Content == nil || *req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": ""})
		return
	}

	// 
	comment, err := h.commentService.UpdateComment(commentID, userID.(string), &req)
	if err != nil {
		h.logger.Error("Failed to update comment", zap.String("comment_id", commentID), zap.Error(err))
		if err.Error() == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		}
		return
	}

	c.JSON(http.StatusOK, comment.ToResponse())
}

// DeleteComment 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param id path string true "ID"
// @Success 200 {object} map[string]interface{} ""
// @Failure 401 {object} map[string]interface{} ""
// @Failure 403 {object} map[string]interface{} ""
// @Failure 404 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/comments/{id} [delete]
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
		return
	}

	commentID := c.Param("id")
	if commentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	// 
	err := h.commentService.DeleteComment(commentID, userID.(string))
	if err != nil {
		h.logger.Error("Failed to delete comment", zap.String("comment_id", commentID), zap.Error(err))
		if err.Error() == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": ""})
}

// GetCommentStats 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Success 200 {object} models.CommentStatsResponse ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/comments/stats [get]
func (h *CommentHandler) GetCommentStats(c *gin.Context) {
	stats, err := h.commentService.GetCommentStats()
	if err != nil {
		h.logger.Error("Failed to get comment stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetUserComments 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param user_id path string true "ID"
// @Param page query int false "" default(1)
// @Param page_size query int false "" default(20)
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/users/{user_id}/comments [get]
func (h *CommentHandler) GetUserComments(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	// 
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 
	response, err := h.commentService.GetUserComments(userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get user comments", zap.String("user_id", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	// 
	totalPages := (int(response.Total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"comments": response.Comments,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       response.Total,
			"total_pages": totalPages,
		},
	})
}

