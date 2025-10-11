package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/taishanglaojun/core-services/community/models"
	"github.com/codetaoist/taishanglaojun/core-services/community/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PostHandler 帖子处理�?
type PostHandler struct {
	postService *services.PostService
	logger      *zap.Logger
}

// NewPostHandler 创建帖子处理器实�?
func NewPostHandler(postService *services.PostService, logger *zap.Logger) *PostHandler {
	return &PostHandler{
		postService: postService,
		logger:      logger,
	}
}

// CreatePost 创建帖子
// @Summary 创建帖子
// @Description 创建新的社区帖子
// @Tags 帖子管理
// @Accept json
// @Produce json
// @Param request body models.CreatePostRequest true "创建帖子请求"
// @Success 201 {object} models.PostResponse "创建成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授�?
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /community/posts [post]
func (h *PostHandler) CreatePost(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访�?})
		return
	}

	var req models.PostCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 验证请求参数
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "标题不能为空"})
		return
	}
	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "内容不能为空"})
		return
	}

	// 创建帖子
	post, err := h.postService.CreatePost(userID.(string), &req)
	if err != nil {
		h.logger.Error("Failed to create post", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建帖子失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, post.ToResponse())
}

// GetPost 获取帖子详情
// @Summary 获取帖子详情
// @Description 根据ID获取帖子详细信息
// @Tags 帖子管理
// @Produce json
// @Param id path string true "帖子ID"
// @Success 200 {object} models.PostResponse "获取成功"
// @Failure 404 {object} map[string]interface{} "帖子不存�?
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /community/posts/{id} [get]
func (h *PostHandler) GetPost(c *gin.Context) {
	postID := c.Param("id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID不能为空"})
		return
	}

	// 获取当前用户ID（如果已登录�?
	var userID *string
	if uid, exists := c.Get("user_id"); exists {
		if uidStr, ok := uid.(string); ok {
			userID = &uidStr
		}
	}

	// 获取帖子详情
	post, err := h.postService.GetPost(postID, userID)
	if err != nil {
		h.logger.Error("Failed to get post", zap.String("post_id", postID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存�?})
		return
	}

	c.JSON(http.StatusOK, post.ToResponse())
}

// GetPosts 获取帖子列表
// @Summary 获取帖子列表
// @Description 分页获取帖子列表，支持按分类、标签、状态筛�?
// @Tags 帖子管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param category query string false "分类"
// @Param tag query string false "标签"
// @Param status query string false "状�?
// @Param sort query string false "排序方式" Enums(latest,hot,top)
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /community/posts [get]
func (h *PostHandler) GetPosts(c *gin.Context) {
	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	category := c.Query("category")
	tag := c.Query("tag")
	authorID := c.Query("author_id")
	status := c.Query("status")
	sort := c.DefaultQuery("sort", "latest")
	keyword := c.Query("keyword")

	// 验证参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 构建请求参数
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

	// 获取当前用户ID（如果已登录�?
	var userID *string
	if uid, exists := c.Get("user_id"); exists {
		if uidStr, ok := uid.(string); ok {
			userID = &uidStr
		}
	}

	// 获取帖子列表
	response, err := h.postService.GetPosts(req, userID)
	if err != nil {
		h.logger.Error("Failed to get posts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取帖子列表失败"})
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

// UpdatePost 更新帖子
// @Summary 更新帖子
// @Description 更新帖子信息（仅作者可操作�?
// @Tags 帖子管理
// @Accept json
// @Produce json
// @Param id path string true "帖子ID"
// @Param request body models.UpdatePostRequest true "更新帖子请求"
// @Success 200 {object} models.PostResponse "更新成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授�?
// @Failure 403 {object} map[string]interface{} "无权�?
// @Failure 404 {object} map[string]interface{} "帖子不存�?
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /community/posts/{id} [put]
func (h *PostHandler) UpdatePost(c *gin.Context) {
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

	var req models.PostUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 更新帖子
	post, err := h.postService.UpdatePost(postID, userID.(string), &req)
	if err != nil {
		h.logger.Error("Failed to update post", zap.String("post_id", postID), zap.Error(err))
		if err.Error() == "帖子不存�? {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == "无权限操�? {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新帖子失败"})
		}
		return
	}

	c.JSON(http.StatusOK, post.ToResponse())
}

// DeletePost 删除帖子
// @Summary 删除帖子
// @Description 删除帖子（仅作者可操作�?
// @Tags 帖子管理
// @Produce json
// @Param id path string true "帖子ID"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 401 {object} map[string]interface{} "未授�?
// @Failure 403 {object} map[string]interface{} "无权�?
// @Failure 404 {object} map[string]interface{} "帖子不存�?
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /community/posts/{id} [delete]
func (h *PostHandler) DeletePost(c *gin.Context) {
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

	// 删除帖子
	err := h.postService.DeletePost(postID, userID.(string))
	if err != nil {
		h.logger.Error("Failed to delete post", zap.String("post_id", postID), zap.Error(err))
		if err.Error() == "帖子不存�? {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == "无权限操�? {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除帖子失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetPostStats 获取帖子统计
// @Summary 获取帖子统计
// @Description 获取帖子相关统计信息
// @Tags 帖子管理
// @Produce json
// @Success 200 {object} models.PostStatsResponse "获取成功"
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /community/posts/stats [get]
func (h *PostHandler) GetPostStats(c *gin.Context) {
	stats, err := h.postService.GetPostStats()
	if err != nil {
		h.logger.Error("Failed to get post stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计信息失败"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// SetPostSticky 设置帖子置顶
// @Summary 设置帖子置顶
// @Description 设置或取消帖子置顶状态（管理员操作）
// @Tags 帖子管理
// @Accept json
// @Produce json
// @Param id path string true "帖子ID"
// @Param request body map[string]bool true "置顶状�? example({"sticky": true})
// @Success 200 {object} map[string]interface{} "设置成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授�?
// @Failure 404 {object} map[string]interface{} "帖子不存�?
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /community/posts/{id}/sticky [patch]
func (h *PostHandler) SetPostSticky(c *gin.Context) {
	postID := c.Param("id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID不能为空"})
		return
	}

	var req struct {
		Sticky bool `json:"sticky"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	err := h.postService.SetPostSticky(postID, req.Sticky)
	if err != nil {
		h.logger.Error("Failed to set post sticky", zap.String("post_id", postID), zap.Error(err))
		if err.Error() == "帖子不存�? {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "设置置顶失败"})
		}
		return
	}

	message := "取消置顶成功"
	if req.Sticky {
		message = "设置置顶成功"
	}
	c.JSON(http.StatusOK, gin.H{"message": message})
}

// SetPostHot 设置帖子热门
// @Summary 设置帖子热门
// @Description 设置或取消帖子热门状态（管理员操作）
// @Tags 帖子管理
// @Accept json
// @Produce json
// @Param id path string true "帖子ID"
// @Param request body map[string]bool true "热门状�? example({"hot": true})
// @Success 200 {object} map[string]interface{} "设置成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授�?
// @Failure 404 {object} map[string]interface{} "帖子不存�?
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /community/posts/{id}/hot [patch]
func (h *PostHandler) SetPostHot(c *gin.Context) {
	postID := c.Param("id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID不能为空"})
		return
	}

	var req struct {
		Hot bool `json:"hot"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	err := h.postService.SetPostHot(postID, req.Hot)
	if err != nil {
		h.logger.Error("Failed to set post hot", zap.String("post_id", postID), zap.Error(err))
		if err.Error() == "帖子不存�? {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "设置热门失败"})
		}
		return
	}

	message := "取消热门成功"
	if req.Hot {
		message = "设置热门成功"
	}
	c.JSON(http.StatusOK, gin.H{"message": message})
}

// SearchPosts 搜索帖子
// @Summary 搜索帖子
// @Description 根据关键词搜索帖�?
// @Tags 帖子管理
// @Produce json
// @Param q query string true "搜索关键�?
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{} "搜索成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /community/posts/search [get]
func (h *PostHandler) SearchPosts(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为�?})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 验证参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	response, err := h.postService.SearchPosts(keyword, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to search posts", zap.String("keyword", keyword), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索失败"})
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
