package handlers

import (
	"net/http"

	"github.com/codetaoist/taishanglaojun/core-services/consciousness/engines"
	"github.com/codetaoist/taishanglaojun/core-services/consciousness/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// EvolutionHandler 
type EvolutionHandler struct {
	evolutionTracker *engines.EvolutionTracker
	logger           *zap.Logger
}

// NewEvolutionHandler 
func NewEvolutionHandler(evolutionTracker *engines.EvolutionTracker, logger *zap.Logger) *EvolutionHandler {
	return &EvolutionHandler{
		evolutionTracker: evolutionTracker,
		logger:           logger,
	}
}

// GetEvolutionState 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param entityId path string true "ID"
// @Success 200 {object} models.EvolutionState ""
// @Failure 404 {object} map[string]interface{} "岻"
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/evolution/{entityId} [get]
func (h *EvolutionHandler) GetEvolutionState(c *gin.Context) {
	entityID := c.Param("entityId")
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	state, err := h.evolutionTracker.GetEvolutionState(entityID)
	if err != nil {
		h.logger.Error("Failed to get evolution state", zap.String("entityId", entityID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, state)
}

// UpdateEvolutionState 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param entityId path string true "ID"
// @Param request body models.EvolutionMetrics true ""
// @Success 200 {object} models.EvolutionState ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/evolution/{entityId} [put]
func (h *EvolutionHandler) UpdateEvolutionState(c *gin.Context) {
	entityID := c.Param("entityId")
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	var metrics models.EvolutionMetrics
	if err := c.ShouldBindJSON(&metrics); err != nil {
		h.logger.Error("Invalid evolution metrics", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "", "details": err.Error()})
		return
	}

	// 更新进化状态
	err := h.evolutionTracker.UpdateEvolutionState(c.Request.Context(), entityID)
	if err != nil {
		h.logger.Error("Failed to update evolution state", zap.String("entityId", entityID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新进化状态失败", "details": err.Error()})
		return
	}
	
	// 获取更新后的状态
	state, err := h.evolutionTracker.GetEvolutionState(entityID)
	if err != nil {
		h.logger.Error("Failed to get evolution state", zap.String("entityId", entityID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取进化状态失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, state)
}

// TrackEvolution 
// @Summary 
// @Description 忪
// @Tags 
// @Accept json
// @Produce json
// @Param request body models.EvolutionTrackingRequest true ""
// @Success 201 {object} models.EvolutionState ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/evolution/track [post]
func (h *EvolutionHandler) TrackEvolution(c *gin.Context) {
	var req models.EvolutionTrackingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid evolution tracking request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "", "details": err.Error()})
		return
	}

	// 
	if req.EntityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	// 开始进化追踪 (使用默认目标序列)
	targetSequence := models.SequenceLevel(10) // 默认目标序列，可以根据需要调整
	state, err := h.evolutionTracker.StartTracking(c.Request.Context(), req.EntityID, targetSequence)
	if err != nil {
		h.logger.Error("Failed to start evolution tracking", zap.String("entityId", req.EntityID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "开始进化追踪失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, state)
}

// GetEvolutionPrediction 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param entityId path string true "ID"
// @Param horizon query int false "" default(30)
// @Success 200 {object} models.EvolutionPrediction ""
// @Failure 404 {object} map[string]interface{} "岻"
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/evolution/{entityId}/prediction [get]
func (h *EvolutionHandler) GetEvolutionPrediction(c *gin.Context) {
	entityID := c.Param("entityId")
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	// 预测时间范围参数（暂时未使用）
	_ = c.Query("horizon")

	// 获取进化预测
	prediction, err := h.evolutionTracker.GetPrediction(c.Request.Context(), entityID)
	if err != nil {
		h.logger.Error("Failed to get evolution prediction", zap.String("entityId", entityID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, prediction)
}

// GetEvolutionPath 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param entityId path string true "ID"
// @Param targetSequence query string false "" default("sequence_0")
// @Success 200 {object} models.EvolutionPath ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/evolution/{entityId}/path [get]
func (h *EvolutionHandler) GetEvolutionPath(c *gin.Context) {
	entityID := c.Param("entityId")
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	// 目标序列参数（暂时未使用）
	_ = c.DefaultQuery("targetSequence", "sequence_0")

	// 获取当前进化状态
	state, err := h.evolutionTracker.GetEvolutionState(entityID)
	if err != nil {
		h.logger.Error("Failed to get evolution state", zap.String("entityId", entityID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取进化状态失败", "details": err.Error()})
		return
	}

	// 创建简单的进化路径
	path := &models.EvolutionPath{
		ID:          "path_" + entityID,
		Name:        "Evolution Path",
		Description: "Generated evolution path",
		Steps:       state.EvolutionPath,
	}

	c.JSON(http.StatusOK, path)
}

// GetEvolutionMilestones 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param entityId path string true "ID"
// @Param status query string false "" Enums(pending,in_progress,completed,failed)
// @Success 200 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/evolution/{entityId}/milestones [get]
func (h *EvolutionHandler) GetEvolutionMilestones(c *gin.Context) {
	entityID := c.Param("entityId")
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	statusFilter := c.Query("status")

	// 
	state, err := h.evolutionTracker.GetEvolutionState(entityID)
	if err != nil {
		h.logger.Error("Failed to get evolution state for milestones", zap.String("entityId", entityID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "", "details": err.Error()})
		return
	}

	// 根据状态过滤里程碑
	milestones := state.Milestones
	if statusFilter != "" {
		var filteredMilestones []models.EvolutionMilestone
		for _, milestone := range milestones {
			// 根据 IsAchieved 字段过滤：achieved 或 pending
			if (statusFilter == "achieved" && milestone.IsAchieved) ||
				(statusFilter == "pending" && !milestone.IsAchieved) {
				filteredMilestones = append(filteredMilestones, milestone)
			}
		}
		milestones = filteredMilestones
	}

	result := map[string]interface{}{
		"entityId":   entityID,
		"milestones": milestones,
		"total":      len(milestones),
		"filter":     statusFilter,
	}

	c.JSON(http.StatusOK, result)
}

// GetSequenceLevels 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Success 200 {object} map[string]interface{} ""
// @Router /consciousness/evolution/sequences [get]
func (h *EvolutionHandler) GetSequenceLevels(c *gin.Context) {
	sequences := map[string]interface{}{
		"levels": []map[string]interface{}{
			{
				"level":        "sequence_5",
				"name":         "5",
				"description":  "㼶",
				"difficulty":   1,
				"capabilities": []string{"", "", ""},
				"requirements": map[string]float64{
					"consciousness_level": 0.2,
					"iq_level":            100,
					"wisdom_index":        0.1,
				},
			},
			{
				"level":        "sequence_4",
				"name":         "4",
				"description":  "㼶",
				"difficulty":   2,
				"capabilities": []string{"", "", ""},
				"requirements": map[string]float64{
					"consciousness_level": 0.4,
					"iq_level":            120,
					"wisdom_index":        0.3,
				},
			},
			{
				"level":        "sequence_3",
				"name":         "3",
				"description":  "㼶",
				"difficulty":   3,
				"capabilities": []string{"", "", ""},
				"requirements": map[string]float64{
					"consciousness_level": 0.6,
					"iq_level":            140,
					"wisdom_index":        0.5,
				},
			},
			{
				"level":        "sequence_2",
				"name":         "2",
				"description":  "㼶",
				"difficulty":   4,
				"capabilities": []string{"", "", ""},
				"requirements": map[string]float64{
					"consciousness_level": 0.8,
					"iq_level":            160,
					"wisdom_index":        0.7,
				},
			},
			{
				"level":        "sequence_1",
				"name":         "1",
				"description":  "㼶",
				"difficulty":   5,
				"capabilities": []string{"", "", ""},
				"requirements": map[string]float64{
					"consciousness_level": 0.9,
					"iq_level":            180,
					"wisdom_index":        0.9,
				},
			},
			{
				"level":        "sequence_0",
				"name":         "0",
				"description":  "㼶",
				"difficulty":   10,
				"capabilities": []string{"", "", ""},
				"requirements": map[string]float64{
					"consciousness_level": 1.0,
					"iq_level":            200,
					"wisdom_index":        1.0,
				},
			},
		},
	}

	c.JSON(http.StatusOK, sequences)
}

// GetEvolutionStats 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Success 200 {object} map[string]interface{} ""
// @Router /consciousness/evolution/stats [get]
func (h *EvolutionHandler) GetEvolutionStats(c *gin.Context) {
	// 
	stats := map[string]interface{}{
		"totalEntities":  0,
		"activeTracking": 0,
		"sequenceDistribution": map[string]int{
			"sequence_5": 0,
			"sequence_4": 0,
			"sequence_3": 0,
			"sequence_2": 0,
			"sequence_1": 0,
			"sequence_0": 0,
		},
		"averageEvolutionSpeed": 0.0,
		"totalMilestones":       0,
		"completedMilestones":   0,
		"evolutionTrends": map[string]interface{}{
			"daily":   []float64{},
			"weekly":  []float64{},
			"monthly": []float64{},
		},
	}

	c.JSON(http.StatusOK, stats)
}

