package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/taishanglaojun/core-services/community/models"
	"github.com/codetaoist/taishanglaojun/core-services/community/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PostHandler 
type PostHandler struct {
	postService *services.PostService
	logger      *zap.Logger
}

// NewPostHandler 
func NewPostHandler(postService *services.PostService, logger *zap.Logger) *PostHandler {
	return &PostHandler{
		postService: postService,
		logger:      logger,
	}
}

// CreatePost 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param request body models.CreatePostRequest true ""
// @Success 201 {object} models.PostResponse ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 401 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/posts [post]
func (h *PostHandler) CreatePost(c *gin.Context) {
	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
		return
	}

	var req models.PostCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "", "details": err.Error()})
		return
	}

	// 
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": ""})
		return
	}
	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": ""})
		return
	}

	// 
	post, err := h.postService.CreatePost(userID.(string), &req)
	if err != nil {
		h.logger.Error("Failed to create post", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, post.ToResponse())
}

// GetPost 
// @Summary 
// @Description ID
// @Tags 
// @Produce json
// @Param id path string true "ID"
// @Success 200 {object} models.PostResponse ""
// @Failure 404 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/posts/{id} [get]
func (h *PostHandler) GetPost(c *gin.Context) {
	postID := c.Param("id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	// ID
	var userID *string
	if uid, exists := c.Get("user_id"); exists {
		if uidStr, ok := uid.(string); ok {
			userID = &uidStr
		}
	}

	// 
	post, err := h.postService.GetPost(postID, userID)
	if err != nil {
		h.logger.Error("Failed to get post", zap.String("post_id", postID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": ""})
		return
	}

	c.JSON(http.StatusOK, post.ToResponse())
}

// GetPosts 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param page query int false "" default(1)
// @Param page_size query int false "" default(20)
// @Param category query string false ""
// @Param tag query string false ""
// @Param status query string false "" Enums(active,hidden,deleted)
// @Param sort query string false "" Enums(latest,hot,top)
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/posts [get]
func (h *PostHandler) GetPosts(c *gin.Context) {
	// 
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	category := c.Query("category")
	tag := c.Query("tag")
	authorID := c.Query("author_id")
	status := c.Query("status")
	sort := c.DefaultQuery("sort", "latest")
	keyword := c.Query("keyword")

	// 
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 
	req := &models.PostListRequest{
		Page:     page,
		PageSize: pageSize,
		Category: category,
		Tag:      tag,
		AuthorID: authorID,
		Status:   status,
		SortBy:   sort,
		Keyword:  keyword,
	}

	// ID
	var userID *string
	if uid, exists := c.Get("user_id"); exists {
		if uidStr, ok := uid.(string); ok {
			userID = &uidStr
		}
	}

	// 
	response, err := h.postService.GetPosts(req, userID)
	if err != nil {
		h.logger.Error("Failed to get posts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": response.Posts,
		"pagination": gin.H{
			"page":        response.Page,
			"page_size":   response.PageSize,
			"total":       response.Total,
			"total_pages": response.TotalPages,
		},
	})
}

// UpdatePost 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param request body models.UpdatePostRequest true ""
// @Success 200 {object} models.PostResponse ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 401 {object} map[string]interface{} ""
// @Failure 403 {object} map[string]interface{} ""
// @Failure 404 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/posts/{id} [put]
func (h *PostHandler) UpdatePost(c *gin.Context) {
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

	var req models.PostUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "", "details": err.Error()})
		return
	}

	// 
	post, err := h.postService.UpdatePost(postID, userID.(string), &req)
	if err != nil {
		h.logger.Error("Failed to update post", zap.String("post_id", postID), zap.Error(err))
		if err.Error() == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		}
		return
	}

	c.JSON(http.StatusOK, post.ToResponse())
}

// DeletePost 
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
// @Router /community/posts/{id} [delete]
func (h *PostHandler) DeletePost(c *gin.Context) {
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

	// 
	err := h.postService.DeletePost(postID, userID.(string))
	if err != nil {
		h.logger.Error("Failed to delete post", zap.String("post_id", postID), zap.Error(err))
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

// GetPostStats 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Success 200 {object} models.PostStatsResponse ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/posts/stats [get]
func (h *PostHandler) GetPostStats(c *gin.Context) {
	stats, err := h.postService.GetPostStats()
	if err != nil {
		h.logger.Error("Failed to get post stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// SetPostSticky 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param request body map[string]bool true "" example({"sticky": true})
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 401 {object} map[string]interface{} ""
// @Failure 403 {object} map[string]interface{} ""
// @Failure 404 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/posts/{id}/sticky [patch]
func (h *PostHandler) SetPostSticky(c *gin.Context) {
	postID := c.Param("id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	var req struct {
		Sticky bool `json:"sticky"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": ""})
		return
	}

	err := h.postService.SetPostSticky(postID, req.Sticky)
	if err != nil {
		h.logger.Error("Failed to set post sticky", zap.String("post_id", postID), zap.Error(err))
		if err.Error() == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		}
		return
	}

	message := ""
	if req.Sticky {
		message = ""
	}
	c.JSON(http.StatusOK, gin.H{"message": message})
}

// SetPostHot 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param request body map[string]bool true "" example({"hot": true})
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 401 {object} map[string]interface{} ""
// @Failure 403 {object} map[string]interface{} ""
// @Failure 404 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/posts/{id}/hot [patch]
func (h *PostHandler) SetPostHot(c *gin.Context) {
	postID := c.Param("id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	var req struct {
		Hot bool `json:"hot"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": ""})
		return
	}

	err := h.postService.SetPostHot(postID, req.Hot)
	if err != nil {
		h.logger.Error("Failed to set post hot", zap.String("post_id", postID), zap.Error(err))
		if err.Error() == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		}
		return
	}

	message := ""
	if req.Hot {
		message = ""
	}
	c.JSON(http.StatusOK, gin.H{"message": message})
}

// SearchPosts 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param q query string true ""
// @Param page query int false "" default(1)
// @Param page_size query int false "" default(20)
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/posts/search [get]
func (h *PostHandler) SearchPosts(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": ""})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	response, err := h.postService.SearchPosts(keyword, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to search posts", zap.String("keyword", keyword), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": response.Posts,
		"pagination": gin.H{
			"page":        response.Page,
			"page_size":   response.PageSize,
			"total":       response.Total,
			"total_pages": response.TotalPages,
		},
		"keyword": keyword,
	})
}

