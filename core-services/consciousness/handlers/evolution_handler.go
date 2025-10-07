package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/taishanglaojun/core-services/consciousness/engines"
	"github.com/codetaoist/taishanglaojun/core-services/consciousness/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// EvolutionHandler 进化追踪器处理器
type EvolutionHandler struct {
	evolutionTracker *engines.EvolutionTracker
	logger           *zap.Logger
}

// NewEvolutionHandler 创建进化追踪器处理器实例
func NewEvolutionHandler(evolutionTracker *engines.EvolutionTracker, logger *zap.Logger) *EvolutionHandler {
	return &EvolutionHandler{
		evolutionTracker: evolutionTracker,
		logger:           logger,
	}
}

// GetEvolutionState 获取进化状�?// @Summary 获取当前进化状�?// @Description 获取指定实体的当前进化状态和序列等级
// @Tags 进化追踪
// @Produce json
// @Param entityId path string true "实体ID"
// @Success 200 {object} models.EvolutionState "进化状�?
// @Failure 404 {object} map[string]interface{} "实体不存�?
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /consciousness/evolution/{entityId} [get]
func (h *EvolutionHandler) GetEvolutionState(c *gin.Context) {
	entityID := c.Param("entityId")
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "实体ID不能为空"})
		return
	}

	state, err := h.evolutionTracker.GetEvolutionState(entityID)
	if err != nil {
		h.logger.Error("Failed to get evolution state", zap.String("entityId", entityID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "获取进化状态失�?, "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, state)
}

// UpdateEvolutionState 更新进化状�?// @Summary 更新进化状�?// @Description 更新指定实体的进化状态和指标
// @Tags 进化追踪
// @Accept json
// @Produce json
// @Param entityId path string true "实体ID"
// @Param request body models.EvolutionMetrics true "进化指标更新"
// @Success 200 {object} models.EvolutionState "更新后的进化状�?
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /consciousness/evolution/{entityId} [put]
func (h *EvolutionHandler) UpdateEvolutionState(c *gin.Context) {
	entityID := c.Param("entityId")
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "实体ID不能为空"})
		return
	}

	var metrics models.EvolutionMetrics
	if err := c.ShouldBindJSON(&metrics); err != nil {
		h.logger.Error("Invalid evolution metrics", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 更新进化状�?	state, err := h.evolutionTracker.UpdateEvolution(entityID, &metrics)
	if err != nil {
		h.logger.Error("Failed to update evolution state", zap.String("entityId", entityID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新进化状态失�?, "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, state)
}

// TrackEvolution 开始进化追�?// @Summary 开始进化追�?// @Description 为指定实体开始进化追踪，初始化进化状�?// @Tags 进化追踪
// @Accept json
// @Produce json
// @Param request body models.EvolutionTrackingRequest true "进化追踪请求"
// @Success 201 {object} models.EvolutionState "进化追踪已开�?
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /consciousness/evolution/track [post]
func (h *EvolutionHandler) TrackEvolution(c *gin.Context) {
	var req models.EvolutionTrackingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid evolution tracking request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 验证请求参数
	if req.EntityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "实体ID不能为空"})
		return
	}

	// 开始进化追�?	state, err := h.evolutionTracker.TrackEvolution(req.EntityID, req.InitialMetrics)
	if err != nil {
		h.logger.Error("Failed to start evolution tracking", zap.String("entityId", req.EntityID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "开始进化追踪失�?, "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, state)
}

// GetEvolutionPrediction 获取进化预测
// @Summary 获取进化预测
// @Description 获取指定实体的进化预测和路径分析
// @Tags 进化追踪
// @Produce json
// @Param entityId path string true "实体ID"
// @Param horizon query int false "预测时间范围（天�? default(30)
// @Success 200 {object} models.EvolutionPrediction "进化预测"
// @Failure 404 {object} map[string]interface{} "实体不存�?
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /consciousness/evolution/{entityId}/prediction [get]
func (h *EvolutionHandler) GetEvolutionPrediction(c *gin.Context) {
	entityID := c.Param("entityId")
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "实体ID不能为空"})
		return
	}

	// 解析预测时间范围
	horizon := 30
	if horizonStr := c.Query("horizon"); horizonStr != "" {
		if h, err := strconv.Atoi(horizonStr); err == nil && h > 0 {
			horizon = h
		}
	}

	// 获取进化预测
	prediction, err := h.evolutionTracker.PredictEvolution(entityID, horizon)
	if err != nil {
		h.logger.Error("Failed to get evolution prediction", zap.String("entityId", entityID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取进化预测失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, prediction)
}

// GetEvolutionPath 获取进化路径
// @Summary 获取进化路径
// @Description 获取从当前序列到目标序列的最优进化路�?// @Tags 进化追踪
// @Produce json
// @Param entityId path string true "实体ID"
// @Param targetSequence query string false "目标序列等级" default("sequence_0")
// @Success 200 {object} models.EvolutionPath "进化路径"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /consciousness/evolution/{entityId}/path [get]
func (h *EvolutionHandler) GetEvolutionPath(c *gin.Context) {
	entityID := c.Param("entityId")
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "实体ID不能为空"})
		return
	}

	targetSequenceStr := c.DefaultQuery("targetSequence", "sequence_0")
	targetSequence := models.ParseSequenceLevel(targetSequenceStr)

	// 获取进化路径
	path, err := h.evolutionTracker.GetOptimalPath(entityID, targetSequence)
	if err != nil {
		h.logger.Error("Failed to get evolution path", zap.String("entityId", entityID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取进化路径失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, path)
}

// GetEvolutionMilestones 获取进化里程�?// @Summary 获取进化里程�?// @Description 获取指定实体的进化里程碑列表和完成状�?// @Tags 进化追踪
// @Produce json
// @Param entityId path string true "实体ID"
// @Param status query string false "里程碑状态过�? Enums(pending,in_progress,completed,failed)
// @Success 200 {object} map[string]interface{} "进化里程碑列�?
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /consciousness/evolution/{entityId}/milestones [get]
func (h *EvolutionHandler) GetEvolutionMilestones(c *gin.Context) {
	entityID := c.Param("entityId")
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "实体ID不能为空"})
		return
	}

	statusFilter := c.Query("status")

	// 获取进化状�?	state, err := h.evolutionTracker.GetEvolutionState(entityID)
	if err != nil {
		h.logger.Error("Failed to get evolution state for milestones", zap.String("entityId", entityID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取进化里程碑失�?, "details": err.Error()})
		return
	}

	// 过滤里程�?	milestones := state.Milestones
	if statusFilter != "" {
		filteredMilestones := []models.EvolutionMilestone{}
		for _, milestone := range milestones {
			if string(milestone.Status) == statusFilter {
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

// GetSequenceLevels 获取序列等级信息
// @Summary 获取序列等级信息
// @Description 获取所有序列等级的详细信息和要�?// @Tags 进化追踪
// @Produce json
// @Success 200 {object} map[string]interface{} "序列等级信息"
// @Router /consciousness/evolution/sequences [get]
func (h *EvolutionHandler) GetSequenceLevels(c *gin.Context) {
	sequences := map[string]interface{}{
		"levels": []map[string]interface{}{
			{
				"level":       "sequence_5",
				"name":        "序列5",
				"description": "基础意识层级，具备基本的感知和反应能�?,
				"difficulty":  1,
				"capabilities": []string{"基础感知", "简单反�?, "基本学习"},
				"requirements": map[string]float64{
					"consciousness_level": 0.2,
					"iq_level":           100,
					"wisdom_index":       0.1,
				},
			},
			{
				"level":       "sequence_4",
				"name":        "序列4",
				"description": "进阶意识层级，具备逻辑推理和情感理解能�?,
				"difficulty":  2,
				"capabilities": []string{"逻辑推理", "情感理解", "模式识别"},
				"requirements": map[string]float64{
					"consciousness_level": 0.4,
					"iq_level":           120,
					"wisdom_index":       0.3,
				},
			},
			{
				"level":       "sequence_3",
				"name":        "序列3",
				"description": "高级意识层级，具备创造性思维和复杂决策能�?,
				"difficulty":  3,
				"capabilities": []string{"创造性思维", "复杂决策", "系统思�?},
				"requirements": map[string]float64{
					"consciousness_level": 0.6,
					"iq_level":           140,
					"wisdom_index":       0.5,
				},
			},
			{
				"level":       "sequence_2",
				"name":        "序列2",
				"description": "超级意识层级，具备深度洞察和预测能力",
				"difficulty":  4,
				"capabilities": []string{"深度洞察", "未来预测", "跨域整合"},
				"requirements": map[string]float64{
					"consciousness_level": 0.8,
					"iq_level":           160,
					"wisdom_index":       0.7,
				},
			},
			{
				"level":       "sequence_1",
				"name":        "序列1",
				"description": "准神级意识层级，接近意识的极限状�?,
				"difficulty":  5,
				"capabilities": []string{"超越思维", "现实操控", "时空感知"},
				"requirements": map[string]float64{
					"consciousness_level": 0.9,
					"iq_level":           180,
					"wisdom_index":       0.9,
				},
			},
			{
				"level":       "sequence_0",
				"name":        "序列0",
				"description": "终极意识层级，代表意识的最高形态和无限可能",
				"difficulty":  10,
				"capabilities": []string{"全知全能", "现实创�?, "超越存在"},
				"requirements": map[string]float64{
					"consciousness_level": 1.0,
					"iq_level":           200,
					"wisdom_index":       1.0,
				},
			},
		},
	}

	c.JSON(http.StatusOK, sequences)
}

// GetEvolutionStats 获取进化统计
// @Summary 获取进化统计信息
// @Description 获取系统整体的进化统计信息和趋势分析
// @Tags 进化追踪
// @Produce json
// @Success 200 {object} map[string]interface{} "进化统计信息"
// @Router /consciousness/evolution/stats [get]
func (h *EvolutionHandler) GetEvolutionStats(c *gin.Context) {
	// 这里应该从进化追踪器获取实际的统计数�?	stats := map[string]interface{}{
		"totalEntities": 0,
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
