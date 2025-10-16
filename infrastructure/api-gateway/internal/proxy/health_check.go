package proxy

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/logger"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/registry"
)

// HealthChecker 健康检查器
type HealthChecker struct {
	registry registry.Registry
	client   *http.Client
	logger   logger.Logger
	
	// 检查状态
	checking map[string]bool
	mu       sync.RWMutex
	
	// 停止信号
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	InstanceID string
	Healthy    bool
	Error      error
	Latency    time.Duration
	Timestamp  time.Time
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(reg registry.Registry, client *http.Client, log logger.Logger) *HealthChecker {
	return &HealthChecker{
		registry: reg,
		client:   client,
		logger:   log,
		checking: make(map[string]bool),
		stopCh:   make(chan struct{}),
	}
}

// CheckAll 检查所有服务实例
func (h *HealthChecker) CheckAll() {
	ctx := context.Background()
	
	// 获取所有服务
	services, err := h.registry.ListServices(ctx)
	if err != nil {
		h.logger.Errorf("Failed to list services for health check: %v", err)
		return
	}
	
	// 并发检查所有实例
	var wg sync.WaitGroup
	for serviceName, instances := range services {
		for _, instance := range instances {
			wg.Add(1)
			go func(svcName string, inst *registry.ServiceInstance) {
				defer wg.Done()
				h.checkInstance(ctx, svcName, inst)
			}(serviceName, instance)
		}
	}
	
	wg.Wait()
}

// CheckService 检查指定服务的所有实例
func (h *HealthChecker) CheckService(serviceName string) error {
	ctx := context.Background()
	
	instances, err := h.registry.Discover(ctx, serviceName)
	if err != nil {
		return fmt.Errorf("failed to discover service instances: %w", err)
	}
	
	var wg sync.WaitGroup
	for _, instance := range instances {
		wg.Add(1)
		go func(inst *registry.ServiceInstance) {
			defer wg.Done()
			h.checkInstance(ctx, serviceName, inst)
		}(instance)
	}
	
	wg.Wait()
	return nil
}

// CheckInstance 检查单个实例
func (h *HealthChecker) CheckInstance(instanceID string) (*HealthCheckResult, error) {
	ctx := context.Background()
	
	// 查找实例
	services, err := h.registry.ListServices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	
	for serviceName, instances := range services {
		for _, instance := range instances {
			if instance.ID == instanceID {
				return h.performHealthCheck(ctx, serviceName, instance), nil
			}
		}
	}
	
	return nil, fmt.Errorf("instance not found: %s", instanceID)
}

// checkInstance 检查实例（内部方法）
func (h *HealthChecker) checkInstance(ctx context.Context, serviceName string, instance *registry.ServiceInstance) {
	// 防止重复检查
	h.mu.Lock()
	if h.checking[instance.ID] {
		h.mu.Unlock()
		return
	}
	h.checking[instance.ID] = true
	h.mu.Unlock()
	
	defer func() {
		h.mu.Lock()
		delete(h.checking, instance.ID)
		h.mu.Unlock()
	}()
	
	result := h.performHealthCheck(ctx, serviceName, instance)
	
	// 更新健康状态
	newStatus := registry.HealthStatusUnhealthy
	if result.Healthy {
		newStatus = registry.HealthStatusHealthy
	}
	
	if err := h.registry.UpdateHealth(ctx, instance.ID, newStatus); err != nil {
		h.logger.Errorf("Failed to update health status for %s: %v", instance.ID, err)
	}
	
	// 记录日志
	if result.Error != nil {
		h.logger.WithFields(map[string]interface{}{
			"service":   serviceName,
			"instance":  instance.ID,
			"error":     result.Error.Error(),
			"latency":   result.Latency,
		}).Warn("Health check failed")
	} else {
		h.logger.WithFields(map[string]interface{}{
			"service":  serviceName,
			"instance": instance.ID,
			"healthy":  result.Healthy,
			"latency":  result.Latency,
		}).Debug("Health check completed")
	}
}

// performHealthCheck 执行健康检查
func (h *HealthChecker) performHealthCheck(ctx context.Context, serviceName string, instance *registry.ServiceInstance) *HealthCheckResult {
	start := time.Now()
	result := &HealthCheckResult{
		InstanceID: instance.ID,
		Timestamp:  start,
	}
	
	// 构建健康检查URL
	healthCheckPath := instance.Meta["health_check_path"]
	if healthCheckPath == "" {
		healthCheckPath = "/health"
	}
	
	baseURL := instance.Meta["url"]
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://%s:%d", instance.Address, instance.Port)
	}
	
	healthCheckURL := baseURL + healthCheckPath
	
	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", healthCheckURL, nil)
	if err != nil {
		result.Error = fmt.Errorf("failed to create health check request: %w", err)
		result.Latency = time.Since(start)
		return result
	}
	
	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(ctx)
	
	// 执行请求
	resp, err := h.client.Do(req)
	result.Latency = time.Since(start)
	
	if err != nil {
		result.Error = fmt.Errorf("health check request failed: %w", err)
		return result
	}
	defer resp.Body.Close()
	
	// 检查响应状态
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Healthy = true
	} else {
		result.Error = fmt.Errorf("health check returned status %d", resp.StatusCode)
	}
	
	return result
}

// StartPeriodicCheck 启动定期健康检查
func (h *HealthChecker) StartPeriodicCheck(interval time.Duration) {
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				h.CheckAll()
			case <-h.stopCh:
				return
			}
		}
	}()
	
	h.logger.Infof("Started periodic health check with interval: %v", interval)
}

// Close 关闭健康检查器
func (h *HealthChecker) Close() {
	close(h.stopCh)
	h.wg.Wait()
	h.logger.Info("Health checker closed")
}

// GetHealthStatus 获取实例健康状态
func (h *HealthChecker) GetHealthStatus(instanceID string) (registry.HealthStatus, error) {
	ctx := context.Background()
	
	services, err := h.registry.ListServices(ctx)
	if err != nil {
		return registry.HealthStatusUnknown, fmt.Errorf("failed to list services: %w", err)
	}
	
	for _, instances := range services {
		for _, instance := range instances {
			if instance.ID == instanceID {
				return instance.Health, nil
			}
		}
	}
	
	return registry.HealthStatusUnknown, fmt.Errorf("instance not found: %s", instanceID)
}

// GetServiceHealth 获取服务健康统计
func (h *HealthChecker) GetServiceHealth(serviceName string) (map[string]int, error) {
	ctx := context.Background()
	
	instances, err := h.registry.Discover(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to discover service instances: %w", err)
	}
	
	stats := map[string]int{
		"total":     len(instances),
		"healthy":   0,
		"unhealthy": 0,
		"unknown":   0,
	}
	
	for _, instance := range instances {
		switch instance.Health {
		case registry.HealthStatusHealthy:
			stats["healthy"]++
		case registry.HealthStatusUnhealthy:
			stats["unhealthy"]++
		default:
			stats["unknown"]++
		}
	}
	
	return stats, nil
}

// IsServiceHealthy 检查服务是否健康
func (h *HealthChecker) IsServiceHealthy(serviceName string) (bool, error) {
	stats, err := h.GetServiceHealth(serviceName)
	if err != nil {
		return false, err
	}
	
	// 如果有健康的实例，则认为服务是健康的
	return stats["healthy"] > 0, nil
}

// GetUnhealthyInstances 获取不健康的实例
func (h *HealthChecker) GetUnhealthyInstances() ([]*registry.ServiceInstance, error) {
	ctx := context.Background()
	
	services, err := h.registry.ListServices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	
	var unhealthyInstances []*registry.ServiceInstance
	for _, instances := range services {
		for _, instance := range instances {
			if instance.Health == registry.HealthStatusUnhealthy {
				unhealthyInstances = append(unhealthyInstances, instance)
			}
		}
	}
	
	return unhealthyInstances, nil
}