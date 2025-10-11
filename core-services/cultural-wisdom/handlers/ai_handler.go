package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/services"
	"go.uber.org/zap"
)

// AIHandler AI处理�?
type AIHandler struct {
	aiService *services.AIService
	logger    *zap.Logger
}

// NewAIHandler 创建AI处理�?
func NewAIHandler(aiService *services.AIService, logger *zap.Logger) *AIHandler {
	return &AIHandler{
		aiService: aiService,
		logger:    logger,
	}
}

// InterpretWisdom 解读文化智慧
// @Summary 解读文化智慧
// @Description 使用AI深度解读指定的文化智慧内�?
// @Tags 文化智慧AI
// @Accept json
// @Produce json
// @Param wisdom_id path string true "智慧ID"
// @Success 200 {object} services.WisdomInterpretation
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/cultural-wisdom/{wisdom_id}/interpret [post]
func (h *AIHandler) InterpretWisdom(c *gin.Context) {
	wisdomID := c.Param("wisdom_id")
	if wisdomID == "" {
		h.logger.Error("Missing wisdom_id parameter")
		c.JSON(http.StatusBadRequest, AIErrorResponse{
			Code:    "MISSING_PARAMETER",
			Message: "智慧ID不能为空",
		})
		return
	}

	// 调用AI服务进行解读
	interpretation, err := h.aiService.InterpretWisdom(c.Request.Context(), wisdomID)
	if err != nil {
		h.logger.Error("Failed to interpret wisdom", 
			zap.Error(err), 
			zap.String("wisdom_id", wisdomID))
		
		// 根据错误类型返回不同的HTTP状态码
		if err.Error() == "wisdom not found: "+wisdomID {
			c.JSON(http.StatusNotFound, AIErrorResponse{
				Code:    "WISDOM_NOT_FOUND",
				Message: "指定的智慧内容不存在",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, AIErrorResponse{
			Code:    "INTERPRETATION_ERROR",
			Message: "智慧解读失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "智慧解读成功",
		"data":    interpretation,
	})
}

// RecommendWisdom 推荐相关智慧
// @Summary 推荐相关智慧
// @Description 基于指定智慧推荐相关的文化智慧内�?
// @Tags 文化智慧AI
// @Accept json
// @Produce json
// @Param wisdom_id path string true "智慧ID"
// @Param limit query int false "推荐数量限制" default(5)
// @Success 200 {object} object{code=string,message=string,data=[]services.WisdomRecommendation}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/cultural-wisdom/{wisdom_id}/recommend [get]
func (h *AIHandler) RecommendWisdom(c *gin.Context) {
	wisdomID := c.Param("wisdom_id")
	if wisdomID == "" {
		h.logger.Error("Missing wisdom_id parameter")
		c.JSON(http.StatusBadRequest, AIErrorResponse{
			Code:    "MISSING_PARAMETER",
			Message: "智慧ID不能为空",
		})
		return
	}

	// 解析limit参数
	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20 // 限制最大推荐数�?
	}

	// 调用AI服务进行推荐
	recommendations, err := h.aiService.RecommendRelatedWisdom(c.Request.Context(), wisdomID, limit)
	if err != nil {
		h.logger.Error("Failed to recommend wisdom", 
			zap.Error(err), 
			zap.String("wisdom_id", wisdomID),
			zap.Int("limit", limit))
		
		// 根据错误类型返回不同的HTTP状态码
		if err.Error() == "failed to get current wisdom: record not found" {
			c.JSON(http.StatusNotFound, AIErrorResponse{
				Code:    "WISDOM_NOT_FOUND",
				Message: "指定的智慧内容不存在",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, AIErrorResponse{
			Code:    "RECOMMENDATION_ERROR",
			Message: "智慧推荐失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "智慧推荐成功",
		"data":    recommendations,
		"total":   len(recommendations),
	})
}

// GetAIAnalysis 获取智慧的AI分析摘要
// @Summary 获取AI分析摘要
// @Description 获取指定智慧的AI分析摘要，包括关键概念和现代应用
// @Tags 文化智慧AI
// @Accept json
// @Produce json
// @Param wisdom_id path string true "智慧ID"
// @Success 200 {object} object{code=string,message=string,data=object}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/cultural-wisdom/{wisdom_id}/analysis [get]
func (h *AIHandler) GetAIAnalysis(c *gin.Context) {
	wisdomID := c.Param("wisdom_id")
	if wisdomID == "" {
		h.logger.Error("Missing wisdom_id parameter")
		c.JSON(http.StatusBadRequest, AIErrorResponse{
			Code:    "MISSING_PARAMETER",
			Message: "智慧ID不能为空",
		})
		return
	}

	// 获取智慧解读
	interpretation, err := h.aiService.InterpretWisdom(c.Request.Context(), wisdomID)
	if err != nil {
		h.logger.Error("Failed to get AI analysis", 
			zap.Error(err), 
			zap.String("wisdom_id", wisdomID))
		
		if err.Error() == "wisdom not found: "+wisdomID {
			c.JSON(http.StatusNotFound, AIErrorResponse{
				Code:    "WISDOM_NOT_FOUND",
				Message: "指定的智慧内容不存在",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, AIErrorResponse{
			Code:    "ANALYSIS_ERROR",
			Message: "AI分析失败: " + err.Error(),
		})
		return
	}

	// 获取相关推荐
	recommendations, err := h.aiService.RecommendRelatedWisdom(c.Request.Context(), wisdomID, 3)
	if err != nil {
		h.logger.Warn("Failed to get recommendations for analysis", 
			zap.Error(err), 
			zap.String("wisdom_id", wisdomID))
		recommendations = []services.WisdomRecommendation{} // 使用空数组作为备�?
	}

	// 构建分析摘要
	analysis := gin.H{
		"wisdom_id":        interpretation.WisdomID,
		"title":           interpretation.Title,
		"key_points":      interpretation.KeyPoints,
		"modern_relevance": interpretation.ModernRelevance,
		"related_concepts": interpretation.RelatedConcepts,
		"recommendations":  recommendations,
		"analysis_summary": gin.H{
			"total_key_points":      len(interpretation.KeyPoints),
			"total_related_concepts": len(interpretation.RelatedConcepts),
			"total_recommendations": len(recommendations),
			"has_practical_advice":  interpretation.PracticalAdvice != "",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "AI分析获取成功",
		"data":    analysis,
	})
}

// BatchRecommend 批量推荐智慧
// @Summary 批量推荐智慧
// @Description 基于多个智慧ID批量获取推荐
// @Tags 文化智慧AI
// @Accept json
// @Produce json
// @Param request body object{wisdom_ids=[]string,limit=int} true "批量推荐请求"
// @Success 200 {object} object{code=string,message=string,data=map[string][]services.WisdomRecommendation}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/cultural-wisdom/batch-recommend [post]
func (h *AIHandler) BatchRecommend(c *gin.Context) {
	var req struct {
		WisdomIDs []string `json:"wisdom_ids" binding:"required"`
		Limit     int      `json:"limit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, AIErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效: " + err.Error(),
		})
		return
	}

	if len(req.WisdomIDs) == 0 {
		c.JSON(http.StatusBadRequest, AIErrorResponse{
			Code:    "EMPTY_WISDOM_IDS",
			Message: "智慧ID列表不能为空",
		})
		return
	}

	if len(req.WisdomIDs) > 10 {
		c.JSON(http.StatusBadRequest, AIErrorResponse{
			Code:    "TOO_MANY_WISDOM_IDS",
			Message: "一次最多只能处�?0个智慧ID",
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
	results := make(map[string][]services.WisdomRecommendation)
	errors := make(map[string]string)

	for _, wisdomID := range req.WisdomIDs {
		recommendations, err := h.aiService.RecommendRelatedWisdom(c.Request.Context(), wisdomID, req.Limit)
		if err != nil {
			h.logger.Warn("Failed to get recommendations for wisdom", 
				zap.Error(err), 
				zap.String("wisdom_id", wisdomID))
			errors[wisdomID] = err.Error()
			results[wisdomID] = []services.WisdomRecommendation{}
		} else {
			results[wisdomID] = recommendations
		}
	}

	response := gin.H{
		"code":    "SUCCESS",
		"message": "批量推荐完成",
		"data":    results,
	}

	if len(errors) > 0 {
		response["errors"] = errors
		response["message"] = "批量推荐部分完成，部分智慧推荐失�?
	}

	c.JSON(http.StatusOK, response)
}

// IntelligentQA 智能问答
func (h *AIHandler) IntelligentQA(c *gin.Context) {
	var request struct {
		Question  string `json:"question" binding:"required"`
		WisdomID  string `json:"wisdom_id,omitempty"`
		Context   string `json:"context,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, AIErrorResponse{
			Code:    "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	// 验证问题长度
	if len(request.Question) < 5 {
		c.JSON(http.StatusBadRequest, AIErrorResponse{
			Code:    "Question too short",
			Message: "问题长度至少需�?个字�?,
		})
		return
	}

	if len(request.Question) > 500 {
		c.JSON(http.StatusBadRequest, AIErrorResponse{
			Code:    "Question too long",
			Message: "问题长度不能超过500个字�?,
		})
		return
	}

	// 构建问答请求
	qaRequest := services.QARequest{
		Question: request.Question,
		WisdomID: request.WisdomID,
		Context:  request.Context,
	}

	// 调用AI服务进行问答
	response, err := h.aiService.IntelligentQA(c.Request.Context(), qaRequest)
	if err != nil {
		h.logger.Error("Failed to process intelligent QA", zap.Error(err))
		c.JSON(http.StatusInternalServerError, AIErrorResponse{
			Code:    "QA processing failed",
			Message: "智能问答处理失败，请稍后重试",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// AnalyzeWisdomInDepth 深度分析文化智慧
// @Summary 深度分析文化智慧
// @Description 使用AI对指定的文化智慧进行多维度深度分析，包括情感分析、历史背景、哲学内核和文化影响
// @Tags 文化智慧AI
// @Accept json
// @Produce json
// @Param wisdom_id path string true "智慧ID"
// @Success 200 {object} services.WisdomAnalysis
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/cultural-wisdom/{wisdom_id}/depth-analysis [post]
func (h *AIHandler) AnalyzeWisdomInDepth(c *gin.Context) {
	wisdomID := c.Param("wisdom_id")
	if wisdomID == "" {
		h.logger.Error("Missing wisdom_id parameter")
		c.JSON(http.StatusBadRequest, AIErrorResponse{
			Code:    "MISSING_PARAMETER",
			Message: "智慧ID不能为空",
		})
		return
	}

	// 调用AI服务进行深度分析
	analysis, err := h.aiService.AnalyzeWisdomInDepth(c.Request.Context(), wisdomID)
	if err != nil {
		h.logger.Error("Failed to analyze wisdom in depth", 
			zap.Error(err), 
			zap.String("wisdom_id", wisdomID))
		
		// 根据错误类型返回不同的HTTP状态码
		if err.Error() == "wisdom not found: "+wisdomID {
			c.JSON(http.StatusNotFound, AIErrorResponse{
				Code:    "WISDOM_NOT_FOUND",
				Message: "指定的智慧内容不存在",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, AIErrorResponse{
			Code:    "DEPTH_ANALYSIS_ERROR",
			Message: "智慧深度分析失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "智慧深度分析成功",
		"data":    analysis,
	})
}

// AIErrorResponse AI处理器专用错误响�?
type AIErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
