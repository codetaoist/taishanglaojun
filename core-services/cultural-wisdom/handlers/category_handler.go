package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/services"
	"github.com/gin-gonic/gin"
)

// CategoryHandler 
type CategoryHandler struct {
	categoryService *services.CategoryService
}

// NewCategoryHandler 
func NewCategoryHandler(categoryService *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// GetCategories 
// @Summary 
// @Description 㼶
// @Tags 
// @Produce json
// @Param parent_id query int false "ID"
// @Param include_children query bool false "" default(false)
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
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, CategoriesResponse{
		Code:    200,
		Message: "",
		Data:    categories,
	})
}

// GetCategoryByID 
// @Summary 
// @Description ID
// @Tags 
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} models.Category
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/categories/{id} [get]
func (h *CategoryHandler) GetCategoryByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "ID",
		})
		return
	}

	category, err := h.categoryService.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "CATEGORY_NOT_FOUND",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, category)
}

// CreateCategory 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param request body CreateCategoryRequest true ""
// @Success 201 {object} models.Category
// @Failure 400 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/categories [post]
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "",
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
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, createdCategory)
}

// UpdateCategory 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param request body UpdateCategoryRequest true ""
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
			Message: "ID",
		})
		return
	}

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	updatedCategory, err := h.categoryService.UpdateCategory(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "UPDATE_CATEGORY_ERROR",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, updatedCategory)
}

// DeleteCategory 
// @Summary 
// @Description IsActivefalse
// @Tags 
// @Produce json
// @Param id path int true "ID"
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
			Message: "ID",
		})
		return
	}

	err = h.categoryService.DeleteCategory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "DELETE_CATEGORY_ERROR",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "",
	})
}

// GetCategoryStats 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} CategoryStatsResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/categories/{id}/stats [get]
func (h *CategoryHandler) GetCategoryStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "ID",
		})
		return
	}

	stats, err := h.categoryService.GetCategoryStats(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_STATS_ERROR",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, CategoryStatsResponse{
		Code:    200,
		Message: "",
		Data:    *stats,
	})
}

// 
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
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    []models.Category `json:"data"`
}

type CategoryStatsResponse struct {
	Code    int                  `json:"code"`
	Message string               `json:"message"`
	Data    models.CategoryStats `json:"data"`
}

