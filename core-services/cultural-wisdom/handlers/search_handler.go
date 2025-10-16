package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/services"
	"github.com/gin-gonic/gin"
)

// SearchHandler API
type SearchHandler struct {
	searchService *services.SearchService
}

// NewSearchHandler 
func NewSearchHandler(searchService *services.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

// FullTextSearch 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param q query string true ""
// @Param category query string false "ID"
// @Param school query string false "" Enums(,,)
// @Param tags query string false ""
// @Param difficulty query string false ""
// @Param page query int false "" default(1)
// @Param size query int false "" default(20)
// @Success 200 {object} SearchResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/search [get]
func (h *SearchHandler) FullTextSearch(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_QUERY",
			Message: "",
		})
		return
	}

	req := &services.SearchRequest{
		Query:      query,
		CategoryID: c.Query("category"),
		School:     c.Query("school"),
		Page:       1,
		Size:       20,
	}

	// 
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	// 
	if sizeStr := c.Query("size"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 && size <= 100 {
			req.Size = size
		}
	}

	// 
	if tagsStr := c.Query("tags"); tagsStr != "" {
		// 
		// req.Tags = strings.Split(tagsStr, ",")
	}

	result, err := h.searchService.FullTextSearch(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "SEARCH_ERROR",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// SemanticSearch 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param request body SemanticSearchRequest true ""
// @Success 200 {object} SearchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/search/semantic [post]
func (h *SearchHandler) SemanticSearch(c *gin.Context) {
	// 
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "",
		})
		return
	}

	var req SemanticSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	// 
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 20
	}
	if req.Threshold == 0 {
		req.Threshold = 0.7 // 
	}

	searchReq := &services.SemanticSearchRequest{
		Query:      req.Query,
		UserID:     userID.(string),
		CategoryID: req.CategoryID,
		School:     req.School,
		Tags:       req.Tags,
		Threshold:  req.Threshold,
		Page:       req.Page,
		Size:       req.Size,
	}

	results, err := h.searchService.SemanticSearch(c.Request.Context(), searchReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "SEMANTIC_SEARCH_ERROR",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetSearchSuggestions 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param q query string true ""
// @Param limit query int false "" default(10)
// @Success 200 {object} SuggestionsResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/search/suggestions [get]
func (h *SearchHandler) GetSearchSuggestions(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_QUERY",
			Message: "",
		})
		return
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	suggestions, err := h.searchService.GetSearchSuggestions(c.Request.Context(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "SUGGESTIONS_ERROR",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuggestionsResponse{
		Suggestions: suggestions,
		Count:       len(suggestions),
	})
}

// GetPopularSearches 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param limit query int false "" default(20)
// @Success 200 {object} PopularSearchesResponse
// @Router /api/v1/search/popular [get]
func (h *SearchHandler) GetPopularSearches(c *gin.Context) {
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	searches, err := h.searchService.GetPopularSearches(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "POPULAR_SEARCHES_ERROR",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, PopularSearchesResponse{
		Searches: convertPopularSearches(searches),
		Count:    len(searches),
	})
}

// 
type SemanticSearchRequest struct {
	Query      string   `json:"query" binding:"required"`
	CategoryID string   `json:"category_id"`
	School     string   `json:"school"`
	Tags       []string `json:"tags"`
	Threshold  float32  `json:"threshold"` //  0-1
	Page       int      `json:"page"`
	Size       int      `json:"size"`
}

type SuggestionsResponse struct {
	Suggestions []string `json:"suggestions"`
	Count       int      `json:"count"`
}

type PopularSearchesResponse struct {
	Searches []PopularSearch `json:"searches"`
	Count    int             `json:"count"`
}

type PopularSearch struct {
	Query string `json:"query"`
	Count int64  `json:"count"`
	Rank  int    `json:"rank"`
}

// convertPopularSearches services.PopularSearchhandlers.PopularSearch
func convertPopularSearches(searches []services.PopularSearch) []PopularSearch {
	result := make([]PopularSearch, len(searches))
	for i, search := range searches {
		result[i] = PopularSearch{
			Query: search.Query,
			Count: search.Count,
			Rank:  search.Rank,
		}
	}
	return result
}

// GetCategories 
// @Summary 
// @Description 
// @Tags search
// @Accept json
// @Produce json
// @Success 200 {object} CategoriesResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/search/categories [get]
func (h *SearchHandler) GetCategories(c *gin.Context) {
	categories, err := h.searchService.GetCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CATEGORIES_ERROR",
			Message: ": " + err.Error(),
		})
		return
	}

	response := CategoriesResponse{
		Code:    200,
		Message: "",
		Data:    categories,
	}

	c.JSON(http.StatusOK, response)
}

// SearchByCategory 
// @Summary 
// @Description ID
// @Tags search
// @Accept json
// @Produce json
// @Param category path string true ""
// @Param page query int false "" default(1)
// @Param size query int false "" default(10)
// @Success 200 {object} SearchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/search/category/{category} [get]
func (h *SearchHandler) SearchByCategory(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_CATEGORY",
			Message: "",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	offset := (page - 1) * size

	wisdoms, err := h.searchService.SearchByCategory(c.Request.Context(), category, size, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CATEGORY_SEARCH_ERROR",
			Message: ": " + err.Error(),
		})
		return
	}

	response := SearchResponse{
		Code:    200,
		Message: "",
		Data: SearchData{
			Wisdoms: wisdoms,
			Total:   len(wisdoms),
			Page:    page,
			Size:    size,
		},
	}

	c.JSON(http.StatusOK, response)
}

type SearchData struct {
	Wisdoms []*models.Wisdom `json:"wisdoms"`
	Total   int              `json:"total"`
	Page    int              `json:"page"`
	Size    int              `json:"size"`
}

type SearchResponse struct {
	Results   interface{} `json:"results,omitempty"`
	Query     string      `json:"query,omitempty"`
	Code      int         `json:"code,omitempty"`
	Message   string      `json:"message,omitempty"`
	Data      SearchData  `json:"data,omitempty"`
	Page      int         `json:"page,omitempty"`
	Size      int         `json:"size,omitempty"`
	Total     int64       `json:"total,omitempty"`
	TimeTaken int64       `json:"time_taken,omitempty"` // 
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse 
type SuccessResponse struct {
	Message string `json:"message"`
}

// EnhancedSemanticSearchRequest 
type EnhancedSemanticSearchRequest struct {
	Query      string   `json:"query" binding:"required"`
	UserID     string   `json:"user_id"`
	CategoryID string   `json:"category_id"`
	School     string   `json:"school"`
	Tags       []string `json:"tags"`
	Threshold  float32  `json:"threshold"` //  0-1
	Page       int      `json:"page"`
	Size       int      `json:"size"`
}

// VectorSearchRequest 
type VectorSearchRequest struct {
	Query     string  `json:"query" binding:"required"`
	Threshold float32 `json:"threshold"` //  0-1
	Page      int     `json:"page"`
	Size      int     `json:"size"`
}

// EnhancedSearchResponse 
type EnhancedSearchResponse struct {
	Code        int                    `json:"code"`
	Message     string                 `json:"message"`
	Results     []*models.Wisdom       `json:"results"`
	Total       int                    `json:"total"`
	SearchType  string                 `json:"search_type"` // semantic, keyword, hybrid
	QueryTime   int64                  `json:"query_time"`  // 
	Suggestions []string               `json:"suggestions"`
	Facets      map[string]interface{} `json:"facets"`
	Page        int                    `json:"page"`
	Size        int                    `json:"size"`
}

// SearchAnalyticsResponse 
type SearchAnalyticsResponse struct {
	Code            int             `json:"code"`
	Message         string          `json:"message"`
	Period          string          `json:"period"`
	PopularSearches []PopularSearch `json:"popular_searches"`
	TotalSearches   int64           `json:"total_searches"`
}

// AdvancedSearch 
func (h *SearchHandler) AdvancedSearch(c *gin.Context) {
	var req services.AdvancedSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ": " + err.Error()})
		return
	}

	// 
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	if req.Size > 100 {
		req.Size = 100
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	result, err := h.searchService.AdvancedSearch(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ": " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"data":    result,
		"message": "",
	})
}

// GetSearchFilters 
func (h *SearchHandler) GetSearchFilters(c *gin.Context) {
	filters, err := h.searchService.GetSearchFilters(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ": " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"data":    filters,
		"message": "",
	})
}

// SearchWithFacets 
func (h *SearchHandler) SearchWithFacets(c *gin.Context) {
	var req services.SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ": " + err.Error()})
		return
	}

	// 
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	if req.Size > 100 {
		req.Size = 100
	}

	result, err := h.searchService.SearchWithFacets(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ": " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"data":    result,
		"message": "",
	})
}

// EnhancedSemanticSearch 
// @Summary 
// @Description AI
// @Tags 
// @Accept json
// @Produce json
// @Param request body EnhancedSemanticSearchRequest true ""
// @Success 200 {object} EnhancedSearchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/search/enhanced-semantic [post]
func (h *SearchHandler) EnhancedSemanticSearch(c *gin.Context) {
	var req EnhancedSemanticSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	// 
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	if req.Size > 100 {
		req.Size = 100
	}

	// 
	serviceReq := &services.SemanticSearchRequest{
		Query:      req.Query,
		UserID:     req.UserID,
		CategoryID: req.CategoryID,
		School:     req.School,
		Tags:       req.Tags,
		Threshold:  req.Threshold,
		Page:       req.Page,
		Size:       req.Size,
	}

	result, err := h.searchService.EnhancedSemanticSearch(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "SEARCH_ERROR",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	response := EnhancedSearchResponse{
		Code:        200,
		Message:     "",
		Results:     result.Results,
		Total:       result.Total,
		SearchType:  result.SearchType,
		QueryTime:   result.QueryTime.Milliseconds(),
		Suggestions: result.Suggestions,
		Facets:      result.Facets,
		Page:        req.Page,
		Size:        req.Size,
	}

	c.JSON(http.StatusOK, response)
}

// VectorSimilaritySearch 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param request body VectorSearchRequest true ""
// @Success 200 {object} SearchResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/search/vector [post]
func (h *SearchHandler) VectorSimilaritySearch(c *gin.Context) {
	var req VectorSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	// 
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}

	// 
	serviceReq := &services.SemanticSearchRequest{
		Query:     req.Query,
		Threshold: req.Threshold,
		Page:      req.Page,
		Size:      req.Size,
	}

	results, err := h.searchService.SemanticSearch(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "SEARCH_ERROR",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	response := SearchResponse{
		Code:    200,
		Message: "",
		Data: SearchData{
			Wisdoms: results,
			Total:   len(results),
			Page:    req.Page,
			Size:    req.Size,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetSearchAnalytics 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param period query string false "" Enums(day,week,month) default(week)
// @Param limit query int false "" default(10)
// @Success 200 {object} SearchAnalyticsResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/search/analytics [get]
func (h *SearchHandler) GetSearchAnalytics(c *gin.Context) {
	period := c.DefaultQuery("period", "week")
	limitStr := c.DefaultQuery("limit", "10")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// 
	popularSearches, err := h.searchService.GetPopularSearches(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "ANALYTICS_ERROR",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	response := SearchAnalyticsResponse{
		Code:            200,
		Message:         "",
		Period:          period,
		PopularSearches: convertPopularSearches(popularSearches),
		TotalSearches:   calculateTotalSearches(popularSearches),
	}

	c.JSON(http.StatusOK, response)
}

// calculateTotalSearches 
func calculateTotalSearches(searches []services.PopularSearch) int64 {
	var total int64
	for _, search := range searches {
		total += search.Count
	}
	return total
}

