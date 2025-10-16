package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/taishanglaojun/core-services/consciousness/coordinators"
	"github.com/codetaoist/taishanglaojun/core-services/consciousness/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CoordinationHandler
type CoordinationHandler struct {
	coordinator *coordinators.ThreeAxisCoordinator
	logger      *zap.Logger
}

// NewCoordinationHandler
func NewCoordinationHandler(coordinator *coordinators.ThreeAxisCoordinator, logger *zap.Logger) *CoordinationHandler {
	return &CoordinationHandler{
		coordinator: coordinator,
		logger:      logger,
	}
}

// StartCoordination
// @Summary
// @Description S-C-T
// @Tags
// @Accept json
// @Produce json
// @Param request body models.CoordinationRequest true ""
// @Success 201 {object} models.CoordinationSession ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/coordination/sessions [post]
func (h *CoordinationHandler) StartCoordination(c *gin.Context) {
	var req models.CoordinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid coordination request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "", "details": err.Error()})
		return
	}

	//
	if req.EntityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	//
	session, err := h.coordinator.CoordinateThreeAxis(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to start coordination session", zap.String("entityId", req.EntityID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, session)
}

// GetCoordinationSession
// @Summary
// @Description ID
// @Tags
// @Produce json
// @Param sessionId path string true "ID"
// @Success 200 {object} models.CoordinationSession ""
// @Failure 404 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/coordination/sessions/{sessionId} [get]
func (h *CoordinationHandler) GetCoordinationSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	session, err := h.coordinator.GetCoordinationStatus(sessionID)
	if err != nil {
		h.logger.Error("Failed to get coordination session", zap.String("sessionId", sessionID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, session)
}

// StopCoordination
// @Summary
// @Description
// @Tags
// @Produce json
// @Param sessionId path string true "ID"
// @Success 200 {object} map[string]interface{} ""
// @Failure 404 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/coordination/sessions/{sessionId} [delete]
func (h *CoordinationHandler) StopCoordination(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	err := h.coordinator.Stop()
	if err != nil {
		h.logger.Error("Failed to stop coordination session", zap.String("sessionId", sessionID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "", "sessionId": sessionID})
}

// ProcessSAxis S
// @Summary S
// @Description S
// @Tags
// @Accept json
// @Produce json
// @Param request body models.SequenceRequest true ""
// @Success 200 {object} models.SequenceResponse "S"
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/coordination/s-axis [post]
func (h *CoordinationHandler) ProcessSAxis(c *gin.Context) {
	var req models.SequenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid S-axis request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "", "details": err.Error()})
		return
	}

	// 处理S轴
	response, err := h.coordinator.ProcessSequenceAxis(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to process S-axis", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "S", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ProcessCAxis C
// @Summary C
// @Description C
// @Tags
// @Accept json
// @Produce json
// @Param request body models.CompositionRequest true ""
// @Success 200 {object} models.CompositionResponse "C"
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/coordination/c-axis [post]
func (h *CoordinationHandler) ProcessCAxis(c *gin.Context) {
	var req models.CompositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid C-axis request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "", "details": err.Error()})
		return
	}

	// 处理C轴
	response, err := h.coordinator.ProcessCompositionAxis(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to process C-axis", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "C", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ProcessTAxis T
// @Summary T
// @Description T
// @Tags
// @Accept json
// @Produce json
// @Param request body models.ThoughtRequest true ""
// @Success 200 {object} models.ThoughtResponse "T"
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/coordination/t-axis [post]
func (h *CoordinationHandler) ProcessTAxis(c *gin.Context) {
	var req models.ThoughtRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid T-axis request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "", "details": err.Error()})
		return
	}

	// 处理T轴
	response, err := h.coordinator.ProcessThoughtAxis(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to process T-axis", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process T axis", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// OptimizeBalance
// @Summary
// @Description S-C-T
// @Tags
// @Accept json
// @Produce json
// @Param request body models.BalanceOptimizationRequest true ""
// @Success 200 {object} models.BalanceOptimizationResult ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/coordination/balance [post]
func (h *CoordinationHandler) OptimizeBalance(c *gin.Context) {
	var req models.BalanceOptimizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid balance optimization request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "", "details": err.Error()})
		return
	}

	// 优化平衡
	result, err := h.coordinator.OptimizeBalance(c.Request.Context(), &req.Coordinate, req.Constraints)
	if err != nil {
		h.logger.Error("Failed to optimize balance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "优化平衡失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CatalyzeSynergy
// @Summary
// @Description S-C-T
// @Tags
// @Accept json
// @Produce json
// @Param request body models.SynergyCatalysisRequest true ""
// @Success 200 {object} models.SynergyCatalysisResult ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/coordination/synergy [post]
func (h *CoordinationHandler) CatalyzeSynergy(c *gin.Context) {
	var req models.SynergyCatalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid synergy catalysis request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "", "details": err.Error()})
		return
	}

	// 催化协同
	result, err := h.coordinator.CatalyzeSynergy(c.Request.Context(), &req.Coordinate)
	if err != nil {
		h.logger.Error("Failed to catalyze synergy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "催化协同失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetCoordinationHistory
// @Summary
// @Description
// @Tags
// @Produce json
// @Param entityId query string false "ID"
// @Param page query int false "" default(1)
// @Param limit query int false "" default(10)
// @Param status query string false "" Enums(active,completed,failed,cancelled)
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/coordination/history [get]
func (h *CoordinationHandler) GetCoordinationHistory(c *gin.Context) {
	//
	entityID := c.Query("entityId")
	status := c.Query("status")

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

	//
	filter := map[string]interface{}{
		"page":   page,
		"limit":  limit,
		"offset": (page - 1) * limit,
	}
	if entityID != "" {
		filter["entityId"] = entityID
	}
	if status != "" {
		filter["status"] = status
	}

	// 获取协调历史
	history := h.coordinator.GetCoordinationHistory(limit)

	// 根据过滤条件进行简单过滤（实际应用中可能需要更复杂的过滤逻辑）
	var filteredHistory []models.CoordinationRecord
	for _, record := range history {
		// 这里可以添加更多的过滤逻辑
		filteredHistory = append(filteredHistory, record)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  filteredHistory,
		"total": len(filteredHistory),
		"page":  page,
		"limit": limit,
	})
}

// GetAxisInfo
// @Summary
// @Description S-C-T
// @Tags
// @Produce json
// @Success 200 {object} map[string]interface{} ""
// @Router /consciousness/coordination/axes [get]
func (h *CoordinationHandler) GetAxisInfo(c *gin.Context) {
	axesInfo := map[string]interface{}{
		"axes": []map[string]interface{}{
			{
				"name":        "S-Axis",
				"fullName":    "Sequence Axis",
				"displayName": "",
				"description": "50",
				"levels": []map[string]interface{}{
					{"level": "sequence_5", "name": "5", "description": "㼶"},
					{"level": "sequence_4", "name": "4", "description": "㼶"},
					{"level": "sequence_3", "name": "3", "description": "㼶"},
					{"level": "sequence_2", "name": "2", "description": "㼶"},
					{"level": "sequence_1", "name": "1", "description": "㼶"},
					{"level": "sequence_0", "name": "0", "description": "㼶"},
				},
				"capabilities": []string{"", "", "", "滮"},
			},
			{
				"name":        "C-Axis",
				"fullName":    "Composition Axis",
				"displayName": "",
				"description": "",
				"layers": []map[string]interface{}{
					{"layer": "basic", "name": "", "description": ""},
					{"layer": "intermediate", "name": "", "description": ""},
					{"layer": "advanced", "name": "", "description": ""},
					{"layer": "master", "name": "", "description": ""},
					{"layer": "transcendent", "name": "", "description": ""},
				},
				"capabilities": []string{"", "", "", ""},
			},
			{
				"name":        "T-Axis",
				"fullName":    "Thought Axis",
				"displayName": "",
				"description": "",
				"realms": []map[string]interface{}{
					{"realm": "mundane", "name": "", "description": ""},
					{"realm": "enlightened", "name": "", "description": ""},
					{"realm": "transcendent", "name": "", "description": ""},
					{"realm": "divine", "name": "", "description": ""},
					{"realm": "absolute", "name": "", "description": ""},
				},
				"capabilities": []string{"", "", "", ""},
			},
		},
		"coordination": map[string]interface{}{
			"principles": []string{
				"S-C-T",
				"1+1+1>3",
				"",
				"",
			},
			"strategies": []string{
				"",
				"㷢",
				"",
				"",
			},
		},
	}

	c.JSON(http.StatusOK, axesInfo)
}

// GetCoordinationMetrics
// @Summary 获取协调指标
// @Description 获取当前协调系统的统计指标
// @Tags 协调
// @Produce json
// @Param entityId query string false "ID"
// @Success 200 {object} map[string]interface{} ""
// @Router /consciousness/coordination/metrics [get]
func (h *CoordinationHandler) GetCoordinationMetrics(c *gin.Context) {
	// 获取协调指标
	metrics := h.coordinator.GetStats()

	c.JSON(http.StatusOK, metrics)
}
