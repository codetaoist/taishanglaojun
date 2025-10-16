package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/consciousness/engines"
	"github.com/codetaoist/taishanglaojun/core-services/consciousness/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// FusionHandler 洦处理
type FusionHandler struct {
	fusionEngine *engines.FusionEngine
	logger       *zap.Logger
}

// NewFusionHandler 洦
func NewFusionHandler(fusionEngine *engines.FusionEngine, logger *zap.Logger) *FusionHandler {
	return &FusionHandler{
		fusionEngine: fusionEngine,
		logger:       logger,
	}
}

// StartFusionSession
// @Summary
// @Description
// @Tags
// @Accept json
// @Produce json
// @Param request body models.FusionRequest true ""
// @Success 201 {object} models.FusionSession ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/fusion/sessions [post]
func (h *FusionHandler) StartFusionSession(c *gin.Context) {
	var req models.FusionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid fusion request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "", "details": err.Error()})
		return
	}

	// 验证输入数据
	if req.CarbonData == nil && req.SiliconData == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "至少需要提供碳基或硅基输入数据"})
		return
	}

	// 启动融合会话
	sessionID, err := h.fusionEngine.StartFusion(c.Request.Context(), req.CarbonData, req.SiliconData)
	if err != nil {
		h.logger.Error("Failed to start fusion session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "启动融合会话失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"session_id": sessionID})
}

// GetFusionSession
// @Summary
// @Description ID
// @Tags
// @Produce json
// @Param sessionId path string true "ID"
// @Success 200 {object} models.FusionSession ""
// @Failure 404 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/fusion/sessions/{sessionId} [get]
func (h *FusionHandler) GetFusionSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	session, err := h.fusionEngine.GetFusionStatus(sessionID)
	if err != nil {
		h.logger.Error("Failed to get fusion session", zap.String("sessionId", sessionID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, session)
}

// CancelFusionSession
// @Summary
// @Description
// @Tags
// @Produce json
// @Param sessionId path string true "ID"
// @Success 200 {object} map[string]interface{} ""
// @Failure 404 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/fusion/sessions/{sessionId} [delete]
func (h *FusionHandler) CancelFusionSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	err := h.fusionEngine.CancelFusion(sessionID)
	if err != nil {
		h.logger.Error("Failed to cancel fusion session", zap.String("sessionId", sessionID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "", "sessionId": sessionID})
}

// GetFusionHistory
// @Summary
// @Description
// @Tags
// @Produce json
// @Param page query int false "" default(1)
// @Param limit query int false "" default(10)
// @Param strategy query string false ""
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/fusion/history [get]
func (h *FusionHandler) GetFusionHistory(c *gin.Context) {
	//
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	strategy := c.Query("strategy")

	//
	filter := map[string]interface{}{
		"page":   page,
		"limit":  limit,
		"offset": (page - 1) * limit,
	}
	if strategy != "" {
		filter["strategy"] = strategy
	}

	//
	// 洢
	history := map[string]interface{}{
		"sessions": []models.FusionState{},
		"total":    0,
		"page":     page,
		"limit":    limit,
		"hasMore":  false,
	}

	c.JSON(http.StatusOK, history)
}

// GetFusionStrategies
// @Summary
// @Description
// @Tags
// @Produce json
// @Success 200 {object} map[string]interface{} ""
// @Router /consciousness/fusion/strategies [get]
func (h *FusionHandler) GetFusionStrategies(c *gin.Context) {
	strategies := map[string]interface{}{
		"strategies": []map[string]interface{}{
			{
				"name":        "complementary",
				"displayName": "",
				"description": "",
				"suitableFor": []string{"", "", ""},
			},
			{
				"name":        "synergetic",
				"displayName": "",
				"description": "",
				"suitableFor": []string{"", "", ""},
			},
			{
				"name":        "hybrid",
				"displayName": "",
				"description": "",
				"suitableFor": []string{"", "", ""},
			},
			{
				"name":        "transcendent",
				"displayName": "",
				"description": "0",
				"suitableFor": []string{"", "", ""},
			},
		},
	}

	c.JSON(http.StatusOK, strategies)
}

// GetFusionMetrics
// @Summary
// @Description
// @Tags
// @Produce json
// @Success 200 {object} map[string]interface{} ""
// @Router /consciousness/fusion/metrics [get]
func (h *FusionHandler) GetFusionMetrics(c *gin.Context) {
	//
	metrics := map[string]interface{}{
		"totalSessions":   0,
		"activeSessions":  0,
		"successRate":     0.0,
		"averageQuality":  0.0,
		"averageDuration": 0,
		"strategyStats": map[string]interface{}{
			"complementary": map[string]interface{}{
				"count":       0,
				"successRate": 0.0,
				"avgQuality":  0.0,
			},
			"synergetic": map[string]interface{}{
				"count":       0,
				"successRate": 0.0,
				"avgQuality":  0.0,
			},
			"hybrid": map[string]interface{}{
				"count":       0,
				"successRate": 0.0,
				"avgQuality":  0.0,
			},
			"transcendent": map[string]interface{}{
				"count":       0,
				"successRate": 0.0,
				"avgQuality":  0.0,
			},
		},
		"lastUpdated": time.Now(),
	}

	c.JSON(http.StatusOK, metrics)
}
