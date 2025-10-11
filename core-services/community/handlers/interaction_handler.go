package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/taishanglaojun/core-services/community/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InteractionHandler 互动处理�?
type InteractionHandler struct {
	interactionService *services.InteractionService
	logger             *zap.Logger
}

// NewInteractionHandler 创建互动处理器实�?
func NewInteractionHandler(interactionService *services.InteractionService, logger *zap.Logger) *InteractionHandler {
	return &InteractionHandler{
		interactionService: interactionService,
		logger:             logger,
	}
}

// LikePost 点赞帖子
// @Summary 点赞帖子
// @Description 对指定帖子进行点�?
// @Tags 互动管理
// @Produce json
// @Param post_id path string true "帖子ID"
// @Success 201 {object} models.LikeResponse "点赞成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授�?
// @Failure 409 {object} map[string]interface{} "已经点赞"
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /community/posts/{post_id}/like [post]
func (h *InteractionHandler) LikePost(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访�?})
		return
	}

	postID := c.Param("id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID不能为空"})
		return
	}

	like, err := h.interactionService.LikePost(userID.(string), postID)
	if err != nil {
		h.logger.Error("Failed to like post", zap.String("user_id", userID.(string)), zap.String("post_id", postID), zap.Error(err))
		if err.Error() == "已经点赞过了" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else if err.Error() == "帖子不存�? {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "点赞失败"})
		}
		return
	}

	c.JSON(http.StatusCreated, like.ToResponse())
}

// UnlikePost 取消点赞帖子
// @Summary 取消点赞帖子
// @Description 取消对指定帖子的点赞
// @Tags 互动管理
// @Produce json
// @Param post_id path string true "帖子ID"
// @Success 200 {object} map[string]interface{} "取消点赞成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授�?
// @Failure 404 {object} map[string]interface{} "未找到点赞记�?
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /community/posts/{post_id}/like [delete]
func (h *InteractionHandler) UnlikePost(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访�?})
		return
	}

	postID := c.Param("id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID不能为空"})
		return
	}

	err := h.interactionService.UnlikePost(userID.(string), postID)
	if err != nil {
		h.logger.Error("Failed to unlike post", zap.String("user_id", userID.(string)), zap.String("post_id", postID), zap.Error(err))
		if err.Error() == "未找到点赞记�? {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "取消点赞失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "取消点赞成功"})
}

// LikeComment 点赞评论
// @Summary 点赞评论
// @Description 对指定评论进行点�?
// @Tags 互动管理
// @Produce json
// @Param comment_id path string true "评论ID"
// @Success 201 {object} models.LikeResponse "点赞成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授�?
// @Failure 409 {object} map[string]interface{} "已经点赞"
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /community/comments/{comment_id}/like [post]
func (h *InteractionHandler) LikeComment(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访�?})
		return
	}

	commentID := c.Param("id")
	if commentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评论ID不能为空"})
		return
	}

	like, err := h.interactionService.LikeComment(userID.(string), commentID)
	if err != nil {
		h.logger.Error("Failed to like comment", zap.String("user_id", userID.(string)), zap.String("comment_id", commentID), zap.Error(err))
		if err.Error() == "已经点赞过了" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else if err.Error() == "评论不存�? {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "点赞失败"})
		}
		return
	}

	c.JSON(http.StatusCreated, like.ToResponse())
}

// UnlikeComment 取消点赞评论
// @Summary 取消点赞评论
// @Description 取消对指定评论的点赞
// @Tags 互动管理
// @Produce json
// @Param comment_id path string true "评论ID"
// @Success 200 {object} map[string]interface{} "取消点赞成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授�?
// @Failure 404 {object} map[string]interface{} "未找到点赞记�?
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /community/comments/{comment_id}/like [delete]
func (h *InteractionHandler) UnlikeComment(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访�?})
		return
	}

	commentID := c.Param("id")
	if commentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评论ID不能为空"})
		return
	}

	err := h.interactionService.UnlikeComment(userID.(string), commentID)
	if err != nil {
		h.logger.Error("Failed to unlike comment", zap.String("user_id", userID.(string)), zap.String("comment_id", commentID), zap.Error(err))
		if err.Error() == "未找到点赞记�? {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "取消点赞失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "取消点赞成功"})
}

// FollowUser 关注用户
// @Summary 关注用户
// @Description 关注指定用户
// @Tags 互动管理
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 201 {object} models.FollowResponse "关注成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授�?
// @Failure 409 {object} map[string]interface{} "已经关注"
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /community/users/{user_id}/follow [post]
func (h *InteractionHandler) FollowUser(c *gin.Context) {
	// 获取用户ID
	followerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访�?})
		return
	}

	followingID := c.Param("id")
	if followingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	follow, err := h.interactionService.FollowUser(followerID.(string), followingID)
	if err != nil {
		h.logger.Error("Failed to follow user", zap.String("follower_id", followerID.(string)), zap.String("following_id", followingID), zap.Error(err))
		if err.Error() == "已经关注过了" || err.Error() == "不能关注自己" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else if err.Error() == "用户不存�? {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "关注失败"})
		}
		return
	}

	c.JSON(http.StatusCreated, follow.ToResponse())
}

// UnfollowUser 取消关注用户
// @Summary 取消关注用户
// @Description 取消关注指定用户
// @Tags 互动管理
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} map[string]interface{} "取消关注成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授�?
// @Failure 404 {object} map[string]interface{} "未找到关注记�?
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /community/users/{user_id}/follow [delete]
func (h *InteractionHandler) UnfollowUser(c *gin.Context) {
	// 获取用户ID
	followerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访�?})
		return
	}

	followingID := c.Param("id")
	if followingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	err := h.interactionService.UnfollowUser(followerID.(string), followingID)
	if err != nil {
		h.logger.Error("Failed to unfollow user", zap.String("follower_id", followerID.(string)), zap.String("following_id", followingID), zap.Error(err))
		if err.Error() == "未找到关注记�? {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "取消关注失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "取消关注成功"})
}

// BookmarkPost 收藏帖子
// @Summary 收藏帖子
// @Description 收藏指定帖子
// @Tags 互动管理
// @Produce json
// @Param post_id path string true "帖子ID"
// @Success 201 {object} models.BookmarkResponse "收藏成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授�?
// @Failure 409 {object} map[string]interface{} "已经收藏"
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /community/posts/{post_id}/bookmark [post]
func (h *InteractionHandler) BookmarkPost(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访�?})
		return
	}

	postID := c.Param("id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID不能为空"})
		return
	}

	bookmark, err := h.interactionService.BookmarkPost(userID.(string), postID)
	if err != nil {
		h.logger.Error("Failed to bookmark post", zap.String("user_id", userID.(string)), zap.String("post_id", postID), zap.Error(err))
		if err.Error() == "已经收藏过了" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else if err.Error() == "帖子不存�? {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "收藏失败"})
		}
		return
	}

	c.JSON(http.StatusCreated, bookmark.ToResponse())
}

// UnbookmarkPost 取消收藏帖子
// @Summary 取消收藏帖子
// @Description 取消收藏指定帖子
// @Tags 互动管理
// @Produce json
// @Param post_id path string true "帖子ID"
// @Success 200 {object} map[string]interface{} "取消收藏成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授�?
// @Failure 404 {object} map[string]interface{} "未找到收藏记�?
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /community/posts/{post_id}/bookmark [delete]
func (h *InteractionHandler) UnbookmarkPost(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访�?})
		return
	}

	postID := c.Param("id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID不能为空"})
		return
	}

	err := h.interactionService.UnbookmarkPost(userID.(string), postID)
	if err != nil {
		h.logger.Error("Failed to unbookmark post", zap.String("user_id", userID.(string)), zap.String("post_id", postID), zap.Error(err))
		if err.Error() == "未找到收藏记�? {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "取消收藏失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "取消收藏成功"})
}

// GetMyBookmarks 获取我的收藏列表
// @Summary 获取我的收藏列表
// @Description 分页获取当前用户的收藏列�?
// @Tags 互动管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授�?
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /community/bookmarks [get]
func (h *InteractionHandler) GetMyBookmarks(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访�?})
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

	bookmarks, total, err := h.interactionService.GetUserBookmarks(userID.(string), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get user bookmarks", zap.String("user_id", userID.(string)), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取收藏列表失败"})
		return
	}

	// 计算分页信息
	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"bookmarks": bookmarks,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetUserFollowers 获取用户粉丝列表
// @Summary 获取用户粉丝列表
// @Description 分页获取指定用户的粉丝列�?
// @Tags 互动管理
// @Produce json
// @Param user_id path string true "用户ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /community/users/{user_id}/followers [get]
func (h *InteractionHandler) GetUserFollowers(c *gin.Context) {
	userID := c.Param("id")
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

	followers, total, err := h.interactionService.GetUserFollowers(userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get user followers", zap.String("user_id", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取粉丝列表失败"})
		return
	}

	// 计算分页信息
	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"followers": followers,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetUserFollowing 获取用户关注列表
// @Summary 获取用户关注列表
// @Description 分页获取指定用户的关注列�?
// @Tags 互动管理
// @Produce json
// @Param user_id path string true "用户ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /community/users/{user_id}/following [get]
func (h *InteractionHandler) GetUserFollowing(c *gin.Context) {
	userID := c.Param("id")
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

	following, total, err := h.interactionService.GetUserFollowing(userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get user following", zap.String("user_id", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取关注列表失败"})
		return
	}

	// 计算分页信息
	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"following": following,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetInteractionStats 获取互动统计
// @Summary 获取互动统计
// @Description 获取互动相关统计信息
// @Tags 互动管理
// @Produce json
// @Success 200 {object} models.InteractionStatsResponse "获取成功"
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /community/interactions/stats [get]
func (h *InteractionHandler) GetInteractionStats(c *gin.Context) {
	stats, err := h.interactionService.GetInteractionStats()
	if err != nil {
		h.logger.Error("Failed to get interaction stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计信息失败"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// CheckInteractionStatus 检查互动状�?
// @Summary 检查互动状�?
// @Description 检查用户对帖子/评论/用户的互动状态（点赞、关注、收藏）
// @Tags 互动管理
// @Produce json
// @Param type query string true "类型" Enums(post_like,comment_like,user_follow,post_bookmark)
// @Param target_id query string true "目标ID"
// @Success 200 {object} map[string]interface{} "检查成�?
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授�?
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /community/interactions/status [get]
func (h *InteractionHandler) CheckInteractionStatus(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访�?})
		return
	}

	interactionType := c.Query("type")
	targetID := c.Query("target_id")

	if interactionType == "" || targetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "类型和目标ID不能为空"})
		return
	}

	var status bool
	var err error

	switch interactionType {
	case "post_like":
		status, err = h.interactionService.IsPostLiked(userID.(string), targetID)
	case "comment_like":
		status, err = h.interactionService.IsCommentLiked(userID.(string), targetID)
	case "user_follow":
		status, err = h.interactionService.IsUserFollowed(userID.(string), targetID)
	case "post_bookmark":
		status, err = h.interactionService.IsPostBookmarked(userID.(string), targetID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的互动类型"})
		return
	}

	if err != nil {
		h.logger.Error("Failed to check interaction status", 
			zap.String("user_id", userID.(string)), 
			zap.String("type", interactionType), 
			zap.String("target_id", targetID), 
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查状态失�?})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"type":      interactionType,
		"target_id": targetID,
		"status":    status,
	})
}
