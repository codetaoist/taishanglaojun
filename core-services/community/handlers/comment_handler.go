package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/taishanglaojun/core-services/community/models"
	"github.com/codetaoist/taishanglaojun/core-services/community/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CommentHandler 评论处理器
type CommentHandler struct {
	commentService *services.CommentService
	logger         *zap.Logger
}

// NewCommentHandler 创建评论处理器实例
func NewCommentHandler(commentService *services.CommentService, logger *zap.Logger) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
		logger:         logger,
	}
}

// CreateComment 创建评论
// @Summary 创建评论
// @Description 创建新的评论或回复
// @Tags 评论管理
// @Accept json
// @Produce json
// @Param request body models.CreateCommentRequest true "创建评论请求"
// @Success 201 {object} models.CommentResponse "创建成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /community/comments [post]
func (h *CommentHandler) CreateComment(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	var req models.CommentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 验证请求参数
	if req.PostID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID不能为空"})
		return
	}
	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评论内容不能为空"})
		return
	}

	// 创建评论
	comment, err := h.commentService.CreateComment(userID.(string), &req)
	if err != nil {
		h.logger.Error("Failed to create comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建评论失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, comment.ToResponse())
}

// GetComment 获取评论详情
// @Summary 获取评论详情
// @Description 根据ID获取评论详细信息
// @Tags 评论管理
// @Produce json
// @Param id path string true "评论ID"
// @Success 200 {object} models.CommentResponse "获取成功"
// @Failure 404 {object} map[string]interface{} "评论不存在"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /community/comments/{id} [get]
func (h *CommentHandler) GetComment(c *gin.Context) {
	commentID := c.Param("id")
	if commentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评论ID不能为空"})
		return
	}

	comment, err := h.commentService.GetComment(commentID)
	if err != nil {
		h.logger.Error("Failed to get comment", zap.String("comment_id", commentID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "评论不存在"})
		return
	}

	c.JSON(http.StatusOK, comment.ToResponse())
}

// GetPostComments 获取帖子评论列表
// @Summary 获取帖子评论列表
// @Description 分页获取指定帖子的评论列表
// @Tags 评论管理
// @Produce json
// @Param post_id path string true "帖子ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param sort query string false "排序方式" Enums(latest,oldest,hot)
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /community/posts/{post_id}/comments [get]
func (h *CommentHandler) GetPostComments(c *gin.Context) {
	postID := c.Param("id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID不能为空"})
		return
	}

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	sort := c.DefaultQuery("sort", "latest")

	// 验证参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 构建请求参数
	req := &models.CommentListRequest{
		PostID:   postID,
		Page:     page,
		PageSize: pageSize,
		SortBy:   sort,
	}

	// 获取当前用户ID（如果已登录）
	var userID *string
	if uid, exists := c.Get("user_id"); exists {
		if uidStr, ok := uid.(string); ok {
			userID = &uidStr
		}
	}

	// 获取评论列表
	response, err := h.commentService.GetComments(req, userID)
	if err != nil {
		h.logger.Error("Failed to get post comments", zap.String("post_id", postID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取评论列表失败"})
		return
	}

	// 计算分页信息
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

// GetCommentReplies 获取评论回复列表
// @Summary 获取评论回复列表
// @Description 分页获取指定评论的回复列表
// @Tags 评论管理
// @Produce json
// @Param comment_id path string true "评论ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /community/comments/{comment_id}/replies [get]
func (h *CommentHandler) GetCommentReplies(c *gin.Context) {
	commentID := c.Param("comment_id")
	if commentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评论ID不能为空"})
		return
	}

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 验证参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 获取回复列表
	response, err := h.commentService.GetReplies(commentID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get comment replies", zap.String("comment_id", commentID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取回复列表失败"})
		return
	}

	// 计算分页信息
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

// UpdateComment 更新评论
// @Summary 更新评论
// @Description 更新评论内容（仅作者可操作）
// @Tags 评论管理
// @Accept json
// @Produce json
// @Param id path string true "评论ID"
// @Param request body models.UpdateCommentRequest true "更新评论请求"
// @Success 200 {object} models.CommentResponse "更新成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} map[string]interface{} "无权限"
// @Failure 404 {object} map[string]interface{} "评论不存在"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /community/comments/{id} [put]
func (h *CommentHandler) UpdateComment(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	commentID := c.Param("id")
	if commentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评论ID不能为空"})
		return
	}

	var req models.CommentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 验证请求参数
	if req.Content == nil || *req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评论内容不能为空"})
		return
	}

	// 更新评论
	comment, err := h.commentService.UpdateComment(commentID, userID.(string), &req)
	if err != nil {
		h.logger.Error("Failed to update comment", zap.String("comment_id", commentID), zap.Error(err))
		if err.Error() == "评论不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == "无权限操作" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新评论失败"})
		}
		return
	}

	c.JSON(http.StatusOK, comment.ToResponse())
}

// DeleteComment 删除评论
// @Summary 删除评论
// @Description 删除评论（仅作者可操作）
// @Tags 评论管理
// @Produce json
// @Param id path string true "评论ID"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} map[string]interface{} "无权限"
// @Failure 404 {object} map[string]interface{} "评论不存在"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /community/comments/{id} [delete]
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	commentID := c.Param("id")
	if commentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评论ID不能为空"})
		return
	}

	// 删除评论
	err := h.commentService.DeleteComment(commentID, userID.(string))
	if err != nil {
		h.logger.Error("Failed to delete comment", zap.String("comment_id", commentID), zap.Error(err))
		if err.Error() == "评论不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == "无权限操作" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除评论失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetCommentStats 获取评论统计
// @Summary 获取评论统计
// @Description 获取评论相关统计信息
// @Tags 评论管理
// @Produce json
// @Success 200 {object} models.CommentStatsResponse "获取成功"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /community/comments/stats [get]
func (h *CommentHandler) GetCommentStats(c *gin.Context) {
	stats, err := h.commentService.GetCommentStats()
	if err != nil {
		h.logger.Error("Failed to get comment stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计信息失败"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetUserComments 获取用户评论列表
// @Summary 获取用户评论列表
// @Description 分页获取指定用户的评论列表
// @Tags 评论管理
// @Produce json
// @Param user_id path string true "用户ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /community/users/{user_id}/comments [get]
func (h *CommentHandler) GetUserComments(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 验证参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 获取用户评论列表
	response, err := h.commentService.GetUserComments(userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get user comments", zap.String("user_id", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户评论失败"})
		return
	}

	// 计算分页信息
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