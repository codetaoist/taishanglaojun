package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AIHandler AI
type AIHandler struct {
	aiService *services.AIService
	logger    *zap.Logger
}

// NewAIHandler AI
func NewAIHandler(aiService *services.AIService, logger *zap.Logger) *AIHandler {
	return &AIHandler{
		aiService: aiService,
		logger:    logger,
	}
}

// InterpretWisdom 
// @Summary 
// @Description AI
// @Tags AI
// @Accept json
// @Produce json
// @Param wisdom_id path string true "ID"
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
			Message: "ID",
		})
		return
	}

	// AI
	interpretation, err := h.aiService.InterpretWisdom(c.Request.Context(), wisdomID)
	if err != nil {
		h.logger.Error("Failed to interpret wisdom",
			zap.Error(err),
			zap.String("wisdom_id", wisdomID))

		// HTTP
		if err.Error() == "wisdom not found: "+wisdomID {
			c.JSON(http.StatusNotFound, AIErrorResponse{
				Code:    "WISDOM_NOT_FOUND",
				Message: "",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, AIErrorResponse{
			Code:    "INTERPRETATION_ERROR",
			Message: ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "",
		"data":    interpretation,
	})
}

// RecommendWisdom 
// @Summary 
// @Description 
// @Tags AI
// @Accept json
// @Produce json
// @Param wisdom_id path string true "ID"
// @Param limit query int false "" default(5)
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
			Message: "ID",
		})
		return
	}

	// limit
	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20 // 20
	}

	// AI
	recommendations, err := h.aiService.RecommendRelatedWisdom(c.Request.Context(), wisdomID, limit)
	if err != nil {
		h.logger.Error("Failed to recommend wisdom",
			zap.Error(err),
			zap.String("wisdom_id", wisdomID),
			zap.Int("limit", limit))

		// HTTP
		if err.Error() == "failed to get current wisdom: record not found" {
			c.JSON(http.StatusNotFound, AIErrorResponse{
				Code:    "WISDOM_NOT_FOUND",
				Message: "",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, AIErrorResponse{
			Code:    "RECOMMENDATION_ERROR",
			Message: ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "",
		"data":    recommendations,
		"total":   len(recommendations),
	})
}

// GetAIAnalysis AI
// @Summary AI
// @Description AI
// @Tags AI
// @Accept json
// @Produce json
// @Param wisdom_id path string true "ID"
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
			Message: "ID",
		})
		return
	}

	// 
	interpretation, err := h.aiService.InterpretWisdom(c.Request.Context(), wisdomID)
	if err != nil {
		h.logger.Error("Failed to get AI analysis",
			zap.Error(err),
			zap.String("wisdom_id", wisdomID))

		if err.Error() == "wisdom not found: "+wisdomID {
			c.JSON(http.StatusNotFound, AIErrorResponse{
				Code:    "WISDOM_NOT_FOUND",
				Message: "",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, AIErrorResponse{
			Code:    "ANALYSIS_ERROR",
			Message: "AI: " + err.Error(),
		})
		return
	}

	// 
	recommendations, err := h.aiService.RecommendRelatedWisdom(c.Request.Context(), wisdomID, 3)
	if err != nil {
		h.logger.Warn("Failed to get recommendations for analysis",
			zap.Error(err),
			zap.String("wisdom_id", wisdomID))
		recommendations = []services.WisdomRecommendation{} // 
	}

	// 
	analysis := gin.H{
		"wisdom_id":        interpretation.WisdomID,
		"title":            interpretation.Title,
		"key_points":       interpretation.KeyPoints,
		"modern_relevance": interpretation.ModernRelevance,
		"related_concepts": interpretation.RelatedConcepts,
		"recommendations":  recommendations,
		"analysis_summary": gin.H{
			"total_key_points":       len(interpretation.KeyPoints),
			"total_related_concepts": len(interpretation.RelatedConcepts),
			"total_recommendations":  len(recommendations),
			"has_practical_advice":   interpretation.PracticalAdvice != "",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "AI",
		"data":    analysis,
	})
}

// BatchRecommend 
// @Summary 
// @Description ID
// @Tags AI
// @Accept json
// @Produce json
// @Param request body object{wisdom_ids=[]string,limit=int} true ""
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
			Message: ": " + err.Error(),
		})
		return
	}

	if len(req.WisdomIDs) == 0 {
		c.JSON(http.StatusBadRequest, AIErrorResponse{
			Code:    "EMPTY_WISDOM_IDS",
			Message: "ID",
		})
		return
	}

	if len(req.WisdomIDs) > 10 {
		c.JSON(http.StatusBadRequest, AIErrorResponse{
			Code:    "TOO_MANY_WISDOM_IDS",
			Message: "10ID",
		})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 3
	}
	if req.Limit > 10 {
		req.Limit = 10
	}

	// 
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
		"message": "",
		"data":    results,
	}

	if len(errors) > 0 {
		response["errors"] = errors
		response["message"] = ""
	}

	c.JSON(http.StatusOK, response)
}

// IntelligentQA 
func (h *AIHandler) IntelligentQA(c *gin.Context) {
	var request struct {
		Question string `json:"question" binding:"required"`
		WisdomID string `json:"wisdom_id,omitempty"`
		Context  string `json:"context,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, AIErrorResponse{
			Code:    "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	// 
	if len(request.Question) < 5 {
		c.JSON(http.StatusBadRequest, AIErrorResponse{
			Code:    "Question too short",
			Message: "5",
		})
		return
	}

	if len(request.Question) > 500 {
		c.JSON(http.StatusBadRequest, AIErrorResponse{
			Code:    "Question too long",
			Message: "500",
		})
		return
	}

	// 
	qaRequest := services.QARequest{
		Question: request.Question,
		WisdomID: request.WisdomID,
		Context:  request.Context,
	}

	// AI
	response, err := h.aiService.IntelligentQA(c.Request.Context(), qaRequest)
	if err != nil {
		h.logger.Error("Failed to process intelligent QA", zap.Error(err))
		c.JSON(http.StatusInternalServerError, AIErrorResponse{
			Code:    "QA processing failed",
			Message: "",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// AnalyzeWisdomInDepth 
// @Summary 
// @Description AI
// @Tags AI
// @Accept json
// @Produce json
// @Param wisdom_id path string true "ID"
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
			Message: "ID",
		})
		return
	}

	// AI
	analysis, err := h.aiService.AnalyzeWisdomInDepth(c.Request.Context(), wisdomID)
	if err != nil {
		h.logger.Error("Failed to analyze wisdom in depth",
			zap.Error(err),
			zap.String("wisdom_id", wisdomID))

		// HTTP
		if err.Error() == "wisdom not found: "+wisdomID {
			c.JSON(http.StatusNotFound, AIErrorResponse{
				Code:    "WISDOM_NOT_FOUND",
				Message: "",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, AIErrorResponse{
			Code:    "DEPTH_ANALYSIS_ERROR",
			Message: ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "",
		"data":    analysis,
	})
}

// AIErrorResponse AI
type AIErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

