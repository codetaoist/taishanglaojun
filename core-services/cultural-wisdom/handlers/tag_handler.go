package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/services"
)

// TagHandler 标签管理处理器
type TagHandler struct {
	tagService *services.TagService
}

// NewTagHandler 创建标签处理器实例
func NewTagHandler(tagService *services.TagService) *TagHandler {
	return &TagHandler{
		tagService: tagService,
	}
}

// GetTags 获取标签列表
// @Summary 获取标签列表
// @Description 获取所有标签，支持分页和搜索
// @Tags 标签管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(20)
// @Param search query string false "搜索关键词"
// @Param sort_by query string false "排序字段" Enums(name,usage_count,created_at) default(usage_count)
// @Param sort_order query string false "排序方向" Enums(asc,desc) default(desc)
// @Success 200 {object} TagsResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/tags [get]
func (h *TagHandler) GetTags(c *gin.Context) {
	page := 1
	size := 20
	
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if sizeStr := c.Query("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			size = s
		}
	}
	
	search := c.Query("search")
	sortBy := c.DefaultQuery("sort_by", "usage_count")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	
	tags, total, err := h.tagService.GetTags(c.Request.Context(), page, size, search, sortBy, sortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_TAGS_ERROR",
			Message: "获取标签列表失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, TagsResponse{
		Code:    200,
		Message: "获取标签列表成功",
		Data:    tags,
		Page:    page,
		Size:    size,
		Total:   total,
	})
}

// GetTagByID 获取标签详情
// @Summary 获取标签详情
// @Description 根据ID获取标签详情
// @Tags 标签管理
// @Produce json
// @Param id path int true "标签ID"
// @Success 200 {object} models.WisdomTag
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/tags/{id} [get]
func (h *TagHandler) GetTagByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "标签ID格式错误",
		})
		return
	}

	tag, err := h.tagService.GetTagByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "TAG_NOT_FOUND",
			Message: "标签不存在",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, tag)
}

// CreateTag 创建标签
// @Summary 创建标签
// @Description 创建新的标签
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param request body CreateTagRequest true "创建标签请求"
// @Success 201 {object} models.WisdomTag
// @Failure 400 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/tags [post]
func (h *TagHandler) CreateTag(c *gin.Context) {
	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数错误",
			Details: err.Error(),
		})
		return
	}

	tag := &models.WisdomTag{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    true,
	}

	createdTag, err := h.tagService.CreateTag(c.Request.Context(), tag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CREATE_TAG_ERROR",
			Message: "创建标签失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, createdTag)
}

// UpdateTag 更新标签
// @Summary 更新标签
// @Description 更新标签信息
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param id path int true "标签ID"
// @Param request body UpdateTagRequest true "更新标签请求"
// @Success 200 {object} models.WisdomTag
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/tags/{id} [put]
func (h *TagHandler) UpdateTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "标签ID格式错误",
		})
		return
	}

	var req UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数错误",
			Details: err.Error(),
		})
		return
	}

	updatedTag, err := h.tagService.UpdateTag(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "UPDATE_TAG_ERROR",
			Message: "更新标签失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, updatedTag)
}

// DeleteTag 删除标签
// @Summary 删除标签
// @Description 删除标签（软删除）
// @Tags 标签管理
// @Produce json
// @Param id path int true "标签ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/tags/{id} [delete]
func (h *TagHandler) DeleteTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "标签ID格式错误",
		})
		return
	}

	err = h.tagService.DeleteTag(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "DELETE_TAG_ERROR",
			Message: "删除标签失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "删除标签成功",
	})
}

// GetPopularTags 获取热门标签
// @Summary 获取热门标签
// @Description 获取使用频率最高的标签
// @Tags 标签管理
// @Produce json
// @Param limit query int false "返回数量" default(10)
// @Success 200 {object} TagsResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/tags/popular [get]
func (h *TagHandler) GetPopularTags(c *gin.Context) {
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	tags, err := h.tagService.GetPopularTags(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_POPULAR_TAGS_ERROR",
			Message: "获取热门标签失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, TagsResponse{
		Code:    200,
		Message: "获取热门标签成功",
		Data:    tags,
		Total:   int64(len(tags)),
	})
}

// GetTagStats 获取标签统计
// @Summary 获取标签统计
// @Description 获取标签的使用统计信息
// @Tags 标签管理
// @Produce json
// @Param id path int true "标签ID"
// @Success 200 {object} TagStatsResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/tags/{id}/stats [get]
func (h *TagHandler) GetTagStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "标签ID格式错误",
		})
		return
	}

	stats, err := h.tagService.GetTagStats(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_TAG_STATS_ERROR",
			Message: "获取标签统计失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, TagStatsResponse{
		Code:    200,
		Message: "获取标签统计成功",
		Data:    *stats,
	})
}

// 请求和响应结构体
type CreateTagRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description" binding:"max=500"`
}

type UpdateTagRequest struct {
	Name        *string `json:"name" binding:"omitempty,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	IsActive    *bool   `json:"is_active"`
}

type TagsResponse struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    []models.WisdomTag `json:"data"`
	Page    int                `json:"page,omitempty"`
	Size    int                `json:"size,omitempty"`
	Total   int64              `json:"total"`
}

type TagStatsResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    models.TagStats   `json:"data"`
}