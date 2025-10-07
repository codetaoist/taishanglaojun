package types

// LoadBalancerType 负载均衡类型
type LoadBalancerType string

const (
	// RoundRobin 轮询
	RoundRobin LoadBalancerType = "round_robin"
	
	// WeightedRoundRobin 加权轮询
	WeightedRoundRobin LoadBalancerType = "weighted_round_robin"
	
	// LeastConnections 最少连接
	LeastConnections LoadBalancerType = "least_connections"
	
	// WeightedLeastConnections 加权最少连接
	WeightedLeastConnections LoadBalancerType = "weighted_least_connections"
	
	// IPHash IP哈希
	IPHash LoadBalancerType = "ip_hash"
	
	// ConsistentHash 一致性哈希
	ConsistentHash LoadBalancerType = "consistent_hash"
	
	// ConsistentHashWeighted 带权重的一致性哈希
	ConsistentHashWeighted LoadBalancerType = "consistent_hash_weighted"
	
	// Random 随机
	Random LoadBalancerType = "random"
	
	// WeightedRandom 加权随机
	WeightedRandom LoadBalancerType = "weighted_random"
)

// String 返回负载均衡类型的字符串表示
func (lbt LoadBalancerType) String() string {
	return string(lbt)
}

// IsValid 检查负载均衡类型是否有效
func (lbt LoadBalancerType) IsValid() bool {
	switch lbt {
	case RoundRobin, WeightedRoundRobin, LeastConnections, WeightedLeastConnections,
		 IPHash, ConsistentHash, ConsistentHashWeighted, Random, WeightedRandom:
		return true
	default:
		return false
	}
}