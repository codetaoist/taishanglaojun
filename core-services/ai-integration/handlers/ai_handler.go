package handlers

import (
	"net/http"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AIHandler AI功能处理器
type AIHandler struct {
	aiService *services.AIService
	logger    *zap.Logger
}

// NewAIHandler 创建AI处理器
func NewAIHandler(aiService *services.AIService, logger *zap.Logger) *AIHandler {
	return &AIHandler{
		aiService: aiService,
		logger:    logger,
	}
}

// IntentRecognition 意图识别
// @Summary 意图识别
// @Description 识别用户输入文本的意图
// @Tags AI功能
// @Accept json
// @Produce json
// @Param request body providers.IntentRequest true "意图识别请求"
// @Success 200 {object} providers.IntentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ai/intent [post]
func (h *AIHandler) IntentRecognition(c *gin.Context) {
	var req providers.IntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效: " + err.Error(),
		})
		return
	}

	// 从JWT中获取用户ID (如果需要)
	if userID := c.GetString("user_id"); userID != "" {
		req.UserID = userID
	}

	response, err := h.aiService.IntentRecognition(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Intent recognition failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTENT_RECOGNITION_FAILED",
			Message: "意图识别失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SentimentAnalysis 情感分析
// @Summary 情感分析
// @Description 分析用户输入文本的情感倾向
// @Tags AI功能
// @Accept json
// @Produce json
// @Param request body providers.SentimentRequest true "情感分析请求"
// @Success 200 {object} providers.SentimentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ai/sentiment [post]
func (h *AIHandler) SentimentAnalysis(c *gin.Context) {
	var req providers.SentimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效: " + err.Error(),
		})
		return
	}

	// 从JWT中获取用户ID (如果需要)
	if userID := c.GetString("user_id"); userID != "" {
		req.UserID = userID
	}

	response, err := h.aiService.SentimentAnalysis(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Sentiment analysis failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "SENTIMENT_ANALYSIS_FAILED",
			Message: "情感分析失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GenerateSummary 生成摘要
func (h *AIHandler) GenerateSummary(c *gin.Context) {
	var req providers.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效: " + err.Error(),
		})
		return
	}

	response, err := h.aiService.GenerateSummary(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Generate summary failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GENERATE_SUMMARY_FAILED",
			Message: "生成摘要失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GenerateExplanation 生成解释
func (h *AIHandler) GenerateExplanation(c *gin.Context) {
	var req providers.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效: " + err.Error(),
		})
		return
	}

	response, err := h.aiService.GenerateExplanation(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Generate explanation failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GENERATE_EXPLANATION_FAILED",
			Message: "生成解释失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GenerateTranslation 生成翻译
func (h *AIHandler) GenerateTranslation(c *gin.Context) {
	var req providers.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效: " + err.Error(),
		})
		return
	}

	response, err := h.aiService.GenerateTranslation(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Generate translation failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GENERATE_TRANSLATION_FAILED",
			Message: "生成翻译失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ExtractKeywords 提取关键词
func (h *AIHandler) ExtractKeywords(c *gin.Context) {
	var req providers.AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效: " + err.Error(),
		})
		return
	}

	response, err := h.aiService.ExtractKeywords(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Extract keywords failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "EXTRACT_KEYWORDS_FAILED",
			Message: "提取关键词失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CalculateSimilarity 计算相似度
func (h *AIHandler) CalculateSimilarity(c *gin.Context) {
	var req providers.AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效: " + err.Error(),
		})
		return
	}

	response, err := h.aiService.CalculateSimilarity(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Calculate similarity failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CALCULATE_SIMILARITY_FAILED",
			Message: "计算相似度失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GenerateEmbedding 生成嵌入向量
func (h *AIHandler) GenerateEmbedding(c *gin.Context) {
	var req providers.AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效: " + err.Error(),
		})
		return
	}

	response, err := h.aiService.GenerateEmbedding(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Generate embedding failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GENERATE_EMBEDDING_FAILED",
			Message: "生成嵌入向量失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
