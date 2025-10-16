package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/services"
	"github.com/gin-gonic/gin"
)

// WisdomHandler API
type WisdomHandler struct {
	wisdomService *services.WisdomService
}

// NewWisdomHandler 
func NewWisdomHandler(wisdomService *services.WisdomService) *WisdomHandler {
	return &WisdomHandler{
		wisdomService: wisdomService,
	}
}

// GetWisdomList 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param page query int false "" default(1)
// @Param size query int false "" default(20)
// @Param category query string false "ID"
// @Param school query string false "" Enums(,,)
// @Param tags query string false ""
// @Param difficulty query string false ""
// @Param search query string false ""
// @Param sort_by query string false "" Enums(created_at,updated_at,view_count,like_count) default(created_at)
// @Param sort_order query string false "" Enums(asc,desc) default(desc)
// @Success 200 {object} WisdomListResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/wisdom [get]
func (h *WisdomHandler) GetWisdomList(c *gin.Context) {
	filter := h.buildFilterFromQuery(c)

	wisdomList, total, err := h.wisdomService.GetWisdomList(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_WISDOM_LIST_ERROR",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, WisdomListResponse{
		Data:  convertToWisdomSummaryPointers(wisdomList),
		Page:  filter.Page,
		Size:  filter.Size,
		Total: total,
	})
}

// GetWisdomDetail 
// @Summary 
// @Description ID
// @Tags 
// @Produce json
// @Param id path string true "ID"
// @Success 200 {object} models.CulturalWisdom
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/wisdom/{id} [get]
func (h *WisdomHandler) GetWisdomDetail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "ID",
		})
		return
	}

	wisdom, err := h.wisdomService.GetWisdomByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "WISDOM_NOT_FOUND",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	// 
	go h.wisdomService.IncrementViewCount(c.Request.Context(), id)

	c.JSON(http.StatusOK, wisdom)
}

// CreateWisdom 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param request body CreateWisdomRequest true ""
// @Success 201 {object} models.CulturalWisdom
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/wisdom [post]
func (h *WisdomHandler) CreateWisdom(c *gin.Context) {
	var req CreateWisdomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "",
		})
		return
	}

	// L3
	userLevel, _ := c.Get("user_level")
	if userLevel == nil || userLevel.(int) < 3 {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Code:    "INSUFFICIENT_PERMISSION",
			Message: "L3",
		})
		return
	}

	wisdom := &models.CulturalWisdom{
		Title:      req.Title,
		Content:    req.Content,
		Summary:    req.Summary,
		Category:   req.Category.Name,
		Tags:       models.StringSlice(req.Tags),
		Difficulty: strconv.Itoa(req.Difficulty),
		Status:     "draft",
		AuthorID:   userID.(string),
	}

	createdWisdom, err := h.wisdomService.CreateWisdom(c.Request.Context(), wisdom)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CREATE_WISDOM_ERROR",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, createdWisdom)
}

// UpdateWisdom 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param request body UpdateWisdomRequest true ""
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
			Message: "ID",
		})
		return
	}

	var req UpdateWisdomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "",
		})
		return
	}

	// models.UpdateWisdomRequest
	modelReq := &models.UpdateWisdomRequest{}

	if req.Title != nil {
		modelReq.Title = *req.Title
	}

	if req.Content != nil {
		modelReq.Content = *req.Content
	}

	if req.Summary != nil {
		modelReq.Summary = *req.Summary
	}

	if req.Status != nil {
		modelReq.Status = *req.Status
	}

	modelReq.Tags = req.Tags

	if req.Category != nil {
		modelReq.CategoryID = req.Category.Name
	}

	if req.Difficulty != nil {
		modelReq.Difficulty = *req.Difficulty
	}

	updatedWisdom, err := h.wisdomService.UpdateWisdom(c.Request.Context(), id, modelReq, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "UPDATE_WISDOM_ERROR",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, updatedWisdom)
}

// DeleteWisdom 
// @Summary 
// @Description 
// @Tags 
// @Param id path string true "ID"
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
			Message: "ID",
		})
		return
	}

	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "",
		})
		return
	}

	err := h.wisdomService.DeleteWisdom(c.Request.Context(), id, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "DELETE_WISDOM_ERROR",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "",
	})
}

// GetWisdomStats 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Success 200 {object} models.WisdomStats
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/wisdom/stats [get]
func (h *WisdomHandler) GetWisdomStats(c *gin.Context) {
	stats, err := h.wisdomService.GetWisdomStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_WISDOM_STATS_ERROR",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// buildFilterFromQuery 
func (h *WisdomHandler) buildFilterFromQuery(c *gin.Context) *models.WisdomFilter {
	filter := &models.WisdomFilter{
		Page:      1,
		Size:      20,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	// 
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}

	// 
	if sizeStr := c.Query("size"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 && size <= 100 {
			filter.Size = size
		}
	}

	// 
	filter.CategoryID = c.Query("category")

	// 
	filter.School = c.Query("school")

	// 
	filter.SearchQuery = c.Query("search")

	// 
	if sortBy := c.Query("sort_by"); sortBy != "" {
		filter.SortBy = sortBy
	}

	// 
	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		filter.SortOrder = sortOrder
	}

	return filter
}

// 
type CreateWisdomRequest struct {
	Title      string          `json:"title" binding:"required"`
	Content    string          `json:"content" binding:"required"`
	Summary    string          `json:"summary"`
	Category   models.Category `json:"category" binding:"required"`
	Tags       []string        `json:"tags"`
	Source     models.Source   `json:"source"`
	Difficulty int             `json:"difficulty" binding:"min=1,max=9"`
}

type UpdateWisdomRequest struct {
	Title      *string          `json:"title"`
	Content    *string          `json:"content"`
	Summary    *string          `json:"summary"`
	Category   *models.Category `json:"category"`
	Tags       []string         `json:"tags"`
	Source     *models.Source   `json:"source"`
	Difficulty *int             `json:"difficulty"`
	Status     *string          `json:"status"`
}

type WisdomListResponse struct {
	Data  []*models.WisdomSummary `json:"data"`
	Page  int                     `json:"page"`
	Size  int                     `json:"size"`
	Total int64                   `json:"total"`
}

// convertToWisdomSummaryPointers 
func convertToWisdomSummaryPointers(summaries []models.WisdomSummary) []*models.WisdomSummary {
	result := make([]*models.WisdomSummary, len(summaries))
	for i := range summaries {
		result[i] = &summaries[i]
	}
	return result
}

// BatchDeleteWisdom 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param request body BatchDeleteRequest true ""
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/cultural-wisdom/batch-delete [post]
func (h *WisdomHandler) BatchDeleteWisdom(c *gin.Context) {
	var req BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "EMPTY_IDS",
			Message: "ID",
		})
		return
	}

	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "",
		})
		return
	}

	// 
	deletedCount, err := h.wisdomService.BatchDeleteWisdom(c.Request.Context(), req.IDs, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "BATCH_DELETE_ERROR",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "",
		"deletedCount": deletedCount,
	})
}

// AdvancedSearchWisdom 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param keyword query string false ""
// @Param category query string false ""
// @Param school query string false ""
// @Param author query string false ""
// @Param tags query []string false ""
// @Param difficulty query []string false ""
// @Param dateRange query []string false ""
// @Param status query string false ""
// @Param page query int false "" default(1)
// @Param pageSize query int false "" default(10)
// @Param sortBy query string false "" default("created_at")
// @Param sortOrder query string false "" default("desc")
// @Success 200 {object} WisdomListResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/cultural-wisdom/advanced-search [get]
func (h *WisdomHandler) AdvancedSearchWisdom(c *gin.Context) {
	// 
	filter := buildAdvancedFilterFromQuery(c)

	// 
	wisdoms, total, err := h.wisdomService.AdvancedSearchWisdom(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "SEARCH_ERROR",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	// 
	wisdomSummaries := convertToWisdomSummaryPointers(wisdoms)

	c.JSON(http.StatusOK, WisdomListResponse{
		Data:  wisdomSummaries,
		Total: total,
		Page:  filter.Page,
		Size:  filter.Size,
	})
}

// BatchDeleteRequest 
type BatchDeleteRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

// buildAdvancedFilterFromQuery 
func buildAdvancedFilterFromQuery(c *gin.Context) *models.WisdomFilter {
	filter := &models.WisdomFilter{
		Page:      1,
		Size:      10,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	// 
	if keyword := c.Query("keyword"); keyword != "" {
		filter.SearchQuery = keyword
	}
	if category := c.Query("category"); category != "" {
		filter.CategoryID = category
	}
	if school := c.Query("school"); school != "" {
		filter.School = school
	}
	if author := c.Query("author"); author != "" {
		filter.AuthorID = author
	}
	if status := c.Query("status"); status != "" {
		filter.Status = status
	}

	// 
	if tags := c.QueryArray("tags"); len(tags) > 0 {
		filter.Tags = tags
	}

	// 
	if difficulties := c.QueryArray("difficulty"); len(difficulties) > 0 {
		filter.Difficulty = make([]int, 0, len(difficulties))
		for _, d := range difficulties {
			if difficulty, err := strconv.Atoi(d); err == nil {
				filter.Difficulty = append(filter.Difficulty, difficulty)
			}
		}
	}

	// 
	if dateRange := c.QueryArray("dateRange"); len(dateRange) == 2 {
		if startDate, err := time.Parse("2006-01-02", dateRange[0]); err == nil {
			filter.DateFrom = &startDate
		}
		if endDate, err := time.Parse("2006-01-02", dateRange[1]); err == nil {
			filter.DateTo = &endDate
		}
	}

	// 
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && page > 0 {
		filter.Page = page
	}
	if pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10")); err == nil && pageSize > 0 {
		filter.Size = pageSize
	}

	// 
	if sortBy := c.Query("sortBy"); sortBy != "" {
		filter.SortBy = sortBy
	}
	if sortOrder := c.Query("sortOrder"); sortOrder != "" {
		filter.SortOrder = sortOrder
	}

	return filter
}

