package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/services"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
)

// SearchHandler 搜索API处理器
type SearchHandler struct {
	searchService *services.SearchService
}

// NewSearchHandler 创建搜索处理器实例
func NewSearchHandler(searchService *services.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

// FullTextSearch 全文搜索
// @Summary 全文搜索
// @Description 在文化智慧内容中进行全文搜索
// @Tags 搜索
// @Produce json
// @Param q query string true "搜索关键词"
// @Param category query string false "分类ID"
// @Param school query string false "学派" Enums(儒家,道家,佛家)
// @Param tags query string false "标签，多个用逗号分隔"
// @Param difficulty query string false "难度等级，多个用逗号分隔"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(20)
// @Success 200 {object} SearchResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/search [get]
func (h *SearchHandler) FullTextSearch(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_QUERY",
			Message: "搜索关键词不能为空",
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

	// 解析页码
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	// 解析每页数量
	if sizeStr := c.Query("size"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 && size <= 100 {
			req.Size = size
		}
	}

	// 解析标签
	if tagsStr := c.Query("tags"); tagsStr != "" {
		// 这里可以实现标签解析逻辑
		// req.Tags = strings.Split(tagsStr, ",")
	}

	result, err := h.searchService.FullTextSearch(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "SEARCH_ERROR",
			Message: "搜索失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// SemanticSearch 语义搜索
// @Summary 语义搜索
// @Description 基于语义理解的智能搜索
// @Tags 搜索
// @Accept json
// @Produce json
// @Param request body SemanticSearchRequest true "语义搜索请求"
// @Success 200 {object} SearchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/search/semantic [post]
func (h *SearchHandler) SemanticSearch(c *gin.Context) {
	// 检查用户认证
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "用户未认证",
		})
		return
	}

	var req SemanticSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 20
	}
	if req.Threshold == 0 {
		req.Threshold = 0.7 // 默认相似度阈值
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
			Message: "语义搜索失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetSearchSuggestions 获取搜索建议
// @Summary 获取搜索建议
// @Description 根据输入获取搜索建议
// @Tags 搜索
// @Produce json
// @Param q query string true "搜索关键词前缀"
// @Param limit query int false "建议数量限制" default(10)
// @Success 200 {object} SuggestionsResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/search/suggestions [get]
func (h *SearchHandler) GetSearchSuggestions(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_QUERY",
			Message: "搜索关键词不能为空",
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
			Message: "获取搜索建议失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuggestionsResponse{
		Suggestions: suggestions,
		Count:       len(suggestions),
	})
}

// GetPopularSearches 获取热门搜索
// @Summary 获取热门搜索
// @Description 获取热门搜索关键词
// @Tags 搜索
// @Produce json
// @Param limit query int false "数量限制" default(20)
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
			Message: "获取热门搜索失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, PopularSearchesResponse{
		Searches: convertPopularSearches(searches),
		Count:    len(searches),
	})
}

// 请求和响应结构体
type SemanticSearchRequest struct {
	Query      string   `json:"query" binding:"required"`
	CategoryID string   `json:"category_id"`
	School     string   `json:"school"`
	Tags       []string `json:"tags"`
	Threshold  float32  `json:"threshold"` // 相似度阈值 0-1
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

// convertPopularSearches 转换services.PopularSearch到handlers.PopularSearch
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


// GetCategories 获取分类列表
// @Summary 获取分类列表
// @Description 获取所有可用的智慧内容分类
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
			Message: "获取分类列表失败: " + err.Error(),
		})
		return
	}

	response := CategoriesResponse{
		Code:    200,
		Message: "获取分类列表成功",
		Data:    categories,
	}

	c.JSON(http.StatusOK, response)
}

// SearchByCategory 按分类搜索
// @Summary 按分类搜索智慧内容
// @Description 根据分类ID搜索智慧内容
// @Tags search
// @Accept json
// @Produce json
// @Param category path string true "分类名称"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} SearchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/search/category/{category} [get]
func (h *SearchHandler) SearchByCategory(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_CATEGORY",
			Message: "分类参数不能为空",
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
			Message: "按分类搜索失败: " + err.Error(),
		})
		return
	}

	response := SearchResponse{
		Code:    200,
		Message: "搜索成功",
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
	TimeTaken int64       `json:"time_taken,omitempty"` // 毫秒
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse 成功响应结构体
type SuccessResponse struct {
	Message string `json:"message"`
}


// AdvancedSearch 高级搜索
func (h *SearchHandler) AdvancedSearch(c *gin.Context) {
	var req services.AdvancedSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数: " + err.Error()})
		return
	}

	// 设置默认值
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": result,
		"message": "搜索成功",
	})
}

// GetSearchFilters 获取搜索过滤器选项
func (h *SearchHandler) GetSearchFilters(c *gin.Context) {
	filters, err := h.searchService.GetSearchFilters(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取过滤器失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": filters,
		"message": "获取过滤器成功",
	})
}

// SearchWithFacets 带分面的搜索
func (h *SearchHandler) SearchWithFacets(c *gin.Context) {
	var req services.SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数: " + err.Error()})
		return
	}

	// 设置默认值
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": result,
		"message": "搜索成功",
	})
}

