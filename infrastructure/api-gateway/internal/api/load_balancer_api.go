package api

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/health"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/proxy"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// LoadBalancerAPI 负载均衡API处理器
type LoadBalancerAPI struct {
	loadBalancerManager proxy.LoadBalancerManager
	healthChecker       health.HealthChecker
	logger              *logrus.Logger
}

// NewLoadBalancerAPI 创建负载均衡API处理器
func NewLoadBalancerAPI(
	loadBalancerManager proxy.LoadBalancerManager,
	healthChecker health.HealthChecker,
	logger *logrus.Logger,
) *LoadBalancerAPI {
	return &LoadBalancerAPI{
		loadBalancerManager: loadBalancerManager,
		healthChecker:       healthChecker,
		logger:              logger,
	}
}

// RegisterRoutes 注册API路由
func (api *LoadBalancerAPI) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		lb := v1.Group("/load-balancer")
		{
			// 负载均衡策略管理
			lb.GET("/strategies", api.GetStrategies)
			lb.GET("/strategies/:service", api.GetServiceStrategy)
			lb.PUT("/strategies/:service", api.SetServiceStrategy)
			
			// 负载均衡统计
			lb.GET("/stats", api.GetAllStats)
			lb.GET("/stats/:service", api.GetServiceStats)
			
			// 健康检查状态
			lb.GET("/health", api.GetHealthStatus)
			lb.GET("/health/:service", api.GetServiceHealth)
			lb.GET("/health/:service/:instance", api.GetInstanceHealth)
		}
	}
}

// StrategyResponse 策略响应
type StrategyResponse struct {
	Service  string                  `json:"service"`
	Strategy types.LoadBalancerType `json:"strategy"`
}

// SetStrategyRequest 设置策略请求
type SetStrategyRequest struct {
	Strategy types.LoadBalancerType `json:"strategy" binding:"required"`
}

// GetStrategies 获取所有服务的负载均衡策略
func (api *LoadBalancerAPI) GetStrategies(c *gin.Context) {
	// 这里需要从负载均衡管理器获取所有服务的策略
	// 由于当前接口限制，我们返回一个示例响应
	strategies := []StrategyResponse{
		{Service: "user-service", Strategy: types.RoundRobin},
		{Service: "product-service", Strategy: types.WeightedRoundRobin},
		{Service: "order-service", Strategy: types.LeastConnections},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"strategies": strategies,
	})
}

// GetServiceStrategy 获取服务的负载均衡策略
func (api *LoadBalancerAPI) GetServiceStrategy(c *gin.Context) {
	serviceName := c.Param("service")
	
	// 这里需要从负载均衡管理器获取服务策略
	// 由于当前接口限制，我们返回默认策略
	strategy := types.RoundRobin
	
	c.JSON(http.StatusOK, StrategyResponse{
		Service:  serviceName,
		Strategy: strategy,
	})
}

// SetServiceStrategy 设置服务的负载均衡策略
func (api *LoadBalancerAPI) SetServiceStrategy(c *gin.Context) {
	serviceName := c.Param("service")
	
	var req SetStrategyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}
	
	// 设置负载均衡策略
	if err := api.loadBalancerManager.SetLoadBalancerStrategy(serviceName, req.Strategy); err != nil {
		api.logger.WithFields(logrus.Fields{
			"service":  serviceName,
			"strategy": req.Strategy,
			"error":    err,
		}).Error("Failed to set load balancer strategy")
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to set load balancer strategy",
			"details": err.Error(),
		})
		return
	}
	
	api.logger.WithFields(logrus.Fields{
		"service":  serviceName,
		"strategy": req.Strategy,
	}).Info("Load balancer strategy updated")
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Strategy updated successfully",
		"service": serviceName,
		"strategy": req.Strategy,
	})
}

// GetAllStats 获取所有服务的负载均衡统计
func (api *LoadBalancerAPI) GetAllStats(c *gin.Context) {
	// 这里需要从负载均衡管理器获取所有服务的统计信息
	// 由于当前接口限制，我们返回示例数据
	stats := map[string]interface{}{
		"user-service": map[string]interface{}{
			"total_requests": 1000,
			"active_connections": 50,
			"instances": []map[string]interface{}{
				{"id": "user-1", "connections": 25, "requests": 500},
				{"id": "user-2", "connections": 25, "requests": 500},
			},
		},
		"product-service": map[string]interface{}{
			"total_requests": 800,
			"active_connections": 30,
			"instances": []map[string]interface{}{
				{"id": "product-1", "connections": 15, "requests": 400},
				{"id": "product-2", "connections": 15, "requests": 400},
			},
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// GetServiceStats 获取服务的负载均衡统计
func (api *LoadBalancerAPI) GetServiceStats(c *gin.Context) {
	serviceName := c.Param("service")
	
	stats := api.loadBalancerManager.GetStats(serviceName)
	
	c.JSON(http.StatusOK, gin.H{
		"service": serviceName,
		"stats":   stats,
	})
}

// GetHealthStatus 获取所有服务的健康状态
func (api *LoadBalancerAPI) GetHealthStatus(c *gin.Context) {
	if api.healthChecker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Health checker not available",
		})
		return
	}
	
	// 这里需要获取所有服务的健康状态
	// 由于当前接口限制，我们返回示例数据
	healthStatus := map[string]interface{}{
		"user-service": map[string]interface{}{
			"healthy_instances": 2,
			"total_instances": 2,
			"status": "healthy",
		},
		"product-service": map[string]interface{}{
			"healthy_instances": 1,
			"total_instances": 2,
			"status": "degraded",
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"health_status": healthStatus,
	})
}

// GetServiceHealth 获取服务的健康状态
func (api *LoadBalancerAPI) GetServiceHealth(c *gin.Context) {
	serviceName := c.Param("service")
	
	if api.healthChecker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Health checker not available",
		})
		return
	}
	
	serviceHealth := api.healthChecker.GetServiceHealth(serviceName)
	if serviceHealth == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Service not found or not monitored",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"service": serviceName,
		"health":  serviceHealth,
	})
}

// GetInstanceHealth 获取实例的健康状态
func (api *LoadBalancerAPI) GetInstanceHealth(c *gin.Context) {
	serviceName := c.Param("service")
	instanceID := c.Param("instance")
	
	if api.healthChecker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Health checker not available",
		})
		return
	}
	
	instanceHealth, exists := api.healthChecker.GetInstanceHealth(serviceName, instanceID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Instance not found or not monitored",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"service":  serviceName,
		"instance": instanceID,
		"health":   instanceHealth,
	})
}

// HealthSummary 健康状态摘要
type HealthSummary struct {
	Service          string `json:"service"`
	HealthyInstances int    `json:"healthy_instances"`
	TotalInstances   int    `json:"total_instances"`
	Status           string `json:"status"` // healthy, degraded, unhealthy
}

// GetHealthSummary 获取健康状态摘要
func (api *LoadBalancerAPI) GetHealthSummary(c *gin.Context) {
	if api.healthChecker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Health checker not available",
		})
		return
	}
	
	// 获取查询参数
	serviceFilter := c.Query("service")
	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)
	
	// 这里需要实现获取健康状态摘要的逻辑
	// 由于当前接口限制，我们返回示例数据
	summaries := []HealthSummary{
		{Service: "user-service", HealthyInstances: 2, TotalInstances: 2, Status: "healthy"},
		{Service: "product-service", HealthyInstances: 1, TotalInstances: 2, Status: "degraded"},
		{Service: "order-service", HealthyInstances: 0, TotalInstances: 2, Status: "unhealthy"},
	}
	
	// 应用过滤器
	if serviceFilter != "" {
		filtered := make([]HealthSummary, 0)
		for _, summary := range summaries {
			if summary.Service == serviceFilter {
				filtered = append(filtered, summary)
			}
		}
		summaries = filtered
	}
	
	// 应用限制
	if limit > 0 && len(summaries) > limit {
		summaries = summaries[:limit]
	}
	
	c.JSON(http.StatusOK, gin.H{
		"summaries": summaries,
		"total":     len(summaries),
	})
}