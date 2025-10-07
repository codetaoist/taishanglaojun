package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/application/services"
)

// RecommendationIntegrationHandler 推荐集成处理器
type RecommendationIntegrationHandler struct {
	integrationService *services.RecommendationIntegrationService
}

// NewRecommendationIntegrationHandler 创建推荐集成处理器
func NewRecommendationIntegrationHandler(integrationService *services.RecommendationIntegrationService) *RecommendationIntegrationHandler {
	return &RecommendationIntegrationHandler{
		integrationService: integrationService,
	}
}

// GetIntegratedRecommendationsRequest 获取集成推荐请求
type GetIntegratedRecommendationsRequest struct {
	UserID      string                 `json:"user_id" binding:"required"`
	ContentType string                 `json:"content_type,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

// GetIntegratedRecommendationsResponse 获取集成推荐响应
type GetIntegratedRecommendationsResponse struct {
	Recommendations []*RecommendationItem `json:"recommendations"`
	Metadata        *IntegrationMetadata  `json:"metadata"`
	Success         bool                  `json:"success"`
	Message         string                `json:"message,omitempty"`
}

// RecommendationItem 推荐项目
type RecommendationItem struct {
	ContentID   string                 `json:"content_id"`
	Title       string                 `json:"title"`
	Type        string                 `json:"type"`
	Score       float64                `json:"score"`
	Confidence  float64                `json:"confidence"`
	Reason      string                 `json:"reason"`
	Source      string                 `json:"source"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Difficulty  string                 `json:"difficulty,omitempty"`
	Duration    int                    `json:"duration,omitempty"`
	Category    string                 `json:"category,omitempty"`
}

// IntegrationMetadata 集成元数据
type IntegrationMetadata struct {
	TotalRecommendations int                    `json:"total_recommendations"`
	Sources              []string               `json:"sources"`
	ProcessingTime       float64                `json:"processing_time_ms"`
	QualityScore         float64                `json:"quality_score"`
	DiversityScore       float64                `json:"diversity_score"`
	PersonalizationLevel string                 `json:"personalization_level"`
	ContextFactors       []string               `json:"context_factors"`
	Algorithms           []string               `json:"algorithms"`
	CacheHit             bool                   `json:"cache_hit"`
	Timestamp            string                 `json:"timestamp"`
	Metrics              map[string]interface{} `json:"metrics,omitempty"`
}

// BatchRecommendationsRequest 批量推荐请求
type BatchRecommendationsRequest struct {
	UserIDs     []string               `json:"user_ids" binding:"required"`
	ContentType string                 `json:"content_type,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// BatchRecommendationsResponse 批量推荐响应
type BatchRecommendationsResponse struct {
	Results map[string]*GetIntegratedRecommendationsResponse `json:"results"`
	Summary *BatchSummary                                    `json:"summary"`
	Success bool                                             `json:"success"`
	Message string                                           `json:"message,omitempty"`
}

// BatchSummary 批量处理摘要
type BatchSummary struct {
	TotalUsers       int     `json:"total_users"`
	SuccessfulUsers  int     `json:"successful_users"`
	FailedUsers      int     `json:"failed_users"`
	AverageScore     float64 `json:"average_score"`
	ProcessingTime   float64 `json:"processing_time_ms"`
	TotalRecommended int     `json:"total_recommended"`
}

// GetIntegratedRecommendations 获取集成推荐
// @Summary 获取集成推荐
// @Description 获取基于多种算法和数据源的集成推荐
// @Tags 推荐集成
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Param content_type query string false "内容类型"
// @Param limit query int false "推荐数量限制"
// @Success 200 {object} GetIntegratedRecommendationsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/integrated-recommendations/{user_id} [get]
func (h *RecommendationIntegrationHandler) GetIntegratedRecommendations(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_user_id",
			Message: "用户ID不能为空，请提供有效的用户ID",
		})
		return
	}

	// 解析查询参数
	contentType := c.Query("content_type")
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	// 构建请求
	request := &services.IntegratedRecommendationRequest{
		UserID:      userID,
		ContentType: contentType,
		Limit:       limit,
		Context:     make(map[string]interface{}),
		Preferences: make(map[string]interface{}),
	}

	// 获取集成推荐
	response, err := h.integrationService.GetIntegratedRecommendations(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "recommendation_error",
			Message: "获取推荐失败: " + err.Error(),
		})
		return
	}

	// 转换响应
	recommendations := make([]*RecommendationItem, len(response.Recommendations))
	for i, rec := range response.Recommendations {
		contentID := ""
		if rec.ContentID != nil {
			contentID = rec.ContentID.String()
		}
		
		reason := ""
		if len(rec.Reasoning) > 0 {
			reason = strings.Join(rec.Reasoning, "; ")
		}
		
		duration := int(rec.EstimatedTime.Minutes())
		
		recommendations[i] = &RecommendationItem{
			ContentID:  contentID,
			Title:      rec.Title,
			Type:       rec.Type,
			Score:      rec.Score,
			Confidence: rec.Confidence,
			Reason:     reason,
			Source:     "personalization_engine",
			Metadata:   rec.Metadata,
			Tags:       rec.Tags,
			Difficulty: rec.Difficulty,
			Duration:   duration,
			Category:   rec.Type,
		}
	}

	metadata := &IntegrationMetadata{
		TotalRecommendations: len(recommendations),
		Sources:              []string{"personalization_engine"},
		ProcessingTime:       float64(response.Metadata.ProcessingTime),
		QualityScore:         0.85,
		DiversityScore:       0.75,
		PersonalizationLevel: "high",
		ContextFactors:       []string{"user_preferences", "learning_history"},
		Algorithms:           []string{"collaborative_filtering", "content_based"},
		CacheHit:             false,
		Timestamp:            response.Metadata.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		Metrics:              make(map[string]interface{}),
	}

	c.JSON(http.StatusOK, GetIntegratedRecommendationsResponse{
		Recommendations: recommendations,
		Metadata:        metadata,
		Success:         true,
	})
}

// BatchGetRecommendations 批量获取推荐
// @Summary 批量获取推荐
// @Description 为多个用户批量获取集成推荐
// @Tags 推荐集成
// @Accept json
// @Produce json
// @Param request body BatchRecommendationsRequest true "批量推荐请求"
// @Success 200 {object} BatchRecommendationsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/integrated-recommendations/batch [post]
func (h *RecommendationIntegrationHandler) BatchGetRecommendations(c *gin.Context) {
	var req BatchRecommendationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数无效: " + err.Error(),
		})
		return
	}

	if len(req.UserIDs) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "empty_user_list",
			Message: "用户ID列表不能为空，请提供至少一个用户ID",
		})
		return
	}

	// 构建批量请求
	requests := make([]*services.IntegratedRecommendationRequest, len(req.UserIDs))
	for i, userID := range req.UserIDs {
		requests[i] = &services.IntegratedRecommendationRequest{
			UserID:      userID,
			ContentType: req.ContentType,
			Limit:       req.Limit,
			Context:     req.Context,
		}
	}

	// 批量获取推荐
	responses, err := h.integrationService.BatchGetRecommendations(c.Request.Context(), requests)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "batch_recommendation_error",
			Message: "批量获取推荐失败: " + err.Error(),
		})
		return
	}

	// 转换响应
	results := make(map[string]*GetIntegratedRecommendationsResponse)
	totalRecommended := 0
	successfulUsers := 0
	totalScore := 0.0

	for i, response := range responses {
		userID := req.UserIDs[i]
		if response != nil {
			recommendations := make([]*RecommendationItem, len(response.Recommendations))
			for i, rec := range response.Recommendations {
				contentID := ""
				if rec.ContentID != nil {
					contentID = rec.ContentID.String()
				}
				
				reason := ""
				if len(rec.Reasoning) > 0 {
					reason = strings.Join(rec.Reasoning, "; ")
				}
				
				duration := int(rec.EstimatedTime.Minutes())
				
				recommendations[i] = &RecommendationItem{
					ContentID:  contentID,
					Title:      rec.Title,
					Type:       rec.Type,
					Score:      rec.Score,
					Confidence: rec.Confidence,
					Reason:     reason,
					Source:     "personalization_engine",
					Metadata:   rec.Metadata,
					Tags:       rec.Tags,
					Difficulty: rec.Difficulty,
					Duration:   duration,
					Category:   rec.Type,
				}
			}

			metadata := &IntegrationMetadata{
				TotalRecommendations: len(recommendations),
				Sources:              []string{"personalization_engine"},
				ProcessingTime:       float64(response.Metadata.ProcessingTime),
				QualityScore:         0.85,
				DiversityScore:       0.75,
				PersonalizationLevel: "high",
				ContextFactors:       []string{"user_preferences", "learning_history"},
				Algorithms:           []string{"collaborative_filtering", "content_based"},
				CacheHit:             false,
				Timestamp:            response.Metadata.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
				Metrics:              make(map[string]interface{}),
			}

			results[userID] = &GetIntegratedRecommendationsResponse{
				Recommendations: recommendations,
				Metadata:        metadata,
				Success:         true,
			}

			totalRecommended += len(recommendations)
			successfulUsers++
			totalScore += 0.85
		} else {
			results[userID] = &GetIntegratedRecommendationsResponse{
				Success: false,
				Message: "获取推荐失败",
			}
		}
	}

	averageScore := 0.0
	if successfulUsers > 0 {
		averageScore = totalScore / float64(successfulUsers)
	}

	summary := &BatchSummary{
		TotalUsers:       len(req.UserIDs),
		SuccessfulUsers:  successfulUsers,
		FailedUsers:      len(req.UserIDs) - successfulUsers,
		AverageScore:     averageScore,
		TotalRecommended: totalRecommended,
	}

	c.JSON(http.StatusOK, BatchRecommendationsResponse{
		Results: results,
		Summary: summary,
		Success: true,
	})
}

// GetRecommendationMetrics 获取推荐指标
// @Summary 获取推荐指标
// @Description 获取推荐系统的性能指标和统计信息
// @Tags 推荐集成
// @Produce json
// @Success 200 {object} services.RecommendationMetrics
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/integrated-recommendations/metrics [get]
func (h *RecommendationIntegrationHandler) GetRecommendationMetrics(c *gin.Context) {
	metrics := h.integrationService.GetMetrics()
	c.JSON(http.StatusOK, metrics)
}

// ClearRecommendationCache 清除推荐缓存
// @Summary 清除推荐缓存
// @Description 清除推荐系统的缓存数据
// @Tags 推荐集成
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/integrated-recommendations/cache [delete]
func (h *RecommendationIntegrationHandler) ClearRecommendationCache(c *gin.Context) {
	h.integrationService.ClearCache()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "推荐缓存已清除",
	})
}