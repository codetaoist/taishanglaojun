package proxy

import (
	"fmt"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/health"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/logger"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/registry"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/types"
)

// LoadBalancerManager 负载均衡管理器
type LoadBalancerManager interface {
	// 获取服务的负载均衡器
	GetLoadBalancer(serviceName string) LoadBalancer
	
	// 设置服务的负载均衡策略
	SetLoadBalancerStrategy(serviceName string, strategy types.LoadBalancerType) error
	
	// 选择服务实例
	SelectInstance(serviceName string, instances []*registry.ServiceInstance, clientIP string) (*registry.ServiceInstance, error)
	
	// 更新连接统计（用于最少连接算法）
	IncrementConnections(serviceName, instanceID string)
	DecrementConnections(serviceName, instanceID string)
	
	// 获取负载均衡统计信息
	GetStats(serviceName string) map[string]interface{}
	
	// 关闭管理器
	Close() error
	
	// RegisterHealthChecker 注册健康检查器
	RegisterHealthChecker(healthChecker health.HealthChecker)
}

// loadBalancerManager 负载均衡管理器实现
type loadBalancerManager struct {
	mu                sync.RWMutex
	balancers         map[string]LoadBalancer           // 服务名 -> 负载均衡器
	strategies        map[string]types.LoadBalancerType       // 服务名 -> 负载均衡策略
	connectionStats   map[string]map[string]int64       // 服务名 -> 实例ID -> 连接数
	requestCounts     map[string]map[string]int64       // 服务名 -> 实例ID -> 请求数
	lastRequestTime   map[string]map[string]time.Time   // 服务名 -> 实例ID -> 最后请求时间
	defaultStrategy   types.LoadBalancerType
	healthChecker     health.HealthChecker
	logger            logger.Logger
}

// NewLoadBalancerManager 创建负载均衡管理器
func NewLoadBalancerManager(defaultStrategy types.LoadBalancerType, log logger.Logger) LoadBalancerManager {
	if defaultStrategy == "" {
		defaultStrategy = types.RoundRobin
	}
	
	return &loadBalancerManager{
		balancers:       make(map[string]LoadBalancer),
		strategies:      make(map[string]types.LoadBalancerType),
		connectionStats: make(map[string]map[string]int64),
		requestCounts:   make(map[string]map[string]int64),
		lastRequestTime: make(map[string]map[string]time.Time),
		defaultStrategy: defaultStrategy,
		logger:          log,
	}
}

// GetLoadBalancer 获取服务的负载均衡器
func (lm *loadBalancerManager) GetLoadBalancer(serviceName string) LoadBalancer {
	lm.mu.RLock()
	balancer, exists := lm.balancers[serviceName]
	strategy := lm.strategies[serviceName]
	lm.mu.RUnlock()
	
	if !exists {
		lm.mu.Lock()
		defer lm.mu.Unlock()
		
		// 双重检查
		if balancer, exists := lm.balancers[serviceName]; exists {
			return balancer
		}
		
		// 使用默认策略或服务指定的策略
		if strategy == "" {
			strategy = lm.defaultStrategy
		}
		
		balancer = NewLoadBalancer(strategy)
		lm.balancers[serviceName] = balancer
		lm.strategies[serviceName] = strategy
		
		// 初始化统计信息
		lm.connectionStats[serviceName] = make(map[string]int64)
		lm.requestCounts[serviceName] = make(map[string]int64)
		lm.lastRequestTime[serviceName] = make(map[string]time.Time)
		
		lm.logger.WithFields(map[string]interface{}{
			"service":  serviceName,
			"strategy": string(strategy),
		}).Debug("Created load balancer for service")
	}
	
	return balancer
}

// SetLoadBalancerStrategy 设置服务的负载均衡策略
func (lm *loadBalancerManager) SetLoadBalancerStrategy(serviceName string, strategy types.LoadBalancerType) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	
	// 创建新的负载均衡器
	balancer := NewLoadBalancer(strategy)
	
	lm.balancers[serviceName] = balancer
	lm.strategies[serviceName] = strategy
	
	// 初始化统计信息（如果不存在）
	if lm.connectionStats[serviceName] == nil {
		lm.connectionStats[serviceName] = make(map[string]int64)
	}
	if lm.requestCounts[serviceName] == nil {
		lm.requestCounts[serviceName] = make(map[string]int64)
	}
	if lm.lastRequestTime[serviceName] == nil {
		lm.lastRequestTime[serviceName] = make(map[string]time.Time)
	}
	
	lm.logger.WithFields(map[string]interface{}{
		"service":  serviceName,
		"strategy": string(strategy),
	}).Info("Updated load balancer strategy for service")
	
	return nil
}

// SelectInstance 选择服务实例
func (lm *loadBalancerManager) SelectInstance(serviceName string, instances []*registry.ServiceInstance, clientIP string) (*registry.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("no available instances for service: %s", serviceName)
	}
	
	// 如果启用了健康检查，只使用健康的实例
	if lm.healthChecker != nil {
		healthyInstances := lm.healthChecker.GetHealthyInstances(serviceName)
		if len(healthyInstances) > 0 {
			// 转换为指针切片
			var healthyInstancePtrs []*registry.ServiceInstance
			for i := range healthyInstances {
				healthyInstancePtrs = append(healthyInstancePtrs, &healthyInstances[i])
			}
			instances = healthyInstancePtrs
		} else {
			// 如果没有健康的实例，记录警告但继续使用所有实例
			// 这样可以避免服务完全不可用
			lm.logger.WithFields(map[string]interface{}{
				"service": serviceName,
			}).Warn("No healthy instances available, using all instances")
		}
	}
	
	if len(instances) == 0 {
		return nil, fmt.Errorf("no healthy instances available for service: %s", serviceName)
	}
	
	balancer := lm.GetLoadBalancer(serviceName)
	
	// 特殊处理一致性哈希和IP哈希
	switch balancer.Algorithm() {
	case "consistent_hash", "consistent_hash_weighted":
		if chb, ok := balancer.(*ConsistentHashBalancer); ok {
			key := clientIP
			if key == "" {
				key = serviceName // 使用服务名作为fallback
			}
			return chb.SelectByKey(instances, key)
		}
		if chbw, ok := balancer.(*ConsistentHashBalancerWithWeight); ok {
			key := clientIP
			if key == "" {
				key = serviceName // 使用服务名作为fallback
			}
			return chbw.SelectByKey(instances, key)
		}
	case string(types.IPHash):
		if ipb, ok := balancer.(*ipHashBalancer); ok && clientIP != "" {
			return ipb.SelectWithIP(instances, clientIP)
		}
	}
	
	// 标准选择
	instance, err := balancer.Select(instances)
	if err != nil {
		return nil, err
	}
	
	// 更新统计信息
	lm.updateStats(serviceName, instance.ID)
	
	return instance, nil
}

// IncrementConnections 增加连接数
func (lm *loadBalancerManager) IncrementConnections(serviceName, instanceID string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	
	if lm.connectionStats[serviceName] == nil {
		lm.connectionStats[serviceName] = make(map[string]int64)
	}
	
	lm.connectionStats[serviceName][instanceID]++
	
	// 如果使用最少连接算法，更新负载均衡器的连接统计
	if balancer, exists := lm.balancers[serviceName]; exists {
		if lcb, ok := balancer.(*leastConnectionsBalancer); ok {
			lcb.IncrementConnections(instanceID)
		} else if hab, ok := balancer.(*HealthAwareLoadBalancer); ok {
			if lcb, ok := hab.balancer.(*leastConnectionsBalancer); ok {
				lcb.IncrementConnections(instanceID)
			}
		}
	}
}

// DecrementConnections 减少连接数
func (lm *loadBalancerManager) DecrementConnections(serviceName, instanceID string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	
	if lm.connectionStats[serviceName] == nil {
		return
	}
	
	if lm.connectionStats[serviceName][instanceID] > 0 {
		lm.connectionStats[serviceName][instanceID]--
	}
	
	// 如果使用最少连接算法，更新负载均衡器的连接统计
	if balancer, exists := lm.balancers[serviceName]; exists {
		if lcb, ok := balancer.(*leastConnectionsBalancer); ok {
			lcb.DecrementConnections(instanceID)
		} else if hab, ok := balancer.(*HealthAwareLoadBalancer); ok {
			if lcb, ok := hab.balancer.(*leastConnectionsBalancer); ok {
				lcb.DecrementConnections(instanceID)
			}
		}
	}
}

// updateStats 更新统计信息
func (lm *loadBalancerManager) updateStats(serviceName, instanceID string) {
	// 更新请求计数
	if lm.requestCounts[serviceName] == nil {
		lm.requestCounts[serviceName] = make(map[string]int64)
	}
	lm.requestCounts[serviceName][instanceID]++
	
	// 更新最后请求时间
	if lm.lastRequestTime[serviceName] == nil {
		lm.lastRequestTime[serviceName] = make(map[string]time.Time)
	}
	lm.lastRequestTime[serviceName][instanceID] = time.Now()
}

// GetStats 获取负载均衡统计信息
func (lm *loadBalancerManager) GetStats(serviceName string) map[string]interface{} {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	
	stats := make(map[string]interface{})
	
	// 基本信息
	if strategy, exists := lm.strategies[serviceName]; exists {
		stats["strategy"] = string(strategy)
	} else {
		stats["strategy"] = string(lm.defaultStrategy)
	}
	
	// 连接统计
	if connStats, exists := lm.connectionStats[serviceName]; exists {
		stats["connections"] = connStats
	}
	
	// 请求统计
	if reqStats, exists := lm.requestCounts[serviceName]; exists {
		stats["requests"] = reqStats
	}
	
	// 最后请求时间
	if timeStats, exists := lm.lastRequestTime[serviceName]; exists {
		lastTimes := make(map[string]string)
		for instanceID, t := range timeStats {
			lastTimes[instanceID] = t.Format(time.RFC3339)
		}
		stats["last_request_times"] = lastTimes
	}
	
	// 一致性哈希特殊信息
	if balancer, exists := lm.balancers[serviceName]; exists {
		if chb, ok := balancer.(*ConsistentHashBalancer); ok {
			stats["hash_ring_info"] = chb.GetRingInfo()
		} else if chbw, ok := balancer.(*ConsistentHashBalancerWithWeight); ok {
			stats["hash_ring_info"] = chbw.GetRingInfo()
		} else if hab, ok := balancer.(*HealthAwareLoadBalancer); ok {
			if chb, ok := hab.balancer.(*ConsistentHashBalancer); ok {
				stats["hash_ring_info"] = chb.GetRingInfo()
			} else if chbw, ok := hab.balancer.(*ConsistentHashBalancerWithWeight); ok {
				stats["hash_ring_info"] = chbw.GetRingInfo()
			}
		}
	}
	
	return stats
}

// RegisterHealthChecker 注册健康检查器
func (lm *loadBalancerManager) RegisterHealthChecker(healthChecker health.HealthChecker) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	
	lm.healthChecker = healthChecker
}

// Close 关闭管理器
func (lm *loadBalancerManager) Close() error {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	
	// 清理资源
	lm.balancers = make(map[string]LoadBalancer)
	lm.strategies = make(map[string]types.LoadBalancerType)
	lm.connectionStats = make(map[string]map[string]int64)
	lm.requestCounts = make(map[string]map[string]int64)
	lm.lastRequestTime = make(map[string]map[string]time.Time)
	
	lm.logger.Info("Load balancer manager closed")
	return nil
}

// LoadBalancerConfig 负载均衡配置
type LoadBalancerConfig struct {
	DefaultStrategy types.LoadBalancerType                    `yaml:"default_strategy" json:"default_strategy"`
	ServiceStrategies map[string]types.LoadBalancerType       `yaml:"service_strategies" json:"service_strategies"`
	ConsistentHashReplicas int                          `yaml:"consistent_hash_replicas" json:"consistent_hash_replicas"`
}

// ApplyConfig 应用负载均衡配置
func (lm *loadBalancerManager) ApplyConfig(config LoadBalancerConfig) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	
	// 更新默认策略
	if config.DefaultStrategy != "" {
		lm.defaultStrategy = config.DefaultStrategy
	}
	
	// 更新服务特定策略
	for serviceName, strategy := range config.ServiceStrategies {
		balancer := NewLoadBalancer(strategy)
		lm.balancers[serviceName] = balancer
		lm.strategies[serviceName] = strategy
		
		// 初始化统计信息（如果不存在）
		if lm.connectionStats[serviceName] == nil {
			lm.connectionStats[serviceName] = make(map[string]int64)
		}
		if lm.requestCounts[serviceName] == nil {
			lm.requestCounts[serviceName] = make(map[string]int64)
		}
		if lm.lastRequestTime[serviceName] == nil {
			lm.lastRequestTime[serviceName] = make(map[string]time.Time)
		}
		
		lm.logger.WithFields(map[string]interface{}{
			"service":  serviceName,
			"strategy": string(strategy),
		}).Info("Applied load balancer configuration for service")
	}
	
	return nil
}