package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/advanced"
	"github.com/gin-gonic/gin"
)

// AdvancedAIHandler AI
type AdvancedAIHandler struct {
	service *advanced.AdvancedAIService
}

// NewAdvancedAIHandler 创建AI
func NewAdvancedAIHandler(service *advanced.AdvancedAIService) *AdvancedAIHandler {
	return &AdvancedAIHandler{
		service: service,
	}
}

// RegisterRoutes 注册路由
func (h *AdvancedAIHandler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1/advanced-ai")
	{
		//
		v1.POST("/process", h.ProcessRequest)
		v1.POST("/agi/task", h.ProcessAGITask)
		v1.POST("/meta-learning/learn", h.MetaLearning)
		v1.POST("/evolution/optimize", h.TriggerEvolution)

		//
		v1.GET("/status", h.GetSystemStatus)
		v1.GET("/metrics", h.GetPerformanceMetrics)
		v1.GET("/health", h.HealthCheck)

		//
		v1.GET("/config", h.GetConfiguration)
		v1.PUT("/config", h.UpdateConfiguration)

		//
		v1.GET("/capabilities", h.GetCapabilities)
		v1.POST("/capabilities/:capability/enable", h.EnableCapability)
		v1.POST("/capabilities/:capability/disable", h.DisableCapability)

		//
		v1.GET("/history", h.GetRequestHistory)
		v1.GET("/statistics", h.GetStatistics)

		//
		v1.POST("/initialize", h.Initialize)
		v1.POST("/shutdown", h.Shutdown)
		v1.POST("/reset", h.Reset)
	}
}

// ProcessRequest 处理请求
func (h *AdvancedAIHandler) ProcessRequest(c *gin.Context) {
	var request advanced.AIRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// ID
	if request.ID == "" {
		request.ID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	//
	response, err := h.service.ProcessRequest(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process request",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ProcessAGITask 处理AGI任务
func (h *AdvancedAIHandler) ProcessAGITask(c *gin.Context) {
	var taskRequest struct {
		Type         string                 `json:"type" binding:"required"`
		Input        map[string]interface{} `json:"input" binding:"required"`
		Context      map[string]interface{} `json:"context"`
		Requirements map[string]interface{} `json:"requirements"`
		Priority     int                    `json:"priority"`
		Timeout      int                    `json:"timeout"` //
	}

	if err := c.ShouldBindJSON(&taskRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid task request",
			"details": err.Error(),
		})
		return
	}

	// AGI
	request := &advanced.AIRequest{
		ID:           fmt.Sprintf("agi_task_%d", time.Now().UnixNano()),
		Type:         taskRequest.Type,
		Capability:   advanced.CapabilityAGI,
		Input:        taskRequest.Input,
		Context:      taskRequest.Context,
		Requirements: taskRequest.Requirements,
		Priority:     taskRequest.Priority,
		Timeout:      time.Duration(taskRequest.Timeout) * time.Second,
	}

	response, err := h.service.ProcessRequest(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process AGI task",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// MetaLearning 元学习
func (h *AdvancedAIHandler) MetaLearning(c *gin.Context) {
	var learningRequest struct {
		TaskType   string                   `json:"task_type" binding:"required"`
		Domain     string                   `json:"domain" binding:"required"`
		Data       []map[string]interface{} `json:"data" binding:"required"`
		Strategy   string                   `json:"strategy"`
		Parameters map[string]interface{}   `json:"parameters"`
	}

	if err := c.ShouldBindJSON(&learningRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid learning request",
			"details": err.Error(),
		})
		return
	}

	// MetaLearning
	request := &advanced.AIRequest{
		ID:         fmt.Sprintf("meta_learn_%d", time.Now().UnixNano()),
		Type:       learningRequest.TaskType,
		Capability: advanced.CapabilityMetaLearning,
		Input: map[string]interface{}{
			"data": learningRequest.Data,
		},
		Context: map[string]interface{}{
			"domain":   learningRequest.Domain,
			"strategy": learningRequest.Strategy,
		},
		Requirements: learningRequest.Parameters,
	}

	response, err := h.service.ProcessRequest(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process meta-learning request",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// TriggerEvolution 触发进化
func (h *AdvancedAIHandler) TriggerEvolution(c *gin.Context) {
	var evolutionRequest struct {
		OptimizationTargets []map[string]interface{} `json:"optimization_targets"`
		Strategy            string                   `json:"strategy"`
		Parameters          map[string]interface{}   `json:"parameters"`
	}

	if err := c.ShouldBindJSON(&evolutionRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid evolution request",
			"details": err.Error(),
		})
		return
	}

	// SelfEvolution
	request := &advanced.AIRequest{
		ID:         fmt.Sprintf("evolution_%d", time.Now().UnixNano()),
		Type:       "optimization",
		Capability: advanced.CapabilitySelfEvolution,
		Input: map[string]interface{}{
			"optimization_targets": evolutionRequest.OptimizationTargets,
		},
		Context: map[string]interface{}{
			"strategy": evolutionRequest.Strategy,
		},
		Requirements: evolutionRequest.Parameters,
	}

	response, err := h.service.ProcessRequest(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to trigger evolution",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetSystemStatus 获取系统状态
func (h *AdvancedAIHandler) GetSystemStatus(c *gin.Context) {
	status := h.service.GetSystemStatus()
	c.JSON(http.StatusOK, status)
}

// GetPerformanceMetrics 获取性能指标
func (h *AdvancedAIHandler) GetPerformanceMetrics(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	metrics := h.service.GetPerformanceMetrics(limit)
	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
		"count":   len(metrics),
	})
}

// HealthCheck 健康检查
func (h *AdvancedAIHandler) HealthCheck(c *gin.Context) {
	status := h.service.GetSystemStatus()

	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"uptime":    time.Since(time.Now().Add(-time.Hour)), //
	}

	//
	if status.OverallHealth < 0.5 {
		health["status"] = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, health)
		return
	} else if status.OverallHealth < 0.8 {
		health["status"] = "degraded"
	}

	health["overall_health"] = status.OverallHealth
	health["active_requests"] = status.ActiveRequests
	health["success_rate"] = status.SuccessRate

	c.JSON(http.StatusOK, health)
}

// GetConfiguration 获取配置
func (h *AdvancedAIHandler) GetConfiguration(c *gin.Context) {
	//
	config := gin.H{
		"enable_agi":              true,
		"enable_meta_learning":    true,
		"enable_evolution":        true,
		"max_concurrent_requests": 50,
		"default_timeout":         "30s",
		"performance_monitoring":  true,
		"auto_optimization":       true,
		"log_level":               "info",
	}

	c.JSON(http.StatusOK, config)
}

// UpdateConfiguration 更新配置
func (h *AdvancedAIHandler) UpdateConfiguration(c *gin.Context) {
	var configMap map[string]interface{}
	if err := c.ShouldBindJSON(&configMap); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid configuration",
			"details": err.Error(),
		})
		return
	}

	// 将map转换为AdvancedAIConfig结构体
	config := &advanced.AdvancedAIConfig{}
	if enableAGI, ok := configMap["enable_agi"].(bool); ok {
		config.EnableAGI = enableAGI
	}
	if enableMetaLearning, ok := configMap["enable_meta_learning"].(bool); ok {
		config.EnableMetaLearning = enableMetaLearning
	}
	if enableEvolution, ok := configMap["enable_evolution"].(bool); ok {
		config.EnableEvolution = enableEvolution
	}
	if maxConcurrent, ok := configMap["max_concurrent_requests"].(float64); ok {
		config.MaxConcurrentRequests = int(maxConcurrent)
	}
	if logLevel, ok := configMap["log_level"].(string); ok {
		config.LogLevel = logLevel
	}

	//
	err := h.service.UpdateConfiguration(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update configuration",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration updated successfully",
		"config":  config,
	})
}

// GetCapabilities 获取能力
func (h *AdvancedAIHandler) GetCapabilities(c *gin.Context) {
	capabilities := gin.H{
		"agi": gin.H{
			"enabled":      true,
			"capabilities": []string{"reasoning", "planning", "learning", "creativity", "multimodal", "metacognition"},
			"status":       "active",
		},
		"meta_learning": gin.H{
			"enabled":    true,
			"strategies": []string{"gradient_based", "model_agnostic", "memory_augmented", "few_shot", "transfer_learning", "online_adaptation"},
			"status":     "active",
		},
		"self_evolution": gin.H{
			"enabled":    true,
			"strategies": []string{"genetic", "neuro_evolution", "gradient_free", "hybrid", "reinforcement", "swarm_intelligence"},
			"status":     "active",
		},
		"hybrid": gin.H{
			"enabled": true,
			"status":  "active",
		},
	}

	c.JSON(http.StatusOK, capabilities)
}

// EnableCapability 启用能力
func (h *AdvancedAIHandler) EnableCapability(c *gin.Context) {
	capability := c.Param("capability")

	//
	err := h.service.EnableCapability(advanced.AdvancedAICapability(capability))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to enable capability",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    fmt.Sprintf("Capability '%s' enabled successfully", capability),
		"capability": capability,
		"enabled":    true,
	})
}

// DisableCapability 禁用能力
func (h *AdvancedAIHandler) DisableCapability(c *gin.Context) {
	capability := c.Param("capability")

	//
	err := h.service.DisableCapability(advanced.AdvancedAICapability(capability))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to disable capability",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    fmt.Sprintf("Capability '%s' disabled successfully", capability),
		"capability": capability,
		"enabled":    false,
	})
}

// GetRequestHistory 获取请求历史
func (h *AdvancedAIHandler) GetRequestHistory(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	capabilityFilter := c.Query("capability")
	statusFilter := c.Query("status")

	//
	history := []gin.H{
		{
			"request_id":        "req_123456789",
			"capability":        "agi",
			"type":              "reasoning",
			"success":           true,
			"confidence":        0.95,
			"process_time":      "1.2s",
			"used_capabilities": []string{"agi"},
			"created_at":        time.Now().Add(-time.Hour),
		},
		{
			"request_id":        "req_123456790",
			"capability":        "meta_learning",
			"type":              "adaptation",
			"success":           true,
			"confidence":        0.88,
			"process_time":      "2.5s",
			"used_capabilities": []string{"meta_learning"},
			"created_at":        time.Now().Add(-time.Minute * 30),
		},
	}

	//
	filteredHistory := make([]gin.H, 0)
	for _, item := range history {
		if capabilityFilter != "" && item["capability"] != capabilityFilter {
			continue
		}
		if statusFilter != "" {
			success := item["success"].(bool)
			if (statusFilter == "success" && !success) || (statusFilter == "failed" && success) {
				continue
			}
		}
		filteredHistory = append(filteredHistory, item)
		if len(filteredHistory) >= limit {
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"history": filteredHistory,
		"count":   len(filteredHistory),
		"total":   len(history),
	})
}

// GetStatistics 获取统计信息
func (h *AdvancedAIHandler) GetStatistics(c *gin.Context) {
	status := h.service.GetSystemStatus()

	statistics := gin.H{
		"total_requests":    status.TotalRequests,
		"success_rate":      status.SuccessRate,
		"avg_response_time": status.AvgResponseTime,
		"active_requests":   status.ActiveRequests,
		"overall_health":    status.OverallHealth,
		"capability_stats": gin.H{
			"agi": gin.H{
				"requests":       int64(float64(status.TotalRequests) * 0.4),
				"success_rate":   0.92,
				"avg_confidence": 0.89,
			},
			"meta_learning": gin.H{
				"requests":       int64(float64(status.TotalRequests) * 0.3),
				"success_rate":   0.88,
				"avg_confidence": 0.85,
			},
			"self_evolution": gin.H{
				"requests":       int64(float64(status.TotalRequests) * 0.2),
				"success_rate":   0.95,
				"avg_confidence": 0.91,
			},
			"hybrid": gin.H{
				"requests":       int64(float64(status.TotalRequests) * 0.1),
				"success_rate":   0.94,
				"avg_confidence": 0.93,
			},
		},
		"performance_trends": gin.H{
			"last_hour": gin.H{
				"requests":     120,
				"success_rate": 0.95,
				"avg_latency":  "1.2s",
			},
			"last_day": gin.H{
				"requests":     2880,
				"success_rate": 0.93,
				"avg_latency":  "1.1s",
			},
			"last_week": gin.H{
				"requests":     20160,
				"success_rate": 0.91,
				"avg_latency":  "1.3s",
			},
		},
		"resource_usage": gin.H{
			"cpu":    "65%",
			"memory": "72%",
			"gpu":    "80%",
			"disk":   "45%",
		},
		"generated_at": time.Now(),
	}

	c.JSON(http.StatusOK, statistics)
}

// Initialize 初始化服务
func (h *AdvancedAIHandler) Initialize(c *gin.Context) {
	var initRequest struct {
		Config map[string]interface{} `json:"config"`
		Force  bool                   `json:"force"`
	}

	if err := c.ShouldBindJSON(&initRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid initialization request",
			"details": err.Error(),
		})
		return
	}

	//
	err := h.service.Initialize(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to initialize service",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Service initialized successfully",
		"initialized": true,
		"timestamp":   time.Now(),
	})
}

// Shutdown 关闭服务
func (h *AdvancedAIHandler) Shutdown(c *gin.Context) {
	var shutdownRequest struct {
		Graceful bool `json:"graceful"`
		Timeout  int  `json:"timeout"` //
	}

	if err := c.ShouldBindJSON(&shutdownRequest); err != nil {
		shutdownRequest.Graceful = true
		shutdownRequest.Timeout = 30
	}

	//
	err := h.service.Shutdown(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to shutdown service",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Service shutdown initiated",
		"graceful":  shutdownRequest.Graceful,
		"timeout":   shutdownRequest.Timeout,
		"timestamp": time.Now(),
	})
}

// Reset 重置服务
func (h *AdvancedAIHandler) Reset(c *gin.Context) {
	var resetRequest struct {
		ResetData    bool `json:"reset_data"`
		ResetConfig  bool `json:"reset_config"`
		ResetMetrics bool `json:"reset_metrics"`
	}

	if err := c.ShouldBindJSON(&resetRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid reset request",
			"details": err.Error(),
		})
		return
	}

	//
	err := h.service.Reset()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to reset service",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Service reset successfully",
		"reset_data":    resetRequest.ResetData,
		"reset_config":  resetRequest.ResetConfig,
		"reset_metrics": resetRequest.ResetMetrics,
		"timestamp":     time.Now(),
	})
}

//

// validateRequest 验证请求
func (h *AdvancedAIHandler) validateRequest(request *advanced.AIRequest) error {
	if request.Type == "" {
		return fmt.Errorf("request type is required")
	}

	if request.Input == nil {
		return fmt.Errorf("request input is required")
	}

	if request.Capability == "" {
		request.Capability = advanced.CapabilityHybrid //
	}

	return nil
}

// formatResponse 格式化响应
func (h *AdvancedAIHandler) formatResponse(response *advanced.AIResponse) gin.H {
	return gin.H{
		"request_id":        response.RequestID,
		"success":           response.Success,
		"result":            response.Result,
		"confidence":        response.Confidence,
		"process_time":      response.ProcessTime.String(),
		"used_capabilities": response.UsedCapabilities,
		"metadata":          response.Metadata,
		"error":             response.Error,
		"created_at":        response.CreatedAt,
	}
}
