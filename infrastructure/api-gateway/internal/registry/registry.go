package registry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/config"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/logger"
)

// ServiceInstance 服务实例
type ServiceInstance struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Tags     []string          `json:"tags"`
	Meta     map[string]string `json:"meta"`
	Health   HealthStatus      `json:"health"`
	Weight   int               `json:"weight"`
	LastSeen time.Time         `json:"last_seen"`
}

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// Registry 服务注册接口
type Registry interface {
	// 注册服务
	Register(ctx context.Context, instance *ServiceInstance) error
	
	// 注销服务
	Deregister(ctx context.Context, instanceID string) error
	
	// 发现服务
	Discover(ctx context.Context, serviceName string) ([]*ServiceInstance, error)
	
	// 获取所有服务
	ListServices(ctx context.Context) (map[string][]*ServiceInstance, error)
	
	// 更新服务健康状态
	UpdateHealth(ctx context.Context, instanceID string, status HealthStatus) error
	
	// 监听服务变化
	Watch(ctx context.Context, serviceName string) (<-chan []*ServiceInstance, error)
	
	// 关闭注册中心
	Close() error
}

// staticRegistry 静态服务注册实现
type staticRegistry struct {
	services map[string][]*ServiceInstance
	watchers map[string][]chan []*ServiceInstance
	mu       sync.RWMutex
	logger   logger.Logger
}

// New 创建服务注册实例
func New(cfg config.RegistryConfig, log logger.Logger) Registry {
	switch cfg.Type {
	case "static":
		return newStaticRegistry(log)
	case "consul":
		registry, err := newConsulRegistry(cfg.Endpoints, cfg.Options, log)
		if err != nil {
			log.Errorf("Failed to create Consul registry: %v, falling back to static registry", err)
			return newStaticRegistry(log)
		}
		return registry
	case "etcd":
		registry, err := newEtcdRegistry(cfg.Endpoints, cfg.Options, log)
		if err != nil {
			log.Errorf("Failed to create etcd registry: %v, falling back to static registry", err)
			return newStaticRegistry(log)
		}
		return registry
	default:
		log.Warnf("Unknown registry type: %s, using static registry", cfg.Type)
		return newStaticRegistry(log)
	}
}

// newStaticRegistry 创建静态注册中心
func newStaticRegistry(log logger.Logger) Registry {
	return &staticRegistry{
		services: make(map[string][]*ServiceInstance),
		watchers: make(map[string][]chan []*ServiceInstance),
		logger:   log,
	}
}

// Register 注册服务
func (r *staticRegistry) Register(ctx context.Context, instance *ServiceInstance) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if instance.ID == "" {
		instance.ID = fmt.Sprintf("%s-%s-%d", instance.Name, instance.Address, instance.Port)
	}
	
	if instance.Weight <= 0 {
		instance.Weight = 1
	}
	
	instance.LastSeen = time.Now()
	instance.Health = HealthStatusHealthy
	
	// 检查是否已存在
	instances := r.services[instance.Name]
	found := false
	for i, existing := range instances {
		if existing.ID == instance.ID {
			instances[i] = instance
			found = true
			break
		}
	}
	
	if !found {
		instances = append(instances, instance)
	}
	
	r.services[instance.Name] = instances
	
	r.logger.Infof("Service registered: %s (%s)", instance.Name, instance.ID)
	
	// 通知监听者
	r.notifyWatchers(instance.Name, instances)
	
	return nil
}

// Deregister 注销服务
func (r *staticRegistry) Deregister(ctx context.Context, instanceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for serviceName, instances := range r.services {
		for i, instance := range instances {
			if instance.ID == instanceID {
				// 移除实例
				r.services[serviceName] = append(instances[:i], instances[i+1:]...)
				
				r.logger.Infof("Service deregistered: %s (%s)", serviceName, instanceID)
				
				// 通知监听者
				r.notifyWatchers(serviceName, r.services[serviceName])
				
				return nil
			}
		}
	}
	
	return fmt.Errorf("service instance not found: %s", instanceID)
}

// Discover 发现服务
func (r *staticRegistry) Discover(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	instances, exists := r.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", serviceName)
	}
	
	// 过滤健康的实例
	var healthyInstances []*ServiceInstance
	for _, instance := range instances {
		if instance.Health == HealthStatusHealthy {
			healthyInstances = append(healthyInstances, instance)
		}
	}
	
	return healthyInstances, nil
}

// ListServices 获取所有服务
func (r *staticRegistry) ListServices(ctx context.Context) (map[string][]*ServiceInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// 深拷贝服务列表
	result := make(map[string][]*ServiceInstance)
	for name, instances := range r.services {
		result[name] = make([]*ServiceInstance, len(instances))
		copy(result[name], instances)
	}
	
	return result, nil
}

// UpdateHealth 更新服务健康状态
func (r *staticRegistry) UpdateHealth(ctx context.Context, instanceID string, status HealthStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for serviceName, instances := range r.services {
		for _, instance := range instances {
			if instance.ID == instanceID {
				oldStatus := instance.Health
				instance.Health = status
				instance.LastSeen = time.Now()
				
				r.logger.Debugf("Service health updated: %s (%s) %s -> %s", 
					serviceName, instanceID, oldStatus, status)
				
				// 如果健康状态发生变化，通知监听者
				if oldStatus != status {
					r.notifyWatchers(serviceName, instances)
				}
				
				return nil
			}
		}
	}
	
	return fmt.Errorf("service instance not found: %s", instanceID)
}

// Watch 监听服务变化
func (r *staticRegistry) Watch(ctx context.Context, serviceName string) (<-chan []*ServiceInstance, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	ch := make(chan []*ServiceInstance, 10)
	
	if r.watchers[serviceName] == nil {
		r.watchers[serviceName] = make([]chan []*ServiceInstance, 0)
	}
	r.watchers[serviceName] = append(r.watchers[serviceName], ch)
	
	// 发送当前状态
	if instances, exists := r.services[serviceName]; exists {
		select {
		case ch <- instances:
		default:
		}
	}
	
	// 启动清理协程
	go func() {
		<-ctx.Done()
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
	
	return ch, nil
}

// notifyWatchers 通知监听者
func (r *staticRegistry) notifyWatchers(serviceName string, instances []*ServiceInstance) {
	watchers := r.watchers[serviceName]
	for _, watcher := range watchers {
		select {
		case watcher <- instances:
		default:
			// 如果通道满了，跳过这次通知
		}
	}
}

// Close 关闭注册中心
func (r *staticRegistry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// 关闭所有监听器
	for _, watchers := range r.watchers {
		for _, watcher := range watchers {
			close(watcher)
		}
	}
	
	r.watchers = make(map[string][]chan []*ServiceInstance)
	r.services = make(map[string][]*ServiceInstance)
	
	r.logger.Info("Registry closed")
	
	return nil
}

// LoadFromConfig 从配置加载服务
func (r *staticRegistry) LoadFromConfig(services []config.ServiceConfig) error {
	ctx := context.Background()
	
	for _, svc := range services {
		instance := &ServiceInstance{
			Name:    svc.Name,
			Address: extractAddress(svc.URL),
			Port:    extractPort(svc.URL),
			Weight:  svc.Weight,
			Meta: map[string]string{
				"url":          svc.URL,
				"health_check": svc.HealthCheck,
			},
		}
		
		if err := r.Register(ctx, instance); err != nil {
			return fmt.Errorf("failed to register service %s: %w", svc.Name, err)
		}
	}
	
	return nil
}

// LoadFromStaticConfig 从静态服务配置加载服务实例
func (r *staticRegistry) LoadFromStaticConfig(staticServices map[string][]config.StaticServiceInstance) error {
	ctx := context.Background()
	
	for serviceName, instances := range staticServices {
		for _, staticInstance := range instances {
			instance := &ServiceInstance{
				ID:      staticInstance.ID,
				Name:    serviceName,
				Address: staticInstance.Address,
				Port:    staticInstance.Port,
				Weight:  staticInstance.Weight,
				Tags:    staticInstance.Tags,
				Meta:    staticInstance.Meta,
			}
			
			if err := r.Register(ctx, instance); err != nil {
				return fmt.Errorf("failed to register static service %s instance %s: %w", serviceName, staticInstance.ID, err)
			}
		}
	}
	
	return nil
}

// LoadStaticServices 辅助函数，用于加载静态服务配置
func LoadStaticServices(registry Registry, staticServices map[string][]config.StaticServiceInstance) error {
	// 尝试类型断言为staticRegistry
	if staticReg, ok := registry.(*staticRegistry); ok {
		return staticReg.LoadFromStaticConfig(staticServices)
	}
	
	// 如果不是staticRegistry类型，返回错误
	return fmt.Errorf("registry does not support static service loading")
}

// extractAddress 从URL提取地址
func extractAddress(url string) string {
	// 简单实现，实际应该使用url.Parse
	// TODO: 使用正确的URL解析
	return "localhost"
}

// extractPort 从URL提取端口
func extractPort(url string) int {
	// 简单实现，实际应该使用url.Parse
	// TODO: 使用正确的URL解析
	return 8080
}