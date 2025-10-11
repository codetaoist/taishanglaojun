package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/services"
)

// CategoryHandler 分类管理处理�?
type CategoryHandler struct {
	categoryService *services.CategoryService
}

// NewCategoryHandler 创建分类处理器实�?
func NewCategoryHandler(categoryService *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// GetCategories 获取分类列表
// @Summary 获取分类列表
// @Description 获取所有分类，支持层级结构
// @Tags 分类管理
// @Produce json
// @Param parent_id query int false "父分类ID，不传则获取顶级分类"
// @Param include_children query bool false "是否包含子分�? default(false)
// @Success 200 {object} CategoriesResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/categories [get]
func (h *CategoryHandler) GetCategories(c *gin.Context) {
	var parentID *int
	if parentIDStr := c.Query("parent_id"); parentIDStr != "" {
		if id, err := strconv.Atoi(parentIDStr); err == nil {
			parentID = &id
		}
	}

	includeChildren := c.Query("include_children") == "true"

	categories, err := h.categoryService.GetCategories(c.Request.Context(), parentID, includeChildren)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_CATEGORIES_ERROR",
			Message: "获取分类列表失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, CategoriesResponse{
		Code:    200,
		Message: "获取分类列表成功",
		Data:    categories,
	})
}

// GetCategoryByID 获取分类详情
// @Summary 获取分类详情
// @Description 根据ID获取分类详情
// @Tags 分类管理
// @Produce json
// @Param id path int true "分类ID"
// @Success 200 {object} models.Category
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/categories/{id} [get]
func (h *CategoryHandler) GetCategoryByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "分类ID格式错误",
		})
		return
	}

	category, err := h.categoryService.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "CATEGORY_NOT_FOUND",
			Message: "分类不存�?,
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, category)
}

// CreateCategory 创建分类
// @Summary 创建分类
// @Description 创建新的分类
// @Tags 分类管理
// @Accept json
// @Produce json
// @Param request body CreateCategoryRequest true "创建分类请求"
// @Success 201 {object} models.Category
// @Failure 400 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/categories [post]
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数错误",
			Details: err.Error(),
		})
		return
	}

	category := &models.Category{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
		SortOrder:   req.SortOrder,
		IsActive:    true,
	}

	createdCategory, err := h.categoryService.CreateCategory(c.Request.Context(), category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CREATE_CATEGORY_ERROR",
			Message: "创建分类失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, createdCategory)
}

// UpdateCategory 更新分类
// @Summary 更新分类
// @Description 更新分类信息
// @Tags 分类管理
// @Accept json
// @Produce json
// @Param id path int true "分类ID"
// @Param request body UpdateCategoryRequest true "更新分类请求"
// @Success 200 {object} models.Category
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "分类ID格式错误",
		})
		return
	}

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数错误",
			Details: err.Error(),
		})
		return
	}

	updatedCategory, err := h.categoryService.UpdateCategory(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "UPDATE_CATEGORY_ERROR",
			Message: "更新分类失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, updatedCategory)
}

// DeleteCategory 删除分类
// @Summary 删除分类
// @Description 删除分类（软删除�?
// @Tags 分类管理
// @Produce json
// @Param id path int true "分类ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "分类ID格式错误",
		})
		return
	}

	err = h.categoryService.DeleteCategory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "DELETE_CATEGORY_ERROR",
			Message: "删除分类失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "删除分类成功",
	})
}

// GetCategoryStats 获取分类统计
// @Summary 获取分类统计
// @Description 获取分类下的内容统计信息
// @Tags 分类管理
// @Produce json
// @Param id path int true "分类ID"
// @Success 200 {object} CategoryStatsResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/categories/{id}/stats [get]
func (h *CategoryHandler) GetCategoryStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "分类ID格式错误",
		})
		return
	}

	stats, err := h.categoryService.GetCategoryStats(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_STATS_ERROR",
			Message: "获取分类统计失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, CategoryStatsResponse{
		Code:    200,
		Message: "获取分类统计成功",
		Data:    *stats,
	})
}

// 请求和响应结构体
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description" binding:"max=500"`
	ParentID    *int   `json:"parent_id"`
	SortOrder   int    `json:"sort_order"`
}

type UpdateCategoryRequest struct {
	Name        *string `json:"name" binding:"omitempty,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	ParentID    *int    `json:"parent_id"`
	SortOrder   *int    `json:"sort_order"`
	IsActive    *bool   `json:"is_active"`
}

type CategoriesResponse struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    []models.Category  `json:"data"`
}

type CategoryStatsResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    models.CategoryStats   `json:"data"`
}
