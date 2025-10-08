package health

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/config"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/registry"
	"github.com/sirupsen/logrus"
)

// HealthStatus 健康状态
type HealthStatus int

const (
	Healthy HealthStatus = iota
	Unhealthy
	Unknown
)

// InstanceHealth 实例健康状态
type InstanceHealth struct {
	Instance      registry.ServiceInstance `json:"instance"`
	Status        HealthStatus             `json:"status"`
	LastCheck     time.Time                `json:"last_check"`
	FailureCount  int                      `json:"failure_count"`
	SuccessCount  int                      `json:"success_count"`
	ResponseTime  time.Duration            `json:"response_time"`
	ErrorMessage  string                   `json:"error_message,omitempty"`
}

// HealthChecker 健康检查器接口
type HealthChecker interface {
	// StartHealthChecks 启动健康检查
	StartHealthChecks(ctx context.Context)
	
	// StopHealthChecks 停止健康检查
	StopHealthChecks()
	
	// GetHealthyInstances 获取健康的实例
	GetHealthyInstances(serviceName string) []registry.ServiceInstance
	
	// GetInstanceHealth 获取实例健康状态
	GetInstanceHealth(serviceName, instanceID string) (*InstanceHealth, bool)
	
	// GetServiceHealth 获取服务健康状态
	GetServiceHealth(serviceName string) map[string]*InstanceHealth
	
	// RegisterService 注册服务进行健康检查
	RegisterService(serviceName string, instances []registry.ServiceInstance)
	
	// UnregisterService 取消注册服务
	UnregisterService(serviceName string)
	
	// UpdateInstances 更新服务实例
	UpdateInstances(serviceName string, instances []registry.ServiceInstance)
}

// healthChecker 健康检查器实现
type healthChecker struct {
	config     config.HealthCheckConfig
	registry   registry.Registry
	logger     *logrus.Logger
	httpClient *http.Client
	
	// 健康状态存储
	healthStatus map[string]map[string]*InstanceHealth
	mutex        sync.RWMutex
	
	// 控制
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(config config.HealthCheckConfig, registry registry.Registry, logger *logrus.Logger) HealthChecker {
	return &healthChecker{
		config:       config,
		registry:     registry,
		logger:       logger,
		healthStatus: make(map[string]map[string]*InstanceHealth),
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// StartHealthChecks 启动健康检查
func (hc *healthChecker) StartHealthChecks(ctx context.Context) {
	if !hc.config.Enabled {
		hc.logger.Info("Health checks are disabled")
		return
	}
	
	hc.ctx, hc.cancel = context.WithCancel(ctx)
	
	hc.logger.Info("Starting health checks")
	
	// 启动健康检查循环
	hc.wg.Add(1)
	go hc.healthCheckLoop()
	
	// 启动清理循环
	hc.wg.Add(1)
	go hc.cleanupLoop()
}

// StopHealthChecks 停止健康检查
func (hc *healthChecker) StopHealthChecks() {
	if hc.cancel != nil {
		hc.cancel()
		hc.wg.Wait()
		hc.logger.Info("Health checks stopped")
	}
}

// GetHealthyInstances 获取健康的实例
func (hc *healthChecker) GetHealthyInstances(serviceName string) []registry.ServiceInstance {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()
	
	serviceHealth, exists := hc.healthStatus[serviceName]
	if !exists {
		return nil
	}
	
	var healthyInstances []registry.ServiceInstance
	for _, health := range serviceHealth {
		if health.Status == Healthy {
			healthyInstances = append(healthyInstances, health.Instance)
		}
	}
	
	return healthyInstances
}

// GetInstanceHealth 获取实例健康状态
func (hc *healthChecker) GetInstanceHealth(serviceName, instanceID string) (*InstanceHealth, bool) {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()
	
	serviceHealth, exists := hc.healthStatus[serviceName]
	if !exists {
		return nil, false
	}
	
	health, exists := serviceHealth[instanceID]
	return health, exists
}

// GetServiceHealth 获取服务健康状态
func (hc *healthChecker) GetServiceHealth(serviceName string) map[string]*InstanceHealth {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()
	
	serviceHealth, exists := hc.healthStatus[serviceName]
	if !exists {
		return nil
	}
	
	// 返回副本以避免并发修改
	result := make(map[string]*InstanceHealth)
	for k, v := range serviceHealth {
		result[k] = v
	}
	
	return result
}

// RegisterService 注册服务进行健康检查
func (hc *healthChecker) RegisterService(serviceName string, instances []registry.ServiceInstance) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	
	if hc.healthStatus[serviceName] == nil {
		hc.healthStatus[serviceName] = make(map[string]*InstanceHealth)
	}
	
	for _, instance := range instances {
		hc.healthStatus[serviceName][instance.ID] = &InstanceHealth{
			Instance:     instance,
			Status:       Unknown,
			LastCheck:    time.Time{},
			FailureCount: 0,
			SuccessCount: 0,
		}
	}
	
	hc.logger.WithFields(logrus.Fields{
		"service":   serviceName,
		"instances": len(instances),
	}).Info("Registered service for health checks")
}

// UnregisterService 取消注册服务
func (hc *healthChecker) UnregisterService(serviceName string) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	
	delete(hc.healthStatus, serviceName)
	
	hc.logger.WithField("service", serviceName).Info("Unregistered service from health checks")
}

// UpdateInstances 更新服务实例
func (hc *healthChecker) UpdateInstances(serviceName string, instances []registry.ServiceInstance) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	
	if hc.healthStatus[serviceName] == nil {
		hc.healthStatus[serviceName] = make(map[string]*InstanceHealth)
	}
	
	// 创建新的实例映射
	newInstances := make(map[string]registry.ServiceInstance)
	for _, instance := range instances {
		newInstances[instance.ID] = instance
	}
	
	// 更新现有实例，删除不存在的实例
	currentHealth := hc.healthStatus[serviceName]
	newHealth := make(map[string]*InstanceHealth)
	
	for instanceID, instance := range newInstances {
		if existingHealth, exists := currentHealth[instanceID]; exists {
			// 更新现有实例
			existingHealth.Instance = instance
			newHealth[instanceID] = existingHealth
		} else {
			// 添加新实例
			newHealth[instanceID] = &InstanceHealth{
				Instance:     instance,
				Status:       Unknown,
				LastCheck:    time.Time{},
				FailureCount: 0,
				SuccessCount: 0,
			}
		}
	}
	
	hc.healthStatus[serviceName] = newHealth
	
	hc.logger.WithFields(logrus.Fields{
		"service":   serviceName,
		"instances": len(instances),
	}).Info("Updated service instances for health checks")
}

// healthCheckLoop 健康检查循环
func (hc *healthChecker) healthCheckLoop() {
	defer hc.wg.Done()
	
	ticker := time.NewTicker(hc.config.Interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-hc.ctx.Done():
			return
		case <-ticker.C:
			hc.performHealthChecks()
		}
	}
}

// performHealthChecks 执行健康检查
func (hc *healthChecker) performHealthChecks() {
	hc.mutex.RLock()
	services := make(map[string]map[string]*InstanceHealth)
	for serviceName, serviceHealth := range hc.healthStatus {
		services[serviceName] = make(map[string]*InstanceHealth)
		for instanceID, health := range serviceHealth {
			services[serviceName][instanceID] = health
		}
	}
	hc.mutex.RUnlock()
	
	for serviceName, serviceHealth := range services {
		for instanceID, health := range serviceHealth {
			hc.wg.Add(1)
			go hc.checkInstance(serviceName, instanceID, health)
		}
	}
}

// checkInstance 检查单个实例
func (hc *healthChecker) checkInstance(serviceName, instanceID string, health *InstanceHealth) {
	defer hc.wg.Done()
	
	start := time.Now()
	
	// 构建健康检查URL
	healthCheckPath := health.Instance.Meta["health_check_path"]
	if healthCheckPath == "" {
		healthCheckPath = "/health"
	}
	
	baseURL := health.Instance.Meta["url"]
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://%s:%d", health.Instance.Address, health.Instance.Port)
	}
	
	healthURL := baseURL + healthCheckPath
	
	hc.logger.WithFields(logrus.Fields{
		"service":     serviceName,
		"instance":    instanceID,
		"health_url":  healthURL,
		"base_url":    baseURL,
		"health_path": healthCheckPath,
	}).Debug("Performing health check")
	
	// 执行健康检查
	ctx, cancel := context.WithTimeout(hc.ctx, hc.config.Timeout)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		hc.logger.WithFields(logrus.Fields{
			"service":  serviceName,
			"instance": instanceID,
			"error":    err.Error(),
		}).Error("Failed to create health check request")
		hc.updateHealthStatus(serviceName, instanceID, false, time.Since(start), err.Error())
		return
	}
	
	resp, err := hc.httpClient.Do(req)
	if err != nil {
		hc.logger.WithFields(logrus.Fields{
			"service":  serviceName,
			"instance": instanceID,
			"error":    err.Error(),
		}).Error("Health check request failed")
		hc.updateHealthStatus(serviceName, instanceID, false, time.Since(start), err.Error())
		return
	}
	defer resp.Body.Close()
	
	// 检查状态码
	isHealthy := hc.isStatusCodeHealthy(resp.StatusCode)
	errorMsg := ""
	if !isHealthy {
		errorMsg = fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
	}
	
	hc.logger.WithFields(logrus.Fields{
		"service":       serviceName,
		"instance":      instanceID,
		"status_code":   resp.StatusCode,
		"is_healthy":    isHealthy,
		"response_time": time.Since(start),
	}).Info("Health check completed")
	
	hc.updateHealthStatus(serviceName, instanceID, isHealthy, time.Since(start), errorMsg)
}

// isStatusCodeHealthy 检查状态码是否健康
func (hc *healthChecker) isStatusCodeHealthy(statusCode int) bool {
	// 默认认为200状态码是健康的
	return statusCode == 200
}

// updateHealthStatus 更新健康状态
func (hc *healthChecker) updateHealthStatus(serviceName, instanceID string, isHealthy bool, responseTime time.Duration, errorMsg string) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	
	serviceHealth, exists := hc.healthStatus[serviceName]
	if !exists {
		return
	}
	
	health, exists := serviceHealth[instanceID]
	if !exists {
		return
	}
	
	health.LastCheck = time.Now()
	health.ResponseTime = responseTime
	health.ErrorMessage = errorMsg
	
	if isHealthy {
		health.SuccessCount++
		health.FailureCount = 0
		
		// 检查是否应该标记为健康
		if health.Status != Healthy && health.SuccessCount >= 3 {
			health.Status = Healthy
			hc.logger.WithFields(logrus.Fields{
				"service":  serviceName,
				"instance": instanceID,
			}).Info("Instance marked as healthy")
		}
	} else {
		health.FailureCount++
		health.SuccessCount = 0
		
		// 检查是否应该标记为不健康
		if health.Status != Unhealthy && health.FailureCount >= 3 {
			health.Status = Unhealthy
			hc.logger.WithFields(logrus.Fields{
				"service":  serviceName,
				"instance": instanceID,
				"error":    errorMsg,
			}).Warn("Instance marked as unhealthy")
		}
	}
}

// cleanupLoop 清理循环
func (hc *healthChecker) cleanupLoop() {
	defer hc.wg.Done()
	
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-hc.ctx.Done():
			return
		case <-ticker.C:
			hc.cleanup()
		}
	}
}

// cleanup 清理过期数据
func (hc *healthChecker) cleanup() {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	
	cutoff := time.Now().Add(-24 * time.Hour)
	
	for serviceName, serviceHealth := range hc.healthStatus {
		for instanceID, health := range serviceHealth {
			if health.LastCheck.Before(cutoff) {
				delete(serviceHealth, instanceID)
				hc.logger.WithFields(logrus.Fields{
					"service":  serviceName,
					"instance": instanceID,
				}).Debug("Cleaned up stale health data")
			}
		}
		
		// 如果服务没有实例，删除服务
		if len(serviceHealth) == 0 {
			delete(hc.healthStatus, serviceName)
		}
	}
}