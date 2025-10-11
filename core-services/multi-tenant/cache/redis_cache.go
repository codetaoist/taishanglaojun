package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	multitenant "github.com/codetaoist/taishanglaojun/core-services/multi-tenant"
	"go.uber.org/zap"
)

// RedisTenantCache Redis租户缓存实现
type RedisTenantCache struct {
	client *redis.Client
	logger *zap.Logger
	config RedisCacheConfig
}

// RedisCacheConfig Redis缓存配置
type RedisCacheConfig struct {
	// 键前缀
	KeyPrefix string `json:"key_prefix"`
	
	// 默认过期时间
	DefaultTTL time.Duration `json:"default_ttl"`
	
	// 租户信息过期时间
	TenantTTL time.Duration `json:"tenant_ttl"`
	
	// 租户上下文过期时�?
	ContextTTL time.Duration `json:"context_ttl"`
	
	// 使用情况过期时间
	UsageTTL time.Duration `json:"usage_ttl"`
	
	// 是否启用压缩
	EnableCompression bool `json:"enable_compression"`
	
	// 序列化格�?
	SerializationFormat string `json:"serialization_format"`
}

// NewRedisTenantCache 创建Redis租户缓存
func NewRedisTenantCache(
	client *redis.Client,
	config RedisCacheConfig,
	logger *zap.Logger,
) *RedisTenantCache {
	// 设置默认�?
	if config.KeyPrefix == "" {
		config.KeyPrefix = "tenant:"
	}
	if config.DefaultTTL == 0 {
		config.DefaultTTL = 1 * time.Hour
	}
	if config.TenantTTL == 0 {
		config.TenantTTL = 30 * time.Minute
	}
	if config.ContextTTL == 0 {
		config.ContextTTL = 15 * time.Minute
	}
	if config.UsageTTL == 0 {
		config.UsageTTL = 5 * time.Minute
	}
	if config.SerializationFormat == "" {
		config.SerializationFormat = "json"
	}

	return &RedisTenantCache{
		client: client,
		config: config,
		logger: logger,
	}
}

// SetTenant 设置租户缓存
func (c *RedisTenantCache) SetTenant(ctx context.Context, tenant *multitenant.Tenant) error {
	key := c.getTenantKey(tenant.ID)
	
	data, err := c.serialize(tenant)
	if err != nil {
		return fmt.Errorf("failed to serialize tenant: %w", err)
	}

	err = c.client.Set(ctx, key, data, c.config.TenantTTL).Err()
	if err != nil {
		c.logger.Error("Failed to set tenant cache",
			zap.String("tenant_id", tenant.ID),
			zap.String("key", key),
			zap.Error(err))
		return fmt.Errorf("failed to set tenant cache: %w", err)
	}

	// 同时设置域名映射
	if tenant.Domain != "" {
		domainKey := c.getDomainKey(tenant.Domain)
		err = c.client.Set(ctx, domainKey, tenant.ID, c.config.TenantTTL).Err()
		if err != nil {
			c.logger.Warn("Failed to set domain mapping cache",
				zap.String("domain", tenant.Domain),
				zap.String("tenant_id", tenant.ID),
				zap.Error(err))
		}
	}

	c.logger.Debug("Tenant cached",
		zap.String("tenant_id", tenant.ID),
		zap.String("key", key))

	return nil
}

// GetTenant 获取租户缓存
func (c *RedisTenantCache) GetTenant(ctx context.Context, tenantID string) (*multitenant.Tenant, error) {
	key := c.getTenantKey(tenantID)
	
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, &CacheError{
				Code:    "CACHE_MISS",
				Message: "Tenant not found in cache",
				Details: map[string]interface{}{"tenant_id": tenantID},
			}
		}
		c.logger.Error("Failed to get tenant cache",
			zap.String("tenant_id", tenantID),
			zap.String("key", key),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get tenant cache: %w", err)
	}

	var tenant multitenant.Tenant
	err = c.deserialize([]byte(data), &tenant)
	if err != nil {
		c.logger.Error("Failed to deserialize tenant",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to deserialize tenant: %w", err)
	}

	c.logger.Debug("Tenant cache hit",
		zap.String("tenant_id", tenantID),
		zap.String("key", key))

	return &tenant, nil
}

// GetTenantByDomain 根据域名获取租户缓存
func (c *RedisTenantCache) GetTenantByDomain(ctx context.Context, domain string) (*multitenant.Tenant, error) {
	// 首先获取域名映射
	domainKey := c.getDomainKey(domain)
	tenantID, err := c.client.Get(ctx, domainKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, &CacheError{
				Code:    "CACHE_MISS",
				Message: "Domain mapping not found in cache",
				Details: map[string]interface{}{"domain": domain},
			}
		}
		return nil, fmt.Errorf("failed to get domain mapping: %w", err)
	}

	// 然后获取租户信息
	return c.GetTenant(ctx, tenantID)
}

// DeleteTenant 删除租户缓存
func (c *RedisTenantCache) DeleteTenant(ctx context.Context, tenantID string) error {
	// 先获取租户信息以获取域名
	tenant, err := c.GetTenant(ctx, tenantID)
	if err != nil && !isCacheMiss(err) {
		c.logger.Warn("Failed to get tenant for deletion",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
	}

	// 删除租户缓存
	key := c.getTenantKey(tenantID)
	err = c.client.Del(ctx, key).Err()
	if err != nil {
		c.logger.Error("Failed to delete tenant cache",
			zap.String("tenant_id", tenantID),
			zap.String("key", key),
			zap.Error(err))
		return fmt.Errorf("failed to delete tenant cache: %w", err)
	}

	// 删除域名映射
	if tenant != nil && tenant.Domain != "" {
		domainKey := c.getDomainKey(tenant.Domain)
		err = c.client.Del(ctx, domainKey).Err()
		if err != nil {
			c.logger.Warn("Failed to delete domain mapping cache",
				zap.String("domain", tenant.Domain),
				zap.String("tenant_id", tenantID),
				zap.Error(err))
		}
	}

	// 删除相关的上下文缓存
	contextPattern := c.getContextKeyPattern(tenantID)
	err = c.deleteByPattern(ctx, contextPattern)
	if err != nil {
		c.logger.Warn("Failed to delete tenant context caches",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
	}

	c.logger.Debug("Tenant cache deleted",
		zap.String("tenant_id", tenantID),
		zap.String("key", key))

	return nil
}

// SetTenantContext 设置租户上下文缓�?
func (c *RedisTenantCache) SetTenantContext(ctx context.Context, tenantID, userID string, tenantContext *multitenant.TenantContext) error {
	key := c.getContextKey(tenantID, userID)
	
	data, err := c.serialize(tenantContext)
	if err != nil {
		return fmt.Errorf("failed to serialize tenant context: %w", err)
	}

	err = c.client.Set(ctx, key, data, c.config.ContextTTL).Err()
	if err != nil {
		c.logger.Error("Failed to set tenant context cache",
			zap.String("tenant_id", tenantID),
			zap.String("user_id", userID),
			zap.String("key", key),
			zap.Error(err))
		return fmt.Errorf("failed to set tenant context cache: %w", err)
	}

	c.logger.Debug("Tenant context cached",
		zap.String("tenant_id", tenantID),
		zap.String("user_id", userID),
		zap.String("key", key))

	return nil
}

// GetTenantContext 获取租户上下文缓�?
func (c *RedisTenantCache) GetTenantContext(ctx context.Context, tenantID, userID string) (*multitenant.TenantContext, error) {
	key := c.getContextKey(tenantID, userID)
	
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, &CacheError{
				Code:    "CACHE_MISS",
				Message: "Tenant context not found in cache",
				Details: map[string]interface{}{
					"tenant_id": tenantID,
					"user_id":   userID,
				},
			}
		}
		c.logger.Error("Failed to get tenant context cache",
			zap.String("tenant_id", tenantID),
			zap.String("user_id", userID),
			zap.String("key", key),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get tenant context cache: %w", err)
	}

	var tenantContext multitenant.TenantContext
	err = c.deserialize([]byte(data), &tenantContext)
	if err != nil {
		c.logger.Error("Failed to deserialize tenant context",
			zap.String("tenant_id", tenantID),
			zap.String("user_id", userID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to deserialize tenant context: %w", err)
	}

	c.logger.Debug("Tenant context cache hit",
		zap.String("tenant_id", tenantID),
		zap.String("user_id", userID),
		zap.String("key", key))

	return &tenantContext, nil
}

// DeleteTenantContext 删除租户上下文缓�?
func (c *RedisTenantCache) DeleteTenantContext(ctx context.Context, tenantID, userID string) error {
	key := c.getContextKey(tenantID, userID)
	
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		c.logger.Error("Failed to delete tenant context cache",
			zap.String("tenant_id", tenantID),
			zap.String("user_id", userID),
			zap.String("key", key),
			zap.Error(err))
		return fmt.Errorf("failed to delete tenant context cache: %w", err)
	}

	c.logger.Debug("Tenant context cache deleted",
		zap.String("tenant_id", tenantID),
		zap.String("user_id", userID),
		zap.String("key", key))

	return nil
}

// SetUsage 设置使用情况缓存
func (c *RedisTenantCache) SetUsage(ctx context.Context, tenantID string, usage []*multitenant.TenantUsage) error {
	key := c.getUsageKey(tenantID)
	
	data, err := c.serialize(usage)
	if err != nil {
		return fmt.Errorf("failed to serialize usage: %w", err)
	}

	err = c.client.Set(ctx, key, data, c.config.UsageTTL).Err()
	if err != nil {
		c.logger.Error("Failed to set usage cache",
			zap.String("tenant_id", tenantID),
			zap.String("key", key),
			zap.Error(err))
		return fmt.Errorf("failed to set usage cache: %w", err)
	}

	c.logger.Debug("Usage cached",
		zap.String("tenant_id", tenantID),
		zap.String("key", key),
		zap.Int("count", len(usage)))

	return nil
}

// GetUsage 获取使用情况缓存
func (c *RedisTenantCache) GetUsage(ctx context.Context, tenantID string) ([]*multitenant.TenantUsage, error) {
	key := c.getUsageKey(tenantID)
	
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, &CacheError{
				Code:    "CACHE_MISS",
				Message: "Usage not found in cache",
				Details: map[string]interface{}{"tenant_id": tenantID},
			}
		}
		c.logger.Error("Failed to get usage cache",
			zap.String("tenant_id", tenantID),
			zap.String("key", key),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get usage cache: %w", err)
	}

	var usage []*multitenant.TenantUsage
	err = c.deserialize([]byte(data), &usage)
	if err != nil {
		c.logger.Error("Failed to deserialize usage",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to deserialize usage: %w", err)
	}

	c.logger.Debug("Usage cache hit",
		zap.String("tenant_id", tenantID),
		zap.String("key", key),
		zap.Int("count", len(usage)))

	return usage, nil
}

// DeleteUsage 删除使用情况缓存
func (c *RedisTenantCache) DeleteUsage(ctx context.Context, tenantID string) error {
	key := c.getUsageKey(tenantID)
	
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		c.logger.Error("Failed to delete usage cache",
			zap.String("tenant_id", tenantID),
			zap.String("key", key),
			zap.Error(err))
		return fmt.Errorf("failed to delete usage cache: %w", err)
	}

	c.logger.Debug("Usage cache deleted",
		zap.String("tenant_id", tenantID),
		zap.String("key", key))

	return nil
}

// Clear 清空所有缓�?
func (c *RedisTenantCache) Clear(ctx context.Context) error {
	pattern := c.config.KeyPrefix + "*"
	
	err := c.deleteByPattern(ctx, pattern)
	if err != nil {
		c.logger.Error("Failed to clear cache", zap.Error(err))
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	c.logger.Info("Cache cleared", zap.String("pattern", pattern))
	return nil
}

// HealthCheck 健康检�?
func (c *RedisTenantCache) HealthCheck(ctx context.Context) error {
	// 执行PING命令
	err := c.client.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	// 测试基本操作
	testKey := c.config.KeyPrefix + "health_check"
	testValue := "ok"
	
	// 设置测试�?
	err = c.client.Set(ctx, testKey, testValue, 10*time.Second).Err()
	if err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}

	// 获取测试�?
	result, err := c.client.Get(ctx, testKey).Result()
	if err != nil {
		return fmt.Errorf("redis get failed: %w", err)
	}

	if result != testValue {
		return fmt.Errorf("redis value mismatch: expected %s, got %s", testValue, result)
	}

	// 删除测试�?
	err = c.client.Del(ctx, testKey).Err()
	if err != nil {
		return fmt.Errorf("redis del failed: %w", err)
	}

	return nil
}

// 辅助方法

// getTenantKey 获取租户�?
func (c *RedisTenantCache) getTenantKey(tenantID string) string {
	return fmt.Sprintf("%stenant:%s", c.config.KeyPrefix, tenantID)
}

// getDomainKey 获取域名�?
func (c *RedisTenantCache) getDomainKey(domain string) string {
	return fmt.Sprintf("%sdomain:%s", c.config.KeyPrefix, domain)
}

// getContextKey 获取上下文键
func (c *RedisTenantCache) getContextKey(tenantID, userID string) string {
	return fmt.Sprintf("%scontext:%s:%s", c.config.KeyPrefix, tenantID, userID)
}

// getContextKeyPattern 获取上下文键模式
func (c *RedisTenantCache) getContextKeyPattern(tenantID string) string {
	return fmt.Sprintf("%scontext:%s:*", c.config.KeyPrefix, tenantID)
}

// getUsageKey 获取使用情况�?
func (c *RedisTenantCache) getUsageKey(tenantID string) string {
	return fmt.Sprintf("%susage:%s", c.config.KeyPrefix, tenantID)
}

// serialize 序列化数�?
func (c *RedisTenantCache) serialize(data interface{}) ([]byte, error) {
	switch c.config.SerializationFormat {
	case "json":
		return json.Marshal(data)
	default:
		return json.Marshal(data)
	}
}

// deserialize 反序列化数据
func (c *RedisTenantCache) deserialize(data []byte, target interface{}) error {
	switch c.config.SerializationFormat {
	case "json":
		return json.Unmarshal(data, target)
	default:
		return json.Unmarshal(data, target)
	}
}

// deleteByPattern 按模式删除键
func (c *RedisTenantCache) deleteByPattern(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	if len(keys) > 0 {
		err := c.client.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
		
		c.logger.Debug("Keys deleted by pattern",
			zap.String("pattern", pattern),
			zap.Int("count", len(keys)))
	}

	return nil
}

// CacheError 缓存错误
type CacheError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *CacheError) Error() string {
	return e.Message
}

// isCacheMiss 检查是否为缓存未命中错�?
func isCacheMiss(err error) bool {
	if cacheErr, ok := err.(*CacheError); ok {
		return cacheErr.Code == "CACHE_MISS"
	}
	return false
}
