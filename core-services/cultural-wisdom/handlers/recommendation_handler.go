package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/services"
)

// RecommendationHandler 推荐处理器
type RecommendationHandler struct {
	recommendationService *services.RecommendationService
	logger                *zap.Logger
}

// NewRecommendationHandler 创建推荐处理器实例
func NewRecommendationHandler(recommendationService *services.RecommendationService, logger *zap.Logger) *RecommendationHandler {
	return &RecommendationHandler{
		recommendationService: recommendationService,
		logger:                logger,
	}
}

// GetRecommendations 获取推荐
// @Summary 获取智慧推荐
// @Description 基于指定智慧获取推荐列表
// @Tags 推荐系统
// @Accept json
// @Produce json
// @Param wisdom_id path string true "智慧ID"
// @Param limit query int false "推荐数量限制" default(5)
// @Param algorithm query string false "推荐算法" Enums(content,collaborative,hybrid) default(hybrid)
// @Param categories query string false "分类过滤，多个用逗号分隔"
// @Param schools query string false "学派过滤，多个用逗号分隔"
// @Param authors query string false "作者过滤，多个用逗号分隔"
// @Success 200 {object} RecommendationsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/cultural-wisdom/{wisdom_id}/recommendations [get]
func (h *RecommendationHandler) GetRecommendations(c *gin.Context) {
	wisdomID := c.Param("wisdom_id")
	if wisdomID == "" {
		h.logger.Error("Missing wisdom_id parameter")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_PARAMETER",
			Message: "智慧ID不能为空",
		})
		return
	}

	// 解析查询参数
	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}
	if limit > 50 {
		limit = 50
	}

	algorithm := c.DefaultQuery("algorithm", "hybrid")
	categories := parseCommaSeparated(c.Query("categories"))
	schools := parseCommaSeparated(c.Query("schools"))
	authors := parseCommaSeparated(c.Query("authors"))

	// 构建推荐请求
	req := services.RecommendationRequest{
		WisdomID:   wisdomID,
		Limit:      limit,
		Algorithm:  algorithm,
		Categories: categories,
		Schools:    schools,
		Authors:    authors,
	}

	// 获取推荐
	recommendations, err := h.recommendationService.GetRecommendations(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get recommendations",
			zap.Error(err),
			zap.String("wisdom_id", wisdomID),
			zap.String("algorithm", algorithm))

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "RECOMMENDATION_ERROR",
			Message: "获取推荐失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RecommendationsResponse{
		Code:    200,
		Message: "获取推荐成功",
		Data:    recommendations,
		Total:   len(recommendations),
	})
}

// GetPersonalizedRecommendations 获取个性化推荐
// @Summary 获取个性化推荐
// @Description 基于用户行为获取个性化推荐
// @Tags 推荐系统
// @Accept json
// @Produce json
// @Param user_id query string true "用户ID"
// @Param limit query int false "推荐数量限制" default(10)
// @Param algorithm query string false "推荐算法" Enums(content,collaborative,hybrid) default(hybrid)
// @Success 200 {object} RecommendationsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/cultural-wisdom/recommendations/personalized [get]
func (h *RecommendationHandler) GetPersonalizedRecommendations(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_PARAMETER",
			Message: "用户ID不能为空",
		})
		return
	}

	// 解析查询参数
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	algorithm := c.DefaultQuery("algorithm", "hybrid")

	// 构建推荐请求
	req := services.RecommendationRequest{
		UserID:    userID,
		Limit:     limit,
		Algorithm: algorithm,
	}

	// 获取推荐
	recommendations, err := h.recommendationService.GetRecommendations(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get personalized recommendations",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.String("algorithm", algorithm))

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "RECOMMENDATION_ERROR",
			Message: "获取个性化推荐失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RecommendationsResponse{
		Code:    200,
		Message: "获取个性化推荐成功",
		Data:    recommendations,
		Total:   len(recommendations),
	})
}

// GetSimilarWisdoms 获取相似智慧
// @Summary 获取相似智慧
// @Description 基于内容相似度获取相似智慧
// @Tags 推荐系统
// @Accept json
// @Produce json
// @Param wisdom_id path string true "智慧ID"
// @Param limit query int false "推荐数量限制" default(5)
// @Success 200 {object} RecommendationsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/cultural-wisdom/{wisdom_id}/similar [get]
func (h *RecommendationHandler) GetSimilarWisdoms(c *gin.Context) {
	wisdomID := c.Param("wisdom_id")
	if wisdomID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_PARAMETER",
			Message: "智慧ID不能为空",
		})
		return
	}

	// 解析查询参数
	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	// 构建推荐请求（强制使用内容相似度算法）
	req := services.RecommendationRequest{
		WisdomID:  wisdomID,
		Limit:     limit,
		Algorithm: "content",
	}

	// 获取推荐
	recommendations, err := h.recommendationService.GetRecommendations(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get similar wisdoms",
			zap.Error(err),
			zap.String("wisdom_id", wisdomID))

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "RECOMMENDATION_ERROR",
			Message: "获取相似智慧失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RecommendationsResponse{
		Code:    200,
		Message: "获取相似智慧成功",
		Data:    recommendations,
		Total:   len(recommendations),
	})
}

// BatchRecommendations 批量获取推荐
// @Summary 批量获取推荐
// @Description 基于多个智慧ID批量获取推荐
// @Tags 推荐系统
// @Accept json
// @Produce json
// @Param request body BatchRecommendationRequest true "批量推荐请求"
// @Success 200 {object} BatchRecommendationsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/cultural-wisdom/recommendations/batch [post]
func (h *RecommendationHandler) BatchRecommendations(c *gin.Context) {
	var req BatchRecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效: " + err.Error(),
		})
		return
	}

	if len(req.WisdomIDs) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "EMPTY_WISDOM_IDS",
			Message: "智慧ID列表不能为空",
		})
		return
	}

	if len(req.WisdomIDs) > 10 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "TOO_MANY_WISDOM_IDS",
			Message: "一次最多只能处理10个智慧ID",
		})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 3
	}
	if req.Limit > 10 {
		req.Limit = 10
	}

	// 批量获取推荐
	results := make(map[string][]services.RecommendationItem)
	errors := make(map[string]string)

	for _, wisdomID := range req.WisdomIDs {
		recReq := services.RecommendationRequest{
			WisdomID:  wisdomID,
			Limit:     req.Limit,
			Algorithm: req.Algorithm,
		}

		recommendations, err := h.recommendationService.GetRecommendations(c.Request.Context(), recReq)
		if err != nil {
			h.logger.Warn("Failed to get recommendations for wisdom",
				zap.Error(err),
				zap.String("wisdom_id", wisdomID))
			errors[wisdomID] = err.Error()
		} else {
			results[wisdomID] = recommendations
		}
	}

	response := BatchRecommendationsResponse{
		Code:    200,
		Message: "批量推荐完成",
		Data:    results,
		Total:   len(results),
	}

	if len(errors) > 0 {
		response.Errors = errors
	}

	c.JSON(http.StatusOK, response)
}

// 响应结构体
type RecommendationsResponse struct {
	Code    int                              `json:"code"`
	Message string                           `json:"message"`
	Data    []services.RecommendationItem    `json:"data"`
	Total   int                              `json:"total"`
}

type BatchRecommendationsResponse struct {
	Code    int                                         `json:"code"`
	Message string                                      `json:"message"`
	Data    map[string][]services.RecommendationItem    `json:"data"`
	Total   int                                         `json:"total"`
	Errors  map[string]string                           `json:"errors,omitempty"`
}

// 请求结构体
type BatchRecommendationRequest struct {
	WisdomIDs []string `json:"wisdom_ids" binding:"required"`
	Limit     int      `json:"limit"`
	Algorithm string   `json:"algorithm"`
}

// parseCommaSeparated 解析逗号分隔的字符串
func parseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	
	var result []string
	for _, item := range strings.Split(s, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	return result
}