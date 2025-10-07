package registry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/logger"
	"github.com/hashicorp/consul/api"
)

// consulRegistry Consul服务注册实现
type consulRegistry struct {
	client   *api.Client
	logger   logger.Logger
	watchers map[string][]chan []*ServiceInstance
	mu       sync.RWMutex
	stopCh   chan struct{}
}

// newConsulRegistry 创建Consul注册中心
func newConsulRegistry(endpoints []string, options map[string]string, log logger.Logger) (Registry, error) {
	config := api.DefaultConfig()
	
	// 设置Consul地址
	if len(endpoints) > 0 {
		config.Address = endpoints[0]
	}
	
	// 设置其他选项
	if token, ok := options["token"]; ok {
		config.Token = token
	}
	
	if datacenter, ok := options["datacenter"]; ok {
		config.Datacenter = datacenter
	}
	
	// 设置超时
	if timeout, ok := options["timeout"]; ok {
		if t, err := time.ParseDuration(timeout); err == nil {
			config.HttpClient.Timeout = t
		}
	}
	
	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}
	
	registry := &consulRegistry{
		client:   client,
		logger:   log,
		watchers: make(map[string][]chan []*ServiceInstance),
		stopCh:   make(chan struct{}),
	}
	
	// 启动健康检查监控
	go registry.startHealthMonitor()
	
	return registry, nil
}

// Register 注册服务到Consul
func (r *consulRegistry) Register(ctx context.Context, instance *ServiceInstance) error {
	if instance.ID == "" {
		instance.ID = fmt.Sprintf("%s-%s-%d", instance.Name, instance.Address, instance.Port)
	}
	
	// 构建Consul服务注册信息
	registration := &api.AgentServiceRegistration{
		ID:      instance.ID,
		Name:    instance.Name,
		Address: instance.Address,
		Port:    instance.Port,
		Tags:    instance.Tags,
		Meta:    instance.Meta,
		Weights: &api.AgentWeights{
			Passing: instance.Weight,
			Warning: 1,
		},
	}
	
	// 添加健康检查
	if healthCheckURL, ok := instance.Meta["health_check"]; ok && healthCheckURL != "" {
		registration.Check = &api.AgentServiceCheck{
			HTTP:                           healthCheckURL,
			Interval:                       "30s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "90s",
		}
	} else {
		// 默认TCP健康检查
		registration.Check = &api.AgentServiceCheck{
			TCP:                            fmt.Sprintf("%s:%d", instance.Address, instance.Port),
			Interval:                       "30s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "90s",
		}
	}
	
	// 注册服务
	if err := r.client.Agent().ServiceRegister(registration); err != nil {
		return fmt.Errorf("failed to register service to consul: %w", err)
	}
	
	r.logger.Infof("Service registered to Consul: %s (%s)", instance.Name, instance.ID)
	
	return nil
}

// Deregister 从Consul注销服务
func (r *consulRegistry) Deregister(ctx context.Context, instanceID string) error {
	if err := r.client.Agent().ServiceDeregister(instanceID); err != nil {
		return fmt.Errorf("failed to deregister service from consul: %w", err)
	}
	
	r.logger.Infof("Service deregistered from Consul: %s", instanceID)
	
	return nil
}

// Discover 从Consul发现服务
func (r *consulRegistry) Discover(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	// 查询健康的服务实例
	services, _, err := r.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to discover service from consul: %w", err)
	}
	
	var instances []*ServiceInstance
	for _, service := range services {
		instance := &ServiceInstance{
			ID:      service.Service.ID,
			Name:    service.Service.Service,
			Address: service.Service.Address,
			Port:    service.Service.Port,
			Tags:    service.Service.Tags,
			Meta:    service.Service.Meta,
			Health:  r.convertConsulHealthStatus(service.Checks),
			Weight:  service.Service.Weights.Passing,
		}
		
		instances = append(instances, instance)
	}
	
	return instances, nil
}

// ListServices 获取所有服务
func (r *consulRegistry) ListServices(ctx context.Context) (map[string][]*ServiceInstance, error) {
	// 获取所有服务列表
	services, _, err := r.client.Catalog().Services(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list services from consul: %w", err)
	}
	
	result := make(map[string][]*ServiceInstance)
	
	// 为每个服务获取实例信息
	for serviceName := range services {
		instances, err := r.Discover(ctx, serviceName)
		if err != nil {
			r.logger.Warnf("Failed to discover service %s: %v", serviceName, err)
			continue
		}
		result[serviceName] = instances
	}
	
	return result, nil
}

// UpdateHealth 更新服务健康状态
func (r *consulRegistry) UpdateHealth(ctx context.Context, instanceID string, status HealthStatus) error {
	var consulStatus string
	switch status {
	case HealthStatusHealthy:
		consulStatus = "pass"
	case HealthStatusUnhealthy:
		consulStatus = "fail"
	default:
		consulStatus = "warn"
	}
	
	// 更新健康检查状态
	checkID := fmt.Sprintf("service:%s", instanceID)
	if err := r.client.Agent().UpdateTTL(checkID, "", consulStatus); err != nil {
		return fmt.Errorf("failed to update health status in consul: %w", err)
	}
	
	r.logger.Debugf("Service health updated in Consul: %s -> %s", instanceID, status)
	
	return nil
}

// Watch 监听服务变化
func (r *consulRegistry) Watch(ctx context.Context, serviceName string) (<-chan []*ServiceInstance, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	ch := make(chan []*ServiceInstance, 10)
	
	if r.watchers[serviceName] == nil {
		r.watchers[serviceName] = make([]chan []*ServiceInstance, 0)
	}
	r.watchers[serviceName] = append(r.watchers[serviceName], ch)
	
	// 启动监听协程
	go r.watchService(ctx, serviceName, ch)
	
	return ch, nil
}

// watchService 监听单个服务的变化
func (r *consulRegistry) watchService(ctx context.Context, serviceName string, ch chan []*ServiceInstance) {
	defer func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		
		// 移除监听器
		watchers := r.watchers[serviceName]
		for i, watcher := range watchers {
			if watcher == ch {
				r.watchers[serviceName] = append(watchers[:i], watchers[i+1:]...)
				break
			}
		}
		
		close(ch)
	}()
	
	var lastIndex uint64
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-r.stopCh:
			return
		default:
		}
		
		// 使用阻塞查询监听服务变化
		queryOptions := &api.QueryOptions{
			WaitIndex: lastIndex,
			WaitTime:  30 * time.Second,
		}
		
		services, meta, err := r.client.Health().Service(serviceName, "", true, queryOptions)
		if err != nil {
			r.logger.Errorf("Failed to watch service %s: %v", serviceName, err)
			time.Sleep(5 * time.Second)
			continue
		}
		
		// 如果索引没有变化，继续等待
		if meta.LastIndex == lastIndex {
			continue
		}
		
		lastIndex = meta.LastIndex
		
		// 转换为内部格式
		var instances []*ServiceInstance
		for _, service := range services {
			instance := &ServiceInstance{
				ID:      service.Service.ID,
				Name:    service.Service.Service,
				Address: service.Service.Address,
				Port:    service.Service.Port,
				Tags:    service.Service.Tags,
				Meta:    service.Service.Meta,
				Health:  r.convertConsulHealthStatus(service.Checks),
				Weight:  service.Service.Weights.Passing,
			}
			instances = append(instances, instance)
		}
		
		// 发送更新
		select {
		case ch <- instances:
		default:
			// 如果通道满了，跳过这次更新
		}
	}
}

// startHealthMonitor 启动健康检查监控
func (r *consulRegistry) startHealthMonitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			r.checkServicesHealth()
		case <-r.stopCh:
			return
		}
	}
}

// checkServicesHealth 检查所有服务的健康状态
func (r *consulRegistry) checkServicesHealth() {
	services, err := r.client.Agent().Services()
	if err != nil {
		r.logger.Errorf("Failed to get services for health check: %v", err)
		return
	}
	
	for _, service := range services {
		// 获取服务的健康检查状态
		checks, _, err := r.client.Health().Checks(service.Service, nil)
		if err != nil {
			r.logger.Errorf("Failed to get health checks for service %s: %v", service.Service, err)
			continue
		}
		
		// 通知监听者健康状态变化
		r.notifyHealthChange(service.Service, checks)
	}
}

// notifyHealthChange 通知健康状态变化
func (r *consulRegistry) notifyHealthChange(serviceName string, checks api.HealthChecks) {
	r.mu.RLock()
	watchers := r.watchers[serviceName]
	r.mu.RUnlock()
	
	if len(watchers) == 0 {
		return
	}
	
	// 获取最新的服务实例
	instances, err := r.Discover(context.Background(), serviceName)
	if err != nil {
		r.logger.Errorf("Failed to discover service %s for health notification: %v", serviceName, err)
		return
	}
	
	// 通知所有监听者
	for _, watcher := range watchers {
		select {
		case watcher <- instances:
		default:
			// 如果通道满了，跳过这次通知
		}
	}
}

// convertConsulHealthStatus 转换Consul健康状态
func (r *consulRegistry) convertConsulHealthStatus(checks api.HealthChecks) HealthStatus {
	if len(checks) == 0 {
		return HealthStatusUnknown
	}
	
	for _, check := range checks {
		switch check.Status {
		case api.HealthCritical:
			return HealthStatusUnhealthy
		case api.HealthWarning:
			return HealthStatusUnhealthy
		}
	}
	
	return HealthStatusHealthy
}

// Close 关闭Consul注册中心
func (r *consulRegistry) Close() error {
	close(r.stopCh)
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// 关闭所有监听器
	for _, watchers := range r.watchers {
		for _, watcher := range watchers {
			close(watcher)
		}
	}
	
	r.watchers = make(map[string][]chan []*ServiceInstance)
	
	r.logger.Info("Consul registry closed")
	
	return nil
}