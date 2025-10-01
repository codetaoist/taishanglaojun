package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/services"
)

// WisdomHandler 文化智慧API处理�?type WisdomHandler struct {
	wisdomService *services.WisdomService
}

// NewWisdomHandler 创建智慧处理器实�?func NewWisdomHandler(wisdomService *services.WisdomService) *WisdomHandler {
	return &WisdomHandler{
		wisdomService: wisdomService,
	}
}

// GetWisdomList 获取智慧内容列表
// @Summary 获取智慧内容列表
// @Description 根据条件获取文化智慧内容列表
// @Tags 文化智慧
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(20)
// @Param category query string false "分类ID"
// @Param school query string false "学派" Enums(�?�?�?�?
// @Param tags query string false "标签，多个用逗号分隔"
// @Param difficulty query string false "难度等级，多个用逗号分隔"
// @Param search query string false "搜索关键�?
// @Param sort_by query string false "排序字段" Enums(created_at,updated_at,view_count,like_count) default(created_at)
// @Param sort_order query string false "排序方向" Enums(asc,desc) default(desc)
// @Success 200 {object} WisdomListResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/wisdom [get]
func (h *WisdomHandler) GetWisdomList(c *gin.Context) {
	filter := h.buildFilterFromQuery(c)
	
	wisdomList, total, err := h.wisdomService.GetWisdomList(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_WISDOM_LIST_ERROR",
			Message: "获取智慧内容列表失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, WisdomListResponse{
		Data:  wisdomList,
		Page:  filter.Page,
		Size:  filter.Size,
		Total: total,
	})
}

// GetWisdomDetail 获取智慧内容详情
// @Summary 获取智慧内容详情
// @Description 根据ID获取文化智慧内容详情
// @Tags 文化智慧
// @Produce json
// @Param id path string true "智慧内容ID"
// @Success 200 {object} models.CulturalWisdom
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/wisdom/{id} [get]
func (h *WisdomHandler) GetWisdomDetail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "智慧内容ID不能为空",
		})
		return
	}

	wisdom, err := h.wisdomService.GetWisdomByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "WISDOM_NOT_FOUND",
			Message: "智慧内容不存�?,
			Details: err.Error(),
		})
		return
	}

	// 增加浏览次数
	go h.wisdomService.IncrementViewCount(c.Request.Context(), id)

	c.JSON(http.StatusOK, wisdom)
}

// CreateWisdom 创建智慧内容
// @Summary 创建智慧内容
// @Description 创建新的文化智慧内容
// @Tags 文化智慧
// @Accept json
// @Produce json
// @Param request body CreateWisdomRequest true "创建请求"
// @Success 201 {object} models.CulturalWisdom
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/wisdom [post]
func (h *WisdomHandler) CreateWisdom(c *gin.Context) {
	var req CreateWisdomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "用户未认�?,
		})
		return
	}

	// 检查用户权限（假设需要L3及以上等级）
	userLevel, _ := c.Get("user_level")
	if userLevel == nil || userLevel.(int) < 3 {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Code:    "INSUFFICIENT_PERMISSION",
			Message: "权限不足，需要L3及以上等�?,
		})
		return
	}

	wisdom := &models.CulturalWisdom{
		Title:      req.Title,
		Content:    req.Content,
		Summary:    req.Summary,
		Category:   req.Category,
		Tags:       req.Tags,
		Source:     req.Source,
		Difficulty: req.Difficulty,
		Status:     "draft",
		AuthorID:   userID.(string),
	}

	createdWisdom, err := h.wisdomService.CreateWisdom(c.Request.Context(), wisdom)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CREATE_WISDOM_ERROR",
			Message: "创建智慧内容失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, createdWisdom)
}

// UpdateWisdom 更新智慧内容
// @Summary 更新智慧内容
// @Description 更新指定的文化智慧内�?// @Tags 文化智慧
// @Accept json
// @Produce json
// @Param id path string true "智慧内容ID"
// @Param request body UpdateWisdomRequest true "更新请求"
// @Success 200 {object} models.CulturalWisdom
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/wisdom/{id} [put]
func (h *WisdomHandler) UpdateWisdom(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "智慧内容ID不能为空",
		})
		return
	}

	var req UpdateWisdomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "用户未认�?,
		})
		return
	}

	updatedWisdom, err := h.wisdomService.UpdateWisdom(c.Request.Context(), id, userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "UPDATE_WISDOM_ERROR",
			Message: "更新智慧内容失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, updatedWisdom)
}

// DeleteWisdom 删除智慧内容
// @Summary 删除智慧内容
// @Description 删除指定的文化智慧内�?// @Tags 文化智慧
// @Param id path string true "智慧内容ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/wisdom/{id} [delete]
func (h *WisdomHandler) DeleteWisdom(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "智慧内容ID不能为空",
		})
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "用户未认�?,
		})
		return
	}

	err := h.wisdomService.DeleteWisdom(c.Request.Context(), id, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "DELETE_WISDOM_ERROR",
			Message: "删除智慧内容失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "智慧内容删除成功",
	})
}

// buildFilterFromQuery 从查询参数构建过滤条�?func (h *WisdomHandler) buildFilterFromQuery(c *gin.Context) *models.WisdomFilter {
	filter := &models.WisdomFilter{
		Page: 1,
		Size: 20,
		SortBy: "created_at",
		SortOrder: "desc",
	}

	// 页码
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}

	// 每页数量
	if sizeStr := c.Query("size"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 && size <= 100 {
			filter.Size = size
		}
	}

	// 分类
	filter.CategoryID = c.Query("category")
	
	// 学派
	filter.School = c.Query("school")
	
	// 搜索关键�?	filter.SearchQuery = c.Query("search")
	
	// 排序字段
	if sortBy := c.Query("sort_by"); sortBy != "" {
		filter.SortBy = sortBy
	}
	
	// 排序方向
	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		filter.SortOrder = sortOrder
	}

	return filter
}

// 请求和响应结构体
type CreateWisdomRequest struct {
	Title      string         `json:"title" binding:"required"`
	Content    string         `json:"content" binding:"required"`
	Summary    string         `json:"summary"`
	Category   models.Category `json:"category" binding:"required"`
	Tags       []string       `json:"tags"`
	Source     models.Source  `json:"source"`
	Difficulty int            `json:"difficulty" binding:"min=1,max=9"`
}

type UpdateWisdomRequest struct {
	Title      *string         `json:"title"`
	Content    *string         `json:"content"`
	Summary    *string         `json:"summary"`
	Category   *models.Category `json:"category"`
	Tags       []string        `json:"tags"`
	Source     *models.Source  `json:"source"`
	Difficulty *int            `json:"difficulty"`
	Status     *string         `json:"status"`
}

type WisdomListResponse struct {
	Data  []*models.WisdomSummary `json:"data"`
	Page  int                     `json:"page"`
	Size  int                     `json:"size"`
	Total int64                   `json:"total"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
