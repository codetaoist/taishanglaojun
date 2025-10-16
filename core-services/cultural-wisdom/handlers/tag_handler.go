package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/services"
	"github.com/gin-gonic/gin"
)

// TagHandler 
type TagHandler struct {
	tagService *services.TagService
}

// NewTagHandler 
func NewTagHandler(tagService *services.TagService) *TagHandler {
	return &TagHandler{
		tagService: tagService,
	}
}

// GetTags 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param page query int false "" default(1)
// @Param size query int false "" default(20)
// @Param search query string false ""
// @Param sort_by query string false "" Enums(name,usage_count,created_at) default(usage_count)
// @Param sort_order query string false "" Enums(asc,desc) default(desc)
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
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, TagsResponse{
		Code:    200,
		Message: "",
		Data:    tags,
		Page:    page,
		Size:    size,
		Total:   total,
	})
}

// GetTagByID 
// @Summary 
// @Description ID
// @Tags 
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} models.WisdomTag
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/tags/{id} [get]
func (h *TagHandler) GetTagByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "ID",
		})
		return
	}

	tag, err := h.tagService.GetTagByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "TAG_NOT_FOUND",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, tag)
}

// CreateTag 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param request body CreateTagRequest true ""
// @Success 201 {object} models.WisdomTag
// @Failure 400 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/tags [post]
func (h *TagHandler) CreateTag(c *gin.Context) {
	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "",
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
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, createdTag)
}

// UpdateTag 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param request body UpdateTagRequest true ""
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
			Message: "ID",
		})
		return
	}

	var req UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	updatedTag, err := h.tagService.UpdateTag(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "UPDATE_TAG_ERROR",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, updatedTag)
}

// DeleteTag 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param id path int true "ID"
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
			Message: "ID",
		})
		return
	}

	err = h.tagService.DeleteTag(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "DELETE_TAG_ERROR",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "",
	})
}

// GetPopularTags 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param limit query int false "" default(10)
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
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, TagsResponse{
		Code:    200,
		Message: "",
		Data:    tags,
		Total:   int64(len(tags)),
	})
}

// GetTagStats 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} TagStatsResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/tags/{id}/stats [get]
func (h *TagHandler) GetTagStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "ID",
		})
		return
	}

	stats, err := h.tagService.GetTagStats(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_TAG_STATS_ERROR",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, TagStatsResponse{
		Code:    200,
		Message: "",
		Data:    *stats,
	})
}

// 
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
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    models.TagStats `json:"data"`
}

