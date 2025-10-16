package config

import (
	"fmt"
	"time"
	
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/types"
)

// LoadBalancerConfig 负载均衡配置
type LoadBalancerConfig struct {
	// 默认负载均衡策略
	DefaultStrategy types.LoadBalancerType `yaml:"default_strategy" json:"default_strategy"`
	
	// 服务特定的负载均衡策略
	ServiceStrategies map[string]types.LoadBalancerType `yaml:"service_strategies" json:"service_strategies"`
	
	// 一致性哈希配置
	ConsistentHash ConsistentHashConfig `yaml:"consistent_hash" json:"consistent_hash"`
	
	// 健康检查配置
	HealthCheck LoadBalancerHealthCheckConfig `yaml:"health_check" json:"health_check"`
	
	// 连接统计配置
	ConnectionStats ConnectionStatsConfig `yaml:"connection_stats" json:"connection_stats"`
}

// ConsistentHashConfig 一致性哈希配置
type ConsistentHashConfig struct {
	// 虚拟节点数量
	VirtualNodes int `yaml:"virtual_nodes" json:"virtual_nodes"`
	
	// 哈希函数类型 (sha1, md5, fnv)
	HashFunction string `yaml:"hash_function" json:"hash_function"`
}

// LoadBalancerHealthCheckConfig 负载均衡器健康检查配置
type LoadBalancerHealthCheckConfig struct {
	// 是否启用健康检查
	Enabled bool `yaml:"enabled" json:"enabled"`
	
	// 检查间隔
	Interval time.Duration `yaml:"interval" json:"interval"`
	
	// 超时时间
	Timeout time.Duration `yaml:"timeout" json:"timeout"`
	
	// 健康检查路径
	Path string `yaml:"path" json:"path"`
	
	// 重试次数
	Retries int `yaml:"retries" json:"retries"`
	
	// 失败阈值
	FailureThreshold int `yaml:"failure_threshold" json:"failure_threshold"`
	
	// 成功阈值
	SuccessThreshold int `yaml:"success_threshold" json:"success_threshold"`
	
	// 期望的状态码
	ExpectedStatusCodes []int `yaml:"expected_status_codes" json:"expected_status_codes"`
}

// ConnectionStatsConfig 连接统计配置
type ConnectionStatsConfig struct {
	// 是否启用连接统计
	Enabled bool `yaml:"enabled" json:"enabled"`
	
	// 统计数据保留时间
	RetentionPeriod time.Duration `yaml:"retention_period" json:"retention_period"`
	
	// 统计数据清理间隔
	CleanupInterval time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
}

// GetDefaultLoadBalancerConfig 获取默认负载均衡配置
func GetDefaultLoadBalancerConfig() LoadBalancerConfig {
	return LoadBalancerConfig{
		DefaultStrategy: types.RoundRobin,
		ServiceStrategies: map[string]types.LoadBalancerType{
			"auth-service":            types.LeastConnections,
			"cultural-wisdom-service": types.WeightedRoundRobin,
		},
		ConsistentHash: ConsistentHashConfig{
			VirtualNodes: 150,
			HashFunction: "sha1",
		},
		HealthCheck: LoadBalancerHealthCheckConfig{
			Enabled:             true,
			Interval:            30 * time.Second,
			Timeout:             5 * time.Second,
			Path:                "/health",
			Retries:             3,
			FailureThreshold:    3,
			SuccessThreshold:    2,
			ExpectedStatusCodes: []int{200, 201, 202},
		},
		ConnectionStats: ConnectionStatsConfig{
			Enabled:         true,
			RetentionPeriod: 24 * time.Hour,
			CleanupInterval: 1 * time.Hour,
		},
	}
}

// ValidateLoadBalancerConfig 验证负载均衡配置
func ValidateLoadBalancerConfig(config LoadBalancerConfig) error {
	// 验证默认策略
	if !isValidLoadBalancerType(config.DefaultStrategy) {
		return fmt.Errorf("invalid default load balancer strategy: %s", config.DefaultStrategy)
	}
	
	// 验证服务特定策略
	for service, strategy := range config.ServiceStrategies {
		if !isValidLoadBalancerType(strategy) {
			return fmt.Errorf("invalid load balancer strategy for service %s: %s", service, strategy)
		}
	}
	
	// 验证一致性哈希配置
	if config.ConsistentHash.VirtualNodes <= 0 {
		return fmt.Errorf("virtual nodes must be greater than 0")
	}
	
	// 验证健康检查配置
	if config.HealthCheck.Enabled {
		if config.HealthCheck.Interval <= 0 {
			return fmt.Errorf("health check interval must be greater than 0")
		}
		if config.HealthCheck.Timeout <= 0 {
			return fmt.Errorf("health check timeout must be greater than 0")
		}
		if config.HealthCheck.FailureThreshold <= 0 {
			return fmt.Errorf("failure threshold must be greater than 0")
		}
		if config.HealthCheck.SuccessThreshold <= 0 {
			return fmt.Errorf("success threshold must be greater than 0")
		}
	}
	
	return nil
}

// isValidLoadBalancerType 检查负载均衡类型是否有效
func isValidLoadBalancerType(lbType types.LoadBalancerType) bool {
	return lbType.IsValid()
}

// LoadBalancerConfigFromFile 从文件加载负载均衡配置
func LoadBalancerConfigFromFile(filename string) (LoadBalancerConfig, error) {
	var config LoadBalancerConfig
	
	// 这里可以添加从YAML或JSON文件加载配置的逻辑
	// 目前返回默认配置
	config = GetDefaultLoadBalancerConfig()
	
	// 验证配置
	if err := ValidateLoadBalancerConfig(config); err != nil {
		return config, fmt.Errorf("invalid load balancer configuration: %w", err)
	}
	
	return config, nil
}