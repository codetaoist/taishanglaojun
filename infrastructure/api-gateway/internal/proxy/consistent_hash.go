package proxy

import (
	"crypto/sha1"
	"fmt"
	"sort"
	"strconv"
	"sync"

	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/registry"
)

// ConsistentHashBalancer 一致性哈希负载均衡器
type ConsistentHashBalancer struct {
	mu           sync.RWMutex
	hashRing     map[uint32]string // 哈希环，key为哈希值，value为实例ID
	sortedHashes []uint32          // 排序的哈希值
	replicas     int               // 虚拟节点数量
	instances    map[string]*registry.ServiceInstance // 实例映射
}

// NewConsistentHashBalancer 创建一致性哈希负载均衡器
func NewConsistentHashBalancer(replicas int) *ConsistentHashBalancer {
	if replicas <= 0 {
		replicas = 150 // 默认虚拟节点数量
	}
	
	return &ConsistentHashBalancer{
		hashRing:  make(map[uint32]string),
		replicas:  replicas,
		instances: make(map[string]*registry.ServiceInstance),
	}
}

// Select 选择服务实例
func (c *ConsistentHashBalancer) Select(instances []*registry.ServiceInstance) (*registry.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("no available instances")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 更新哈希环
	c.updateHashRing(instances)

	// 如果没有提供key，使用第一个实例
	if len(c.sortedHashes) == 0 {
		return instances[0], nil
	}

	// 使用第一个实例的ID作为默认key
	key := instances[0].ID
	return c.selectByKey(key)
}

// GetRingInfo 获取哈希环信息（用于统计和调试）
func (c *ConsistentHashBalancer) GetRingInfo() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	info := make(map[string]interface{})
	info["virtual_nodes"] = c.replicas
	info["ring_size"] = len(c.hashRing)
	
	// 统计每个实例的虚拟节点数量
	nodeCount := make(map[string]int)
	for _, instanceID := range c.hashRing {
		nodeCount[instanceID]++
	}
	info["node_distribution"] = nodeCount
	
	return info
}

// SelectByKey 根据指定的key选择服务实例
func (c *ConsistentHashBalancer) SelectByKey(instances []*registry.ServiceInstance, key string) (*registry.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("no available instances")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 更新哈希环
	c.updateHashRing(instances)

	return c.selectByKey(key)
}

// GetRingInfo 获取哈希环信息（用于统计和调试）
func (c *ConsistentHashBalancerWithWeight) GetRingInfo() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	info := make(map[string]interface{})
	info["virtual_nodes"] = c.replicas
	info["ring_size"] = len(c.hashRing)
	
	// 统计每个实例的虚拟节点数量
	nodeCount := make(map[string]int)
	for _, instanceID := range c.hashRing {
		nodeCount[instanceID]++
	}
	info["node_distribution"] = nodeCount
	
	return info
}

// SelectByKey 根据指定的key选择服务实例
func (c *ConsistentHashBalancerWithWeight) SelectByKey(instances []*registry.ServiceInstance, key string) (*registry.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("no available instances")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 更新哈希环
	c.updateHashRing(instances)

	return c.selectByKey(key)
}

// selectByKey 内部方法，根据key选择实例
func (c *ConsistentHashBalancer) selectByKey(key string) (*registry.ServiceInstance, error) {
	if len(c.sortedHashes) == 0 {
		return nil, fmt.Errorf("no instances in hash ring")
	}

	hash := c.hashKey(key)

	// 在哈希环上查找第一个大于等于hash的节点
	idx := sort.Search(len(c.sortedHashes), func(i int) bool {
		return c.sortedHashes[i] >= hash
	})

	// 如果没找到，使用第一个节点（环形结构）
	if idx == len(c.sortedHashes) {
		idx = 0
	}

	instanceID := c.hashRing[c.sortedHashes[idx]]
	instance, exists := c.instances[instanceID]
	if !exists {
		return nil, fmt.Errorf("instance not found: %s", instanceID)
	}

	return instance, nil
}

// updateHashRing 更新哈希环
func (c *ConsistentHashBalancer) updateHashRing(instances []*registry.ServiceInstance) {
	// 清空现有的哈希环
	c.hashRing = make(map[uint32]string)
	c.instances = make(map[string]*registry.ServiceInstance)
	c.sortedHashes = nil

	// 添加实例到哈希环
	for _, instance := range instances {
		c.instances[instance.ID] = instance
		c.addInstanceToRing(instance.ID)
	}

	// 排序哈希值
	sort.Slice(c.sortedHashes, func(i, j int) bool {
		return c.sortedHashes[i] < c.sortedHashes[j]
	})
}

// addInstanceToRing 添加实例到哈希环
func (c *ConsistentHashBalancer) addInstanceToRing(instanceID string) {
	for i := 0; i < c.replicas; i++ {
		virtualKey := instanceID + "#" + strconv.Itoa(i)
		hash := c.hashKey(virtualKey)
		c.hashRing[hash] = instanceID
		c.sortedHashes = append(c.sortedHashes, hash)
	}
}

// hashKey 计算key的哈希值
func (c *ConsistentHashBalancer) hashKey(key string) uint32 {
	h := sha1.New()
	h.Write([]byte(key))
	hashBytes := h.Sum(nil)
	
	// 取前4个字节转换为uint32
	return uint32(hashBytes[0])<<24 | uint32(hashBytes[1])<<16 | uint32(hashBytes[2])<<8 | uint32(hashBytes[3])
}

// Algorithm 获取算法名称
func (c *ConsistentHashBalancer) Algorithm() string {
	return "consistent_hash"
}

// GetRingInfo 获取哈希环信息（用于调试）


// ConsistentHashBalancerWithWeight 带权重的一致性哈希负载均衡器
type ConsistentHashBalancerWithWeight struct {
	*ConsistentHashBalancer
}

// NewConsistentHashBalancerWithWeight 创建带权重的一致性哈希负载均衡器
func NewConsistentHashBalancerWithWeight(baseReplicas int) *ConsistentHashBalancerWithWeight {
	return &ConsistentHashBalancerWithWeight{
		ConsistentHashBalancer: NewConsistentHashBalancer(baseReplicas),
	}
}

// updateHashRing 重写更新哈希环方法，支持权重
func (c *ConsistentHashBalancerWithWeight) updateHashRing(instances []*registry.ServiceInstance) {
	// 清空现有的哈希环
	c.hashRing = make(map[uint32]string)
	c.instances = make(map[string]*registry.ServiceInstance)
	c.sortedHashes = nil

	// 添加实例到哈希环，根据权重调整虚拟节点数量
	for _, instance := range instances {
		c.instances[instance.ID] = instance
		
		// 根据权重计算虚拟节点数量
		weight := instance.Weight
		if weight <= 0 {
			weight = 1 // 默认权重为1
		}
		
		replicas := c.replicas * weight
		c.addInstanceToRingWithReplicas(instance.ID, replicas)
	}

	// 排序哈希值
	sort.Slice(c.sortedHashes, func(i, j int) bool {
		return c.sortedHashes[i] < c.sortedHashes[j]
	})
}

// addInstanceToRingWithReplicas 添加指定数量的虚拟节点
func (c *ConsistentHashBalancerWithWeight) addInstanceToRingWithReplicas(instanceID string, replicas int) {
	for i := 0; i < replicas; i++ {
		virtualKey := instanceID + "#" + strconv.Itoa(i)
		hash := c.hashKey(virtualKey)
		c.hashRing[hash] = instanceID
		c.sortedHashes = append(c.sortedHashes, hash)
	}
}

// Algorithm 获取算法名称
func (c *ConsistentHashBalancerWithWeight) Algorithm() string {
	return "consistent_hash_weighted"
}