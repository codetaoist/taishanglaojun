package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/adaptive"
)

// AdaptiveLearningHandler 自适应学习处理
// @Summary 自适应学习处理
// @Description 处理与自适应学习相关的HTTP请求
// @Tags adaptive-learning
type AdaptiveLearningHandler struct {
	adaptiveService *adaptive.AdaptiveLearningService
}

// NewAdaptiveLearningHandler 创建自适应学习处理
func NewAdaptiveLearningHandler(adaptiveService *adaptive.AdaptiveLearningService) *AdaptiveLearningHandler {
	return &AdaptiveLearningHandler{
		adaptiveService: adaptiveService,
	}
}

// AdaptLearningPath 适配学习路径
// @Summary 适配学习路径
// @Description 基于学习者表现和上下文动态调整学习路�?
// @Tags adaptive-learning
// @Accept json
// @Produce json
// @Param request body adaptive.PathAdaptationRequest true "路径适配请求"
// @Success 200 {object} adaptive.PathAdaptationResponse "适配成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错�?
// @Router /api/v1/adaptive/adapt-path [post]
func (h *AdaptiveLearningHandler) AdaptLearningPath(c *gin.Context) {
	var req adaptive.PathAdaptationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数格式错误: " + err.Error(),
		})
		return
	}

	// 验证请求参数
	if err := h.validateAdaptationRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_failed",
			Message: err.Error(),
		})
		return
	}

	// 执行路径适配
	response, err := h.adaptiveService.AdaptLearningPath(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "adaptation_failed",
			Message: "路径适配失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetAdaptationRecommendations 获取适配推荐
// @Summary 获取适配推荐
// @Description 基于学习者当前状态获取路径适配推荐
// @Tags adaptive-learning
// @Accept json
// @Produce json
// @Param learner_id path string true "学习者ID"
// @Param path_id query string false "当前路径ID"
// @Param analysis_depth query string false "分析深度" Enums(basic, detailed, comprehensive)
// @Success 200 {object} AdaptationRecommendationsResponse "推荐成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "学习者不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错�?
// @Router /api/v1/adaptive/recommendations/{learner_id} [get]
func (h *AdaptiveLearningHandler) GetAdaptationRecommendations(c *gin.Context) {
	learnerID := c.Param("learner_id")
	if learnerID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameter",
			Message: "学习者ID不能为空",
		})
		return
	}

	pathID := c.Query("path_id")
	analysisDepth := c.DefaultQuery("analysis_depth", "basic")

	// 构建推荐请求
	req := AdaptationRecommendationRequest{
		LearnerID:          learnerID,
		PathID:             pathID,
		AnalysisDepth:      analysisDepth,
		IncludeReasoning:   c.DefaultQuery("include_reasoning", "false") == "true",
		MaxRecommendations: h.parseIntQuery(c, "max_recommendations", 5),
	}

	// 获取适配推荐
	recommendations, err := h.getAdaptationRecommendations(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "recommendation_failed",
			Message: "获取推荐失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, AdaptationRecommendationsResponse{
		LearnerID:       learnerID,
		PathID:          pathID,
		Recommendations: recommendations,
		GeneratedAt:     getCurrentTime(),
		AnalysisDepth:   analysisDepth,
	})
}

// GetLearnerAdaptationHistory 获取学习者适配历史
// @Summary 获取学习者适配历史
// @Description 获取学习者的路径适配历史记录
// @Tags adaptive-learning
// @Accept json
// @Produce json
// @Param learner_id path string true "学习者ID"
// @Param limit query int false "返回记录数限�? default(20)
// @Param offset query int false "偏移�? default(0)
// @Success 200 {object} AdaptationHistoryResponse "获取成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "学习者不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错�?
// @Router /api/v1/adaptive/history/{learner_id} [get]
func (h *AdaptiveLearningHandler) GetLearnerAdaptationHistory(c *gin.Context) {
	learnerID := c.Param("learner_id")
	if learnerID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameter",
			Message: "学习者ID不能为空",
		})
		return
	}

	limit := h.parseIntQuery(c, "limit", 20)
	offset := h.parseIntQuery(c, "offset", 0)

	// 获取适配历史
	history, err := h.getAdaptationHistory(c.Request.Context(), learnerID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "history_retrieval_failed",
			Message: "获取适配历史失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, AdaptationHistoryResponse{
		LearnerID: learnerID,
		History:   history,
		Total:     len(history),
		Limit:     limit,
		Offset:    offset,
	})
}

// AnalyzeLearningEffectiveness 分析学习效果
// @Summary 分析学习效果
// @Description 分析适配后的学习效果和改进建�?
// @Tags adaptive-learning
// @Accept json
// @Produce json
// @Param request body EffectivenessAnalysisRequest true "效果分析请求"
// @Success 200 {object} EffectivenessAnalysisResponse "分析成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错�?
// @Router /api/v1/adaptive/analyze-effectiveness [post]
func (h *AdaptiveLearningHandler) AnalyzeLearningEffectiveness(c *gin.Context) {
	var req EffectivenessAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数格式错误: " + err.Error(),
		})
		return
	}

	// 验证请求参数
	if err := h.validateEffectivenessRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_failed",
			Message: err.Error(),
		})
		return
	}

	// 执行效果分析
	analysis, err := h.analyzeLearningEffectiveness(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "analysis_failed",
			Message: "效果分析失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// PredictLearningOutcome 预测学习结果
// @Summary 预测学习结果
// @Description 基于当前状态和路径预测学习结果
// @Tags adaptive-learning
// @Accept json
// @Produce json
// @Param request body OutcomePredictionRequest true "结果预测请求"
// @Success 200 {object} OutcomePredictionResponse "预测成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错�?
// @Router /api/v1/adaptive/predict-outcome [post]
func (h *AdaptiveLearningHandler) PredictLearningOutcome(c *gin.Context) {
	var req OutcomePredictionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数格式错误: " + err.Error(),
		})
		return
	}

	// 验证请求参数
	if err := h.validatePredictionRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_failed",
			Message: err.Error(),
		})
		return
	}

	// 执行结果预测
	prediction, err := h.predictLearningOutcome(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "prediction_failed",
			Message: "结果预测失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, prediction)
}

// 数据结构定义

// AdaptationRecommendationRequest 适配推荐请求
type AdaptationRecommendationRequest struct {
	LearnerID          string `json:"learner_id"`
	PathID             string `json:"path_id"`
	AnalysisDepth      string `json:"analysis_depth"`
	IncludeReasoning   bool   `json:"include_reasoning"`
	MaxRecommendations int    `json:"max_recommendations"`
}

// AdaptationRecommendationsResponse 适配推荐响应
type AdaptationRecommendationsResponse struct {
	LearnerID       string                     `json:"learner_id"`
	PathID          string                     `json:"path_id"`
	Recommendations []AdaptationRecommendation `json:"recommendations"`
	GeneratedAt     string                     `json:"generated_at"`
	AnalysisDepth   string                     `json:"analysis_depth"`
}

// AdaptationRecommendation 适配推荐
type AdaptationRecommendation struct {
	RecommendationID   string                 `json:"recommendation_id"`
	Type               string                 `json:"type"`
	Title              string                 `json:"title"`
	Description        string                 `json:"description"`
	Priority           string                 `json:"priority"`
	ExpectedImpact     float64                `json:"expected_impact"`
	ImplementationTime string                 `json:"implementation_time"`
	Reasoning          string                 `json:"reasoning,omitempty"`
	Parameters         map[string]interface{} `json:"parameters"`
}

// AdaptationHistoryResponse 适配历史响应
type AdaptationHistoryResponse struct {
	LearnerID string                  `json:"learner_id"`
	History   []AdaptationHistoryItem `json:"history"`
	Total     int                     `json:"total"`
	Limit     int                     `json:"limit"`
	Offset    int                     `json:"offset"`
}

// AdaptationHistoryItem 适配历史�?
type AdaptationHistoryItem struct {
	AdaptationID   string                 `json:"adaptation_id"`
	PathID         string                 `json:"path_id"`
	AdaptationType string                 `json:"adaptation_type"`
	Timestamp      string                 `json:"timestamp"`
	Reason         string                 `json:"reason"`
	Impact         float64                `json:"impact"`
	Success        bool                   `json:"success"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// EffectivenessAnalysisRequest 效果分析请求
type EffectivenessAnalysisRequest struct {
	LearnerID        string   `json:"learner_id" binding:"required"`
	PathID           string   `json:"path_id" binding:"required"`
	AdaptationID     string   `json:"adaptation_id" binding:"required"`
	AnalysisPeriod   string   `json:"analysis_period"`
	MetricsToAnalyze []string `json:"metrics_to_analyze"`
}

// EffectivenessAnalysisResponse 效果分析响应
type EffectivenessAnalysisResponse struct {
	LearnerID      string                  `json:"learner_id"`
	PathID         string                  `json:"path_id"`
	AdaptationID   string                  `json:"adaptation_id"`
	AnalysisPeriod string                  `json:"analysis_period"`
	OverallScore   float64                 `json:"overall_score"`
	MetricAnalysis []MetricAnalysisResult  `json:"metric_analysis"`
	Improvements   []ImprovementSuggestion `json:"improvements"`
	NextSteps      []string                `json:"next_steps"`
	AnalyzedAt     string                  `json:"analyzed_at"`
}

// MetricAnalysisResult 指标分析结果
type MetricAnalysisResult struct {
	MetricName     string  `json:"metric_name"`
	BeforeValue    float64 `json:"before_value"`
	AfterValue     float64 `json:"after_value"`
	ChangePercent  float64 `json:"change_percent"`
	Significance   string  `json:"significance"`
	Interpretation string  `json:"interpretation"`
}

// ImprovementSuggestion 改进建议
type ImprovementSuggestion struct {
	Area                 string  `json:"area"`
	Suggestion           string  `json:"suggestion"`
	ExpectedImpact       float64 `json:"expected_impact"`
	ImplementationEffort string  `json:"implementation_effort"`
}

// OutcomePredictionRequest 结果预测请求
type OutcomePredictionRequest struct {
	LearnerID         string                 `json:"learner_id" binding:"required"`
	PathID            string                 `json:"path_id" binding:"required"`
	CurrentState      map[string]interface{} `json:"current_state"`
	PredictionHorizon string                 `json:"prediction_horizon"`
	Scenarios         []PredictionScenario   `json:"scenarios"`
}

// PredictionScenario 预测场景
type PredictionScenario struct {
	ScenarioID  string                 `json:"scenario_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// OutcomePredictionResponse 结果预测响应
type OutcomePredictionResponse struct {
	LearnerID         string                     `json:"learner_id"`
	PathID            string                     `json:"path_id"`
	PredictionHorizon string                     `json:"prediction_horizon"`
	Predictions       []OutcomePrediction        `json:"predictions"`
	Confidence        float64                    `json:"confidence"`
	Assumptions       []string                   `json:"assumptions"`
	Recommendations   []PredictionRecommendation `json:"recommendations"`
	PredictedAt       string                     `json:"predicted_at"`
}

// OutcomePrediction 结果预测
type OutcomePrediction struct {
	ScenarioID         string   `json:"scenario_id"`
	ScenarioName       string   `json:"scenario_name"`
	SuccessProbability float64  `json:"success_probability"`
	ExpectedCompletion string   `json:"expected_completion"`
	PredictedScore     float64  `json:"predicted_score"`
	RiskFactors        []string `json:"risk_factors"`
	SuccessFactors     []string `json:"success_factors"`
}

// PredictionRecommendation 预测推荐
type PredictionRecommendation struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Impact      float64 `json:"impact"`
	Priority    string  `json:"priority"`
}

// 辅助方法

func (h *AdaptiveLearningHandler) validateAdaptationRequest(req *adaptive.PathAdaptationRequest) error {
	if req.LearnerID == "" {
		return fmt.Errorf("学习者ID不能为空")
	}
	if req.CurrentPathID == "" {
		return fmt.Errorf("当前路径ID不能为空")
	}
	return nil
}

func (h *AdaptiveLearningHandler) validateEffectivenessRequest(req *EffectivenessAnalysisRequest) error {
	if req.LearnerID == "" {
		return fmt.Errorf("学习者ID不能为空")
	}
	if req.PathID == "" {
		return fmt.Errorf("路径ID不能为空")
	}
	if req.AdaptationID == "" {
		return fmt.Errorf("适配ID不能为空")
	}
	return nil
}

func (h *AdaptiveLearningHandler) validatePredictionRequest(req *OutcomePredictionRequest) error {
	if req.LearnerID == "" {
		return fmt.Errorf("学习者ID不能为空")
	}
	if req.PathID == "" {
		return fmt.Errorf("路径ID不能为空")
	}
	return nil
}

func (h *AdaptiveLearningHandler) parseIntQuery(c *gin.Context, key string, defaultValue int) int {
	if value := c.Query(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// 占位符方法实�?
func (h *AdaptiveLearningHandler) getAdaptationRecommendations(ctx context.Context, req *AdaptationRecommendationRequest) ([]AdaptationRecommendation, error) {
	// 这里应该调用服务层方法获取推�?
	return []AdaptationRecommendation{
		{
			RecommendationID:   "rec-001",
			Type:               "difficulty_adjustment",
			Title:              "降低内容难度",
			Description:        "基于最近的表现数据，建议适当降低内容难度以提高学习效�?,
			Priority:           "high",
			ExpectedImpact:     0.25,
			ImplementationTime: "immediate",
			Reasoning:          "学习者在高难度内容上的完成率较低",
			Parameters: map[string]interface{}{
				"difficulty_reduction": 0.2,
				"gradual_adjustment":   true,
			},
		},
	}, nil
}

func (h *AdaptiveLearningHandler) getAdaptationHistory(ctx context.Context, learnerID string, limit, offset int) ([]AdaptationHistoryItem, error) {
	// 这里应该从数据库获取适配历史
	return []AdaptationHistoryItem{
		{
			AdaptationID:   "adapt-001",
			PathID:         "path-123",
			AdaptationType: "difficulty_adjustment",
			Timestamp:      getCurrentTime(),
			Reason:         "表现低于预期",
			Impact:         0.15,
			Success:        true,
			Metadata: map[string]interface{}{
				"original_difficulty": "intermediate",
				"adjusted_difficulty": "beginner",
			},
		},
	}, nil
}

func (h *AdaptiveLearningHandler) analyzeLearningEffectiveness(ctx context.Context, req *EffectivenessAnalysisRequest) (*EffectivenessAnalysisResponse, error) {
	// 这里应该调用服务层方法进行效果分�?
	return &EffectivenessAnalysisResponse{
		LearnerID:      req.LearnerID,
		PathID:         req.PathID,
		AdaptationID:   req.AdaptationID,
		AnalysisPeriod: req.AnalysisPeriod,
		OverallScore:   0.78,
		MetricAnalysis: []MetricAnalysisResult{
			{
				MetricName:     "completion_rate",
				BeforeValue:    0.65,
				AfterValue:     0.82,
				ChangePercent:  26.15,
				Significance:   "high",
				Interpretation: "适配后完成率显著提升",
			},
		},
		Improvements: []ImprovementSuggestion{
			{
				Area:                 "engagement",
				Suggestion:           "增加更多互动元素",
				ExpectedImpact:       0.15,
				ImplementationEffort: "medium",
			},
		},
		NextSteps: []string{
			"继续监控学习进度",
			"考虑进一步个性化调整",
		},
		AnalyzedAt: getCurrentTime(),
	}, nil
}

func (h *AdaptiveLearningHandler) predictLearningOutcome(ctx context.Context, req *OutcomePredictionRequest) (*OutcomePredictionResponse, error) {
	// 这里应该调用服务层方法进行结果预�?
	return &OutcomePredictionResponse{
		LearnerID:         req.LearnerID,
		PathID:            req.PathID,
		PredictionHorizon: req.PredictionHorizon,
		Predictions: []OutcomePrediction{
			{
				ScenarioID:         "scenario-1",
				ScenarioName:       "当前路径继续",
				SuccessProbability: 0.75,
				ExpectedCompletion: "2024-03-15",
				PredictedScore:     0.82,
				RiskFactors:        []string{"时间压力", "难度跳跃"},
				SuccessFactors:     []string{"良好基础", "高动力学�?},
			},
		},
		Confidence: 0.85,
		Assumptions: []string{
			"学习者保持当前学习频�?,
			"无重大外部干�?,
		},
		Recommendations: []PredictionRecommendation{
			{
				Type:        "pacing_adjustment",
				Description: "建议适当放慢学习节奏",
				Impact:      0.1,
				Priority:    "medium",
			},
		},
		PredictedAt: getCurrentTime(),
	}, nil
}

func getCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}
