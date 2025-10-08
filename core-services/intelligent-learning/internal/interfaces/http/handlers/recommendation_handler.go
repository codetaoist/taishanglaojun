package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	domainServices "github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
)

// RecommendationHandler 推荐系统处理器
type RecommendationHandler struct {
	personalizationEngine *domainServices.PersonalizationEngine
	userBehaviorTracker   *domainServices.UserBehaviorTracker
	preferenceAnalyzer    *domainServices.PreferenceAnalyzer
	contextAnalyzer       *domainServices.ContextAnalyzer
}

// NewRecommendationHandler 创建推荐系统处理器
func NewRecommendationHandler(
	personalizationEngine *domainServices.PersonalizationEngine,
	userBehaviorTracker *domainServices.UserBehaviorTracker,
	preferenceAnalyzer *domainServices.PreferenceAnalyzer,
	contextAnalyzer *domainServices.ContextAnalyzer,
) *RecommendationHandler {
	return &RecommendationHandler{
		personalizationEngine: personalizationEngine,
		userBehaviorTracker:   userBehaviorTracker,
		preferenceAnalyzer:    preferenceAnalyzer,
		contextAnalyzer:       contextAnalyzer,
	}
}

// PersonalizedRecommendationRequest 个性化推荐请求
type PersonalizedRecommendationRequest struct {
	UserID      string                 `json:"user_id" binding:"required"`
	ContentType string                 `json:"content_type,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

// RecommendationResponse 推荐响应
type RecommendationResponse struct {
	Recommendations []PersonalizedRecommendation `json:"recommendations"`
	Metadata        RecommendationMetadata       `json:"metadata"`
}

// PersonalizedRecommendation 个性化推荐项
type PersonalizedRecommendation struct {
	ContentID    string                 `json:"content_id"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	ContentType  string                 `json:"content_type"`
	Score        float64                `json:"score"`
	Confidence   float64                `json:"confidence"`
	Reasoning    string                 `json:"reasoning"`
	Tags         []string               `json:"tags"`
	Metadata     map[string]interface{} `json:"metadata"`
	EstimatedTime int                   `json:"estimated_time"`
	Difficulty   string                 `json:"difficulty"`
}

// RecommendationMetadata 推荐元数据
type RecommendationMetadata struct {
	Strategy      string                 `json:"strategy"`
	Timestamp     time.Time              `json:"timestamp"`
	UserProfile   map[string]interface{} `json:"user_profile"`
	Context       map[string]interface{} `json:"context"`
	TotalCount    int                    `json:"total_count"`
	ProcessingTime int64                 `json:"processing_time_ms"`
}

// GetPersonalizedRecommendations 获取个性化推荐
// @Summary 获取个性化推荐
// @Description 基于用户行为、偏好和上下文生成个性化推荐
// @Tags recommendations
// @Accept json
// @Produce json
// @Param request body PersonalizedRecommendationRequest true "推荐请求"
// @Success 200 {object} RecommendationResponse "推荐成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/recommendations/personalized [post]
func (h *RecommendationHandler) GetPersonalizedRecommendations(c *gin.Context) {
	var req PersonalizedRecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数格式错误: " + err.Error(),
		})
		return
	}

	// 设置默认值
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 50 {
		req.Limit = 50
	}

	startTime := time.Now()

	// 构建个性化请求
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_user_id",
			Message: "无效的用户ID格式",
		})
		return
	}

	personalizationReq := &domainServices.PersonalizationRequest{
		LearnerID:           userID,
		RecommendationType:  req.ContentType,
		MaxRecommendations:  req.Limit,
		IncludeExplanations: true,
		Filters:             req.Preferences,
		PersonalizationLevel: "advanced",
	}

	// 生成个性化推荐
	response, err := h.personalizationEngine.GeneratePersonalizedRecommendations(c.Request.Context(), personalizationReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "recommendation_failed",
			Message: "生成推荐失败: " + err.Error(),
		})
		return
	}

	// 转换响应格式
	recommendations := make([]PersonalizedRecommendation, len(response.Recommendations))
	for i, rec := range response.Recommendations {
		contentID := ""
		if rec.ContentID != nil {
			contentID = rec.ContentID.String()
		}
		
		reasoning := ""
		if len(rec.Reasoning) > 0 {
			reasoning = strings.Join(rec.Reasoning, "; ")
		}
		
		estimatedTime := int(rec.EstimatedTime.Minutes())
		
		recommendations[i] = PersonalizedRecommendation{
			ContentID:     contentID,
			Title:         rec.Title,
			Description:   rec.Description,
			ContentType:   rec.Type,
			Score:         rec.Score,
			Confidence:    rec.Confidence,
			Reasoning:     reasoning,
			Tags:          rec.Tags,
			Metadata:      rec.Metadata,
			EstimatedTime: estimatedTime,
			Difficulty:    rec.Difficulty,
		}
	}

	processingTime := time.Since(startTime).Milliseconds()

	result := RecommendationResponse{
		Recommendations: recommendations,
		Metadata: RecommendationMetadata{
			Strategy:       "personalized",
			Timestamp:      response.GeneratedAt,
			UserProfile:    make(map[string]interface{}),
			Context:        make(map[string]interface{}),
			TotalCount:     len(recommendations),
			ProcessingTime: processingTime,
		},
	}

	c.JSON(http.StatusOK, result)
}

// GetRecommendationsByStrategy 根据策略获取推荐
// @Summary 根据策略获取推荐
// @Description 使用指定策略生成推荐
// @Tags recommendations
// @Accept json
// @Produce json
// @Param strategy path string true "推荐策略" Enums(collaborative,content_based,hybrid,popular,trending)
// @Param user_id query string true "用户ID"
// @Param limit query int false "推荐数量限制" default(10)
// @Success 200 {object} RecommendationResponse "推荐成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/recommendations/strategy/{strategy} [get]
func (h *RecommendationHandler) GetRecommendationsByStrategy(c *gin.Context) {
	strategy := c.Param("strategy")
	userIDStr := c.Query("user_id")
	limitStr := c.DefaultQuery("limit", "10")

	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameter",
			Message: "用户ID不能为空",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_parameter",
			Message: "用户ID格式无效",
		})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	startTime := time.Now()

	// 构建请求
	req := &domainServices.PersonalizationRequest{
		LearnerID:           userID,
		MaxRecommendations:  limit,
		PersonalizationLevel: strategy,
		IncludeExplanations: true,
	}

	// 生成推荐
	response, err := h.personalizationEngine.GeneratePersonalizedRecommendations(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "recommendation_failed",
			Message: "生成推荐失败: " + err.Error(),
		})
		return
	}

	// 转换响应格式
	recommendations := make([]PersonalizedRecommendation, len(response.Recommendations))
	for i, rec := range response.Recommendations {
		contentID := ""
		if rec.ContentID != nil {
			contentID = rec.ContentID.String()
		}
		
		reasoning := ""
		if len(rec.Reasoning) > 0 {
			reasoning = strings.Join(rec.Reasoning, "; ")
		}
		
		estimatedTime := int(rec.EstimatedTime.Minutes())
		
		recommendations[i] = PersonalizedRecommendation{
			ContentID:     contentID,
			Title:         rec.Title,
			Description:   rec.Description,
			ContentType:   rec.Type,
			Score:         rec.Score,
			Confidence:    rec.Confidence,
			Reasoning:     reasoning,
			Tags:          rec.Tags,
			Metadata:      rec.Metadata,
			EstimatedTime: estimatedTime,
			Difficulty:    rec.Difficulty,
		}
	}

	processingTime := time.Since(startTime).Milliseconds()

	result := RecommendationResponse{
		Recommendations: recommendations,
		Metadata: RecommendationMetadata{
			Strategy:       "personalized",
			Timestamp:      response.GeneratedAt,
			UserProfile:    make(map[string]interface{}),
			Context:        make(map[string]interface{}),
			TotalCount:     len(recommendations),
			ProcessingTime: processingTime,
		},
	}

	c.JSON(http.StatusOK, result)
}

// RecordUserBehavior 记录用户行为
// @Summary 记录用户行为
// @Description 记录用户的学习行为用于推荐优化
// @Tags recommendations
// @Accept json
// @Produce json
// @Param request body domainServices.BehaviorEvent true "行为事件"
// @Success 200 {object} map[string]interface{} "记录成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/recommendations/behavior [post]
func (h *RecommendationHandler) RecordUserBehavior(c *gin.Context) {
	var event domainServices.BehaviorEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数格式错误: " + err.Error(),
		})
		return
	}

	// 设置时间戳
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 记录行为
	err := h.userBehaviorTracker.TrackBehaviorEvent(c.Request.Context(), &event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "tracking_failed",
			Message: "行为记录失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "行为记录成功",
		"timestamp": event.Timestamp,
	})
}

// GetUserPreferences 获取用户偏好分析
// @Summary 获取用户偏好分析
// @Description 分析用户的学习偏好和兴趣
// @Tags recommendations
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} services.UserPreferences "偏好分析结果"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/recommendations/preferences/{user_id} [get]
func (h *RecommendationHandler) GetUserPreferences(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameter",
			Message: "用户ID不能为空",
		})
		return
	}

	preferences, err := h.preferenceAnalyzer.AnalyzeUserPreferences(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "analysis_failed",
			Message: "偏好分析失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, preferences)
}

// GetLearningContext 获取学习上下文分析
// @Summary 获取学习上下文分析
// @Description 分析用户的学习环境和情境因素
// @Tags recommendations
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} services.LearningContext "上下文分析结果"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/recommendations/context/{user_id} [get]
func (h *RecommendationHandler) GetLearningContext(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameter",
			Message: "用户ID不能为空",
		})
		return
	}

	context, err := h.contextAnalyzer.AnalyzeLearningContext(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "analysis_failed",
			Message: "上下文分析失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, context)
}

// GetBehaviorInsights 获取行为洞察
// @Summary 获取行为洞察
// @Description 获取用户学习行为的深度分析和洞察
// @Tags recommendations
// @Produce json
// @Param user_id path string true "用户ID"
// @Param days query int false "分析天数" default(30)
// @Success 200 {object} services.LearningInsights "行为洞察结果"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/recommendations/insights/{user_id} [get]
func (h *RecommendationHandler) GetBehaviorInsights(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameter",
			Message: "用户ID不能为空",
		})
		return
	}

	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 30
	}

	// 解析用户ID
	learnerID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_user_id",
			Message: "无效的用户ID格式",
		})
		return
	}

	// 创建时间范围
	timeRange := domainServices.BehaviorTimeRange{
		Start: time.Now().AddDate(0, 0, -days),
		End:   time.Now(),
	}

	insights, err := h.userBehaviorTracker.GetLearningInsights(c.Request.Context(), learnerID, timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "analysis_failed",
			Message: "行为洞察分析失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, insights)
}

// BatchRecommendations 批量推荐
// @Summary 批量推荐
// @Description 为多个用户批量生成推荐
// @Tags recommendations
// @Accept json
// @Produce json
// @Param request body BatchRecommendationRequest true "批量推荐请求"
// @Success 200 {object} BatchRecommendationResponse "批量推荐成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/recommendations/batch [post]
func (h *RecommendationHandler) BatchRecommendations(c *gin.Context) {
	var req BatchRecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数格式错误: " + err.Error(),
		})
		return
	}

	if len(req.UserIDs) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameter",
			Message: "用户ID列表不能为空",
		})
		return
	}

	if len(req.UserIDs) > 100 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "too_many_users",
			Message: "批量推荐用户数量不能超过100个",
		})
		return
	}

	startTime := time.Now()
	results := make(map[string]RecommendationResponse)
	errors := make(map[string]string)

	// 为每个用户生成推荐
	for _, userID := range req.UserIDs {
		// 解析用户ID
		learnerID, err := uuid.Parse(userID)
		if err != nil {
			errors[userID] = "无效的用户ID格式"
			continue
		}

		personalizationReq := &domainServices.PersonalizationRequest{
			LearnerID:            learnerID,
			RecommendationType:   req.ContentType,
			MaxRecommendations:   req.Limit,
			IncludeExplanations:  true,
			Filters:              make(map[string]interface{}),
			PersonalizationLevel: "advanced",
		}

		response, err := h.personalizationEngine.GeneratePersonalizedRecommendations(c.Request.Context(), personalizationReq)
		if err != nil {
			errors[userID] = err.Error()
			continue
		}

		// 转换响应格式
		recommendations := make([]PersonalizedRecommendation, len(response.Recommendations))
		for i, rec := range response.Recommendations {
			contentID := ""
			if rec.ContentID != nil {
				contentID = rec.ContentID.String()
			}
			
			reasoning := ""
			if len(rec.Reasoning) > 0 {
				reasoning = strings.Join(rec.Reasoning, "; ")
			}
			
			estimatedTime := int(rec.EstimatedTime.Minutes())
			
			recommendations[i] = PersonalizedRecommendation{
				ContentID:     contentID,
				Title:         rec.Title,
				Description:   rec.Description,
				ContentType:   rec.Type,
				Score:         rec.Score,
				Confidence:    rec.Confidence,
				Reasoning:     reasoning,
				Tags:          rec.Tags,
				Metadata:      rec.Metadata,
				EstimatedTime: estimatedTime,
				Difficulty:    rec.Difficulty,
			}
		}

		results[userID] = RecommendationResponse{
			Recommendations: recommendations,
			Metadata: RecommendationMetadata{
				Strategy:    "personalized",
				Timestamp:   response.GeneratedAt,
				UserProfile: make(map[string]interface{}),
				Context:     make(map[string]interface{}),
				TotalCount:  len(recommendations),
			},
		}
	}

	processingTime := time.Since(startTime).Milliseconds()

	batchResponse := BatchRecommendationResponse{
		Results:        results,
		Errors:         errors,
		TotalUsers:     len(req.UserIDs),
		SuccessCount:   len(results),
		ErrorCount:     len(errors),
		ProcessingTime: processingTime,
	}

	c.JSON(http.StatusOK, batchResponse)
}

// BatchRecommendationRequest 批量推荐请求
type BatchRecommendationRequest struct {
	UserIDs     []string               `json:"user_ids" binding:"required"`
	ContentType string                 `json:"content_type,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

// BatchRecommendationResponse 批量推荐响应
type BatchRecommendationResponse struct {
	Results        map[string]RecommendationResponse `json:"results"`
	Errors         map[string]string                 `json:"errors"`
	TotalUsers     int                               `json:"total_users"`
	SuccessCount   int                               `json:"success_count"`
	ErrorCount     int                               `json:"error_count"`
	ProcessingTime int64                             `json:"processing_time_ms"`
}

// ErrorResponse 错误响应