package proxy

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/registry"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/types"
)

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
	// 选择服务实例
	Select(instances []*registry.ServiceInstance) (*registry.ServiceInstance, error)
	
	// 获取负载均衡算法名称
	Algorithm() string
}



// NewLoadBalancer 创建负载均衡器
func NewLoadBalancer(algorithm types.LoadBalancerType) LoadBalancer {
	switch algorithm {
	case types.RoundRobin:
		return &roundRobinBalancer{}
	case types.WeightedRoundRobin:
		return &weightedRoundRobinBalancer{}
	case types.Random:
		return &randomBalancer{}
	case types.WeightedRandom:
		return &weightedRandomBalancer{}
	case types.LeastConnections:
		return &leastConnectionsBalancer{
			connections: make(map[string]int64),
		}
	case types.IPHash:
		return &ipHashBalancer{}
	case types.ConsistentHash:
		return NewConsistentHashBalancer(150)
	case types.ConsistentHashWeighted:
		return NewConsistentHashBalancerWithWeight(150)
	default:
		return &roundRobinBalancer{}
	}
}

// roundRobinBalancer 轮询负载均衡器
type roundRobinBalancer struct {
	counter uint64
}

func (r *roundRobinBalancer) Select(instances []*registry.ServiceInstance) (*registry.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("no available instances")
	}
	
	index := atomic.AddUint64(&r.counter, 1) % uint64(len(instances))
	return instances[index], nil
}

func (r *roundRobinBalancer) Algorithm() string {
	return string(types.RoundRobin)
}

// weightedRoundRobinBalancer 加权轮询负载均衡器
type weightedRoundRobinBalancer struct {
	mu      sync.Mutex
	weights map[string]int
	current map[string]int
}

func (w *weightedRoundRobinBalancer) Select(instances []*registry.ServiceInstance) (*registry.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("no available instances")
	}
	
	w.mu.Lock()
	defer w.mu.Unlock()
	
	if w.weights == nil {
		w.weights = make(map[string]int)
		w.current = make(map[string]int)
	}
	
	// 更新权重
	totalWeight := 0
	for _, instance := range instances {
		w.weights[instance.ID] = instance.Weight
		totalWeight += instance.Weight
	}
	
	if totalWeight == 0 {
		// 如果所有权重都为0，使用轮询
		return instances[0], nil
	}
	
	// 选择实例
	var selected *registry.ServiceInstance
	maxCurrentWeight := -1
	
	for _, instance := range instances {
		w.current[instance.ID] += w.weights[instance.ID]
		
		if w.current[instance.ID] > maxCurrentWeight {
			maxCurrentWeight = w.current[instance.ID]
			selected = instance
		}
	}
	
	if selected != nil {
		w.current[selected.ID] -= totalWeight
	}
	
	return selected, nil
}

func (w *weightedRoundRobinBalancer) Algorithm() string {
	return string(types.WeightedRoundRobin)
}

// randomBalancer 随机负载均衡器
type randomBalancer struct {
	rand *rand.Rand
	mu   sync.Mutex
}

func (r *randomBalancer) Select(instances []*registry.ServiceInstance) (*registry.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("no available instances")
	}
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.rand == nil {
		r.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	
	index := r.rand.Intn(len(instances))
	return instances[index], nil
}

func (r *randomBalancer) Algorithm() string {
	return string(types.Random)
}

// weightedRandomBalancer 加权随机负载均衡器
type weightedRandomBalancer struct {
	rand *rand.Rand
	mu   sync.Mutex
}

func (w *weightedRandomBalancer) Select(instances []*registry.ServiceInstance) (*registry.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("no available instances")
	}
	
	w.mu.Lock()
	defer w.mu.Unlock()
	
	if w.rand == nil {
		w.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	
	// 计算总权重
	totalWeight := 0
	for _, instance := range instances {
		totalWeight += instance.Weight
	}
	
	if totalWeight == 0 {
		// 如果所有权重都为0，使用随机
		index := w.rand.Intn(len(instances))
		return instances[index], nil
	}
	
	// 生成随机数
	randomWeight := w.rand.Intn(totalWeight)
	
	// 选择实例
	currentWeight := 0
	for _, instance := range instances {
		currentWeight += instance.Weight
		if randomWeight < currentWeight {
			return instance, nil
		}
	}
	
	// 理论上不应该到达这里
	return instances[len(instances)-1], nil
}

func (w *weightedRandomBalancer) Algorithm() string {
	return string(types.WeightedRandom)
}

// leastConnectionsBalancer 最少连接负载均衡器
type leastConnectionsBalancer struct {
	mu          sync.RWMutex
	connections map[string]int64
}

func (l *leastConnectionsBalancer) Select(instances []*registry.ServiceInstance) (*registry.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("no available instances")
	}
	
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	var selected *registry.ServiceInstance
	minConnections := int64(-1)
	
	for _, instance := range instances {
		connections := l.connections[instance.ID]
		if minConnections == -1 || connections < minConnections {
			minConnections = connections
			selected = instance
		}
	}
	
	return selected, nil
}

func (l *leastConnectionsBalancer) Algorithm() string {
	return string(types.LeastConnections)
}

// IncrementConnections 增加连接数
func (l *leastConnectionsBalancer) IncrementConnections(instanceID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.connections[instanceID]++
}

// DecrementConnections 减少连接数
func (l *leastConnectionsBalancer) DecrementConnections(instanceID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.connections[instanceID] > 0 {
		l.connections[instanceID]--
	}
}

// ipHashBalancer IP哈希负载均衡器
type ipHashBalancer struct{}

func (i *ipHashBalancer) Select(instances []*registry.ServiceInstance) (*registry.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("no available instances")
	}
	
	// 注意：这里需要客户端IP，但在这个接口中没有提供
	// 实际使用时需要修改接口或使用上下文传递IP
	// 这里使用简单的轮询作为fallback
	return instances[0], nil
}

func (i *ipHashBalancer) Algorithm() string {
	return string(types.IPHash)
}

// SelectWithIP IP哈希选择（需要客户端IP）
func (i *ipHashBalancer) SelectWithIP(instances []*registry.ServiceInstance, clientIP string) (*registry.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("no available instances")
	}
	
	// 简单的哈希算法
	hash := 0
	for _, b := range []byte(clientIP) {
		hash = hash*31 + int(b)
	}
	
	if hash < 0 {
		hash = -hash
	}
	
	index := hash % len(instances)
	return instances[index], nil
}

// HealthAwareLoadBalancer 健康感知负载均衡器包装器
type HealthAwareLoadBalancer struct {
	balancer LoadBalancer
}

// NewHealthAwareLoadBalancer 创建健康感知负载均衡器
func NewHealthAwareLoadBalancer(balancer LoadBalancer) *HealthAwareLoadBalancer {
	return &HealthAwareLoadBalancer{
		balancer: balancer,
	}
}

// Select 选择健康的服务实例
func (h *HealthAwareLoadBalancer) Select(instances []*registry.ServiceInstance) (*registry.ServiceInstance, error) {
	// 过滤健康的实例
	var healthyInstances []*registry.ServiceInstance
	for _, instance := range instances {
		if instance.Health == registry.HealthStatusHealthy {
			healthyInstances = append(healthyInstances, instance)
		}
	}
	
	if len(healthyInstances) == 0 {
		return nil, fmt.Errorf("no healthy instances available")
	}
	
	return h.balancer.Select(healthyInstances)
}

// Algorithm 获取算法名称
func (h *HealthAwareLoadBalancer) Algorithm() string {
	return h.balancer.Algorithm() + "_health_aware"
}