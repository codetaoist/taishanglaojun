package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/taishanglaojun/core-services/community/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InteractionHandler 
type InteractionHandler struct {
	interactionService *services.InteractionService
	logger             *zap.Logger
}

// NewInteractionHandler 
func NewInteractionHandler(interactionService *services.InteractionService, logger *zap.Logger) *InteractionHandler {
	return &InteractionHandler{
		interactionService: interactionService,
		logger:             logger,
	}
}

// LikePost 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param post_id path string true "ID"
// @Success 201 {object} models.LikeResponse ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 401 {object} map[string]interface{} ""
// @Failure 409 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/posts/{post_id}/like [post]
func (h *InteractionHandler) LikePost(c *gin.Context) {
	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
		return
	}

	postID := c.Param("id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	like, err := h.interactionService.LikePost(userID.(string), postID)
	if err != nil {
		h.logger.Error("Failed to like post", zap.String("user_id", userID.(string)), zap.String("post_id", postID), zap.Error(err))
		if err.Error() == "" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else if err.Error() == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		}
		return
	}

	c.JSON(http.StatusCreated, like.ToResponse())
}

// UnlikePost 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param post_id path string true "ID"
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 401 {object} map[string]interface{} ""
// @Failure 404 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/posts/{post_id}/like [delete]
func (h *InteractionHandler) UnlikePost(c *gin.Context) {
	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
		return
	}

	postID := c.Param("id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	err := h.interactionService.UnlikePost(userID.(string), postID)
	if err != nil {
		h.logger.Error("Failed to unlike post", zap.String("user_id", userID.(string)), zap.String("post_id", postID), zap.Error(err))
		if err.Error() == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": ""})
}

// LikeComment 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param comment_id path string true "ID"
// @Success 201 {object} models.LikeResponse ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 401 {object} map[string]interface{} ""
// @Failure 409 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/comments/{comment_id}/like [post]
func (h *InteractionHandler) LikeComment(c *gin.Context) {
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

	like, err := h.interactionService.LikeComment(userID.(string), commentID)
	if err != nil {
		h.logger.Error("Failed to like comment", zap.String("user_id", userID.(string)), zap.String("comment_id", commentID), zap.Error(err))
		if err.Error() == "" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else if err.Error() == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		}
		return
	}

	c.JSON(http.StatusCreated, like.ToResponse())
}

// UnlikeComment 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param comment_id path string true "ID"
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 401 {object} map[string]interface{} ""
// @Failure 404 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/comments/{comment_id}/like [delete]
func (h *InteractionHandler) UnlikeComment(c *gin.Context) {
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

	err := h.interactionService.UnlikeComment(userID.(string), commentID)
	if err != nil {
		h.logger.Error("Failed to unlike comment", zap.String("user_id", userID.(string)), zap.String("comment_id", commentID), zap.Error(err))
		if err.Error() == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": ""})
}

// FollowUser 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param user_id path string true "ID"
// @Success 201 {object} models.FollowResponse ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 401 {object} map[string]interface{} ""
// @Failure 409 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/users/{user_id}/follow [post]
func (h *InteractionHandler) FollowUser(c *gin.Context) {
	// ID
	followerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
		return
	}

	followingID := c.Param("id")
	if followingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	follow, err := h.interactionService.FollowUser(followerID.(string), followingID)
	if err != nil {
		h.logger.Error("Failed to follow user", zap.String("follower_id", followerID.(string)), zap.String("following_id", followingID), zap.Error(err))
		if err.Error() == "" || err.Error() == "" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else if err.Error() == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		}
		return
	}

	c.JSON(http.StatusCreated, follow.ToResponse())
}

// UnfollowUser 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param user_id path string true "ID"
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 401 {object} map[string]interface{} ""
// @Failure 404 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/users/{user_id}/follow [delete]
func (h *InteractionHandler) UnfollowUser(c *gin.Context) {
	// ID
	followerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
		return
	}

	followingID := c.Param("id")
	if followingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	err := h.interactionService.UnfollowUser(followerID.(string), followingID)
	if err != nil {
		h.logger.Error("Failed to unfollow user", zap.String("follower_id", followerID.(string)), zap.String("following_id", followingID), zap.Error(err))
		if err.Error() == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": ""})
}

// BookmarkPost 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param post_id path string true "ID"
// @Success 201 {object} models.BookmarkResponse ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 401 {object} map[string]interface{} ""
// @Failure 409 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/posts/{post_id}/bookmark [post]
func (h *InteractionHandler) BookmarkPost(c *gin.Context) {
	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
		return
	}

	postID := c.Param("id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	bookmark, err := h.interactionService.BookmarkPost(userID.(string), postID)
	if err != nil {
		h.logger.Error("Failed to bookmark post", zap.String("user_id", userID.(string)), zap.String("post_id", postID), zap.Error(err))
		if err.Error() == "" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else if err.Error() == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		}
		return
	}

	c.JSON(http.StatusCreated, bookmark.ToResponse())
}

// UnbookmarkPost 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param post_id path string true "ID"
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 401 {object} map[string]interface{} ""
// @Failure 404 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/posts/{post_id}/bookmark [delete]
func (h *InteractionHandler) UnbookmarkPost(c *gin.Context) {
	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
		return
	}

	postID := c.Param("id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	err := h.interactionService.UnbookmarkPost(userID.(string), postID)
	if err != nil {
		h.logger.Error("Failed to unbookmark post", zap.String("user_id", userID.(string)), zap.String("post_id", postID), zap.Error(err))
		if err.Error() == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": ""})
}

// GetMyBookmarks 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param page query int false "" default(1)
// @Param page_size query int false "" default(20)
// @Success 200 {object} map[string]interface{} ""
// @Failure 401 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/bookmarks [get]
func (h *InteractionHandler) GetMyBookmarks(c *gin.Context) {
	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
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

	bookmarks, total, err := h.interactionService.GetUserBookmarks(userID.(string), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get user bookmarks", zap.String("user_id", userID.(string)), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	// 
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

// GetUserFollowers 
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
// @Router /community/users/{user_id}/followers [get]
func (h *InteractionHandler) GetUserFollowers(c *gin.Context) {
	userID := c.Param("id")
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

	followers, total, err := h.interactionService.GetUserFollowers(userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get user followers", zap.String("user_id", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	// 
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

// GetUserFollowing 
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
// @Router /community/users/{user_id}/following [get]
func (h *InteractionHandler) GetUserFollowing(c *gin.Context) {
	userID := c.Param("id")
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

	following, total, err := h.interactionService.GetUserFollowing(userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get user following", zap.String("user_id", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	// 
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

// GetInteractionStats 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Success 200 {object} models.InteractionStatsResponse ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/interactions/stats [get]
func (h *InteractionHandler) GetInteractionStats(c *gin.Context) {
	stats, err := h.interactionService.GetInteractionStats()
	if err != nil {
		h.logger.Error("Failed to get interaction stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// CheckInteractionStatus 黥
// @Summary 黥
// @Description //
// @Tags 
// @Produce json
// @Param type query string true "" Enums(post_like,comment_like,user_follow,post_bookmark)
// @Param target_id query string true "ID"
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 401 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/interactions/status [get]
func (h *InteractionHandler) CheckInteractionStatus(c *gin.Context) {
	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
		return
	}

	interactionType := c.Query("type")
	targetID := c.Query("target_id")

	if interactionType == "" || targetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": ""})
		return
	}

	if err != nil {
		h.logger.Error("Failed to check interaction status",
			zap.String("user_id", userID.(string)),
			zap.String("type", interactionType),
			zap.String("target_id", targetID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"type":      interactionType,
		"target_id": targetID,
		"status":    status,
	})
}

