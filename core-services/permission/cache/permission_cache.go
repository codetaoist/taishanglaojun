package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/core-services/permission"
)

// RedisPermissionCache Redis权限缓存实现
type RedisPermissionCache struct {
	client *redis.Client
	logger *zap.Logger
	config RedisPermissionCacheConfig
}

// RedisPermissionCacheConfig Redis权限缓存配置
type RedisPermissionCacheConfig struct {
	// 基本配置
	KeyPrefix        string `json:"key_prefix"`
	DefaultTTL       time.Duration `json:"default_ttl"`
	
	// TTL配置
	RoleTTL          time.Duration `json:"role_ttl"`
	PermissionTTL    time.Duration `json:"permission_ttl"`
	UserRolesTTL     time.Duration `json:"user_roles_ttl"`
	RolePermsTTL     time.Duration `json:"role_permissions_ttl"`
	CheckResultTTL   time.Duration `json:"check_result_ttl"`
	
	// 性能配置
	EnableCompression bool `json:"enable_compression"`
	BatchSize        int  `json:"batch_size"`
	MaxRetries       int  `json:"max_retries"`
	RetryDelay       time.Duration `json:"retry_delay"`
	
	// 序列化配�?
	SerializationFormat string `json:"serialization_format"` // json, msgpack
	
	// 监控配置
	EnableMetrics    bool `json:"enable_metrics"`
	MetricsPrefix    string `json:"metrics_prefix"`
}

// CacheError 缓存错误
type CacheError struct {
	Operation string
	Key       string
	Err       error
}

func (e *CacheError) Error() string {
	return fmt.Sprintf("cache %s failed for key %s: %v", e.Operation, e.Key, e.Err)
}

// NewRedisPermissionCache 创建Redis权限缓存
func NewRedisPermissionCache(client *redis.Client, logger *zap.Logger, config RedisPermissionCacheConfig) permission.PermissionCache {
	// 设置默认配置
	if config.KeyPrefix == "" {
		config.KeyPrefix = "perm:"
	}
	if config.DefaultTTL == 0 {
		config.DefaultTTL = 1 * time.Hour
	}
	if config.RoleTTL == 0 {
		config.RoleTTL = 2 * time.Hour
	}
	if config.PermissionTTL == 0 {
		config.PermissionTTL = 2 * time.Hour
	}
	if config.UserRolesTTL == 0 {
		config.UserRolesTTL = 30 * time.Minute
	}
	if config.RolePermsTTL == 0 {
		config.RolePermsTTL = 1 * time.Hour
	}
	if config.CheckResultTTL == 0 {
		config.CheckResultTTL = 5 * time.Minute
	}
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 100 * time.Millisecond
	}
	if config.SerializationFormat == "" {
		config.SerializationFormat = "json"
	}
	if config.MetricsPrefix == "" {
		config.MetricsPrefix = "permission_cache"
	}

	return &RedisPermissionCache{
		client: client,
		logger: logger,
		config: config,
	}
}

// SetRole 设置角色缓存
func (c *RedisPermissionCache) SetRole(ctx context.Context, role *permission.Role) error {
	key := c.buildRoleKey(role.ID)
	data, err := c.serialize(role)
	if err != nil {
		return &CacheError{Operation: "serialize", Key: key, Err: err}
	}

	err = c.client.Set(ctx, key, data, c.config.RoleTTL).Err()
	if err != nil {
		c.logger.Error("Failed to set role cache", zap.String("key", key), zap.Error(err))
		return &CacheError{Operation: "set", Key: key, Err: err}
	}

	c.logger.Debug("Role cached", zap.String("role_id", role.ID), zap.String("key", key))
	return nil
}

// GetRole 获取角色缓存
func (c *RedisPermissionCache) GetRole(ctx context.Context, roleID string) (*permission.Role, error) {
	key := c.buildRoleKey(roleID)
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命�?
		}
		c.logger.Error("Failed to get role cache", zap.String("key", key), zap.Error(err))
		return nil, &CacheError{Operation: "get", Key: key, Err: err}
	}

	var role permission.Role
	err = c.deserialize([]byte(data), &role)
	if err != nil {
		c.logger.Error("Failed to deserialize role", zap.String("key", key), zap.Error(err))
		return nil, &CacheError{Operation: "deserialize", Key: key, Err: err}
	}

	c.logger.Debug("Role cache hit", zap.String("role_id", roleID), zap.String("key", key))
	return &role, nil
}

// DeleteRole 删除角色缓存
func (c *RedisPermissionCache) DeleteRole(ctx context.Context, roleID string) error {
	key := c.buildRoleKey(roleID)
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		c.logger.Error("Failed to delete role cache", zap.String("key", key), zap.Error(err))
		return &CacheError{Operation: "delete", Key: key, Err: err}
	}

	c.logger.Debug("Role cache deleted", zap.String("role_id", roleID), zap.String("key", key))
	return nil
}

// SetPermission 设置权限缓存
func (c *RedisPermissionCache) SetPermission(ctx context.Context, perm *permission.Permission) error {
	key := c.buildPermissionKey(perm.ID)
	data, err := c.serialize(perm)
	if err != nil {
		return &CacheError{Operation: "serialize", Key: key, Err: err}
	}

	err = c.client.Set(ctx, key, data, c.config.PermissionTTL).Err()
	if err != nil {
		c.logger.Error("Failed to set permission cache", zap.String("key", key), zap.Error(err))
		return &CacheError{Operation: "set", Key: key, Err: err}
	}

	c.logger.Debug("Permission cached", zap.String("permission_id", perm.ID), zap.String("key", key))
	return nil
}

// GetPermission 获取权限缓存
func (c *RedisPermissionCache) GetPermission(ctx context.Context, permissionID string) (*permission.Permission, error) {
	key := c.buildPermissionKey(permissionID)
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命�?
		}
		c.logger.Error("Failed to get permission cache", zap.String("key", key), zap.Error(err))
		return nil, &CacheError{Operation: "get", Key: key, Err: err}
	}

	var perm permission.Permission
	err = c.deserialize([]byte(data), &perm)
	if err != nil {
		c.logger.Error("Failed to deserialize permission", zap.String("key", key), zap.Error(err))
		return nil, &CacheError{Operation: "deserialize", Key: key, Err: err}
	}

	c.logger.Debug("Permission cache hit", zap.String("permission_id", permissionID), zap.String("key", key))
	return &perm, nil
}

// DeletePermission 删除权限缓存
func (c *RedisPermissionCache) DeletePermission(ctx context.Context, permissionID string) error {
	key := c.buildPermissionKey(permissionID)
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		c.logger.Error("Failed to delete permission cache", zap.String("key", key), zap.Error(err))
		return &CacheError{Operation: "delete", Key: key, Err: err}
	}

	c.logger.Debug("Permission cache deleted", zap.String("permission_id", permissionID), zap.String("key", key))
	return nil
}

// SetUserRoles 设置用户角色缓存
func (c *RedisPermissionCache) SetUserRoles(ctx context.Context, userID string, tenantID string, roles []*permission.Role) error {
	key := c.buildUserRolesKey(userID, tenantID)
	data, err := c.serialize(roles)
	if err != nil {
		return &CacheError{Operation: "serialize", Key: key, Err: err}
	}

	err = c.client.Set(ctx, key, data, c.config.UserRolesTTL).Err()
	if err != nil {
		c.logger.Error("Failed to set user roles cache", zap.String("key", key), zap.Error(err))
		return &CacheError{Operation: "set", Key: key, Err: err}
	}

	c.logger.Debug("User roles cached", zap.String("user_id", userID), zap.String("tenant_id", tenantID), zap.String("key", key))
	return nil
}

// GetUserRoles 获取用户角色缓存
func (c *RedisPermissionCache) GetUserRoles(ctx context.Context, userID string, tenantID string) ([]*permission.Role, error) {
	key := c.buildUserRolesKey(userID, tenantID)
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命�?
		}
		c.logger.Error("Failed to get user roles cache", zap.String("key", key), zap.Error(err))
		return nil, &CacheError{Operation: "get", Key: key, Err: err}
	}

	var roles []*permission.Role
	err = c.deserialize([]byte(data), &roles)
	if err != nil {
		c.logger.Error("Failed to deserialize user roles", zap.String("key", key), zap.Error(err))
		return nil, &CacheError{Operation: "deserialize", Key: key, Err: err}
	}

	c.logger.Debug("User roles cache hit", zap.String("user_id", userID), zap.String("tenant_id", tenantID), zap.String("key", key))
	return roles, nil
}

// DeleteUserRoles 删除用户角色缓存
func (c *RedisPermissionCache) DeleteUserRoles(ctx context.Context, userID string, tenantID string) error {
	key := c.buildUserRolesKey(userID, tenantID)
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		c.logger.Error("Failed to delete user roles cache", zap.String("key", key), zap.Error(err))
		return &CacheError{Operation: "delete", Key: key, Err: err}
	}

	c.logger.Debug("User roles cache deleted", zap.String("user_id", userID), zap.String("tenant_id", tenantID), zap.String("key", key))
	return nil
}

// SetRolePermissions 设置角色权限缓存
func (c *RedisPermissionCache) SetRolePermissions(ctx context.Context, roleID string, permissions []*permission.Permission) error {
	key := c.buildRolePermissionsKey(roleID)
	data, err := c.serialize(permissions)
	if err != nil {
		return &CacheError{Operation: "serialize", Key: key, Err: err}
	}

	err = c.client.Set(ctx, key, data, c.config.RolePermsTTL).Err()
	if err != nil {
		c.logger.Error("Failed to set role permissions cache", zap.String("key", key), zap.Error(err))
		return &CacheError{Operation: "set", Key: key, Err: err}
	}

	c.logger.Debug("Role permissions cached", zap.String("role_id", roleID), zap.String("key", key))
	return nil
}

// GetRolePermissions 获取角色权限缓存
func (c *RedisPermissionCache) GetRolePermissions(ctx context.Context, roleID string) ([]*permission.Permission, error) {
	key := c.buildRolePermissionsKey(roleID)
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命�?
		}
		c.logger.Error("Failed to get role permissions cache", zap.String("key", key), zap.Error(err))
		return nil, &CacheError{Operation: "get", Key: key, Err: err}
	}

	var permissions []*permission.Permission
	err = c.deserialize([]byte(data), &permissions)
	if err != nil {
		c.logger.Error("Failed to deserialize role permissions", zap.String("key", key), zap.Error(err))
		return nil, &CacheError{Operation: "deserialize", Key: key, Err: err}
	}

	c.logger.Debug("Role permissions cache hit", zap.String("role_id", roleID), zap.String("key", key))
	return permissions, nil
}

// DeleteRolePermissions 删除角色权限缓存
func (c *RedisPermissionCache) DeleteRolePermissions(ctx context.Context, roleID string) error {
	key := c.buildRolePermissionsKey(roleID)
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		c.logger.Error("Failed to delete role permissions cache", zap.String("key", key), zap.Error(err))
		return &CacheError{Operation: "delete", Key: key, Err: err}
	}

	c.logger.Debug("Role permissions cache deleted", zap.String("role_id", roleID), zap.String("key", key))
	return nil
}

// SetPermissionCheckResult 设置权限检查结果缓�?
func (c *RedisPermissionCache) SetPermissionCheckResult(ctx context.Context, checkKey string, result *permission.PermissionCheckResult) error {
	key := c.buildCheckResultKey(checkKey)
	data, err := c.serialize(result)
	if err != nil {
		return &CacheError{Operation: "serialize", Key: key, Err: err}
	}

	err = c.client.Set(ctx, key, data, c.config.CheckResultTTL).Err()
	if err != nil {
		c.logger.Error("Failed to set permission check result cache", zap.String("key", key), zap.Error(err))
		return &CacheError{Operation: "set", Key: key, Err: err}
	}

	c.logger.Debug("Permission check result cached", zap.String("check_key", checkKey), zap.String("key", key))
	return nil
}

// GetPermissionCheckResult 获取权限检查结果缓�?
func (c *RedisPermissionCache) GetPermissionCheckResult(ctx context.Context, checkKey string) (*permission.PermissionCheckResult, error) {
	key := c.buildCheckResultKey(checkKey)
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命�?
		}
		c.logger.Error("Failed to get permission check result cache", zap.String("key", key), zap.Error(err))
		return nil, &CacheError{Operation: "get", Key: key, Err: err}
	}

	var result permission.PermissionCheckResult
	err = c.deserialize([]byte(data), &result)
	if err != nil {
		c.logger.Error("Failed to deserialize permission check result", zap.String("key", key), zap.Error(err))
		return nil, &CacheError{Operation: "deserialize", Key: key, Err: err}
	}

	c.logger.Debug("Permission check result cache hit", zap.String("check_key", checkKey), zap.String("key", key))
	return &result, nil
}

// DeletePermissionCheckResult 删除权限检查结果缓�?
func (c *RedisPermissionCache) DeletePermissionCheckResult(ctx context.Context, checkKey string) error {
	key := c.buildCheckResultKey(checkKey)
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		c.logger.Error("Failed to delete permission check result cache", zap.String("key", key), zap.Error(err))
		return &CacheError{Operation: "delete", Key: key, Err: err}
	}

	c.logger.Debug("Permission check result cache deleted", zap.String("check_key", checkKey), zap.String("key", key))
	return nil
}

// InvalidateUserCache 失效用户相关缓存
func (c *RedisPermissionCache) InvalidateUserCache(ctx context.Context, userID string, tenantID string) error {
	pattern := c.buildUserCachePattern(userID, tenantID)
	return c.deleteByPattern(ctx, pattern)
}

// InvalidateRoleCache 失效角色相关缓存
func (c *RedisPermissionCache) InvalidateRoleCache(ctx context.Context, roleID string) error {
	// 删除角色缓存
	if err := c.DeleteRole(ctx, roleID); err != nil {
		return err
	}

	// 删除角色权限缓存
	if err := c.DeleteRolePermissions(ctx, roleID); err != nil {
		return err
	}

	// 删除包含该角色的用户角色缓存
	pattern := c.config.KeyPrefix + "user_roles:*"
	return c.deleteByPattern(ctx, pattern)
}

// InvalidatePermissionCache 失效权限相关缓存
func (c *RedisPermissionCache) InvalidatePermissionCache(ctx context.Context, permissionID string) error {
	// 删除权限缓存
	if err := c.DeletePermission(ctx, permissionID); err != nil {
		return err
	}

	// 删除包含该权限的角色权限缓存
	pattern := c.config.KeyPrefix + "role_permissions:*"
	return c.deleteByPattern(ctx, pattern)
}

// Clear 清空所有缓�?
func (c *RedisPermissionCache) Clear(ctx context.Context) error {
	pattern := c.config.KeyPrefix + "*"
	return c.deleteByPattern(ctx, pattern)
}

// HealthCheck 健康检�?
func (c *RedisPermissionCache) HealthCheck(ctx context.Context) error {
	_, err := c.client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}
	return nil
}

// 构建角色缓存�?
func (c *RedisPermissionCache) buildRoleKey(roleID string) string {
	return fmt.Sprintf("%srole:%s", c.config.KeyPrefix, roleID)
}

// 构建权限缓存�?
func (c *RedisPermissionCache) buildPermissionKey(permissionID string) string {
	return fmt.Sprintf("%spermission:%s", c.config.KeyPrefix, permissionID)
}

// 构建用户角色缓存�?
func (c *RedisPermissionCache) buildUserRolesKey(userID, tenantID string) string {
	return fmt.Sprintf("%suser_roles:%s:%s", c.config.KeyPrefix, userID, tenantID)
}

// 构建角色权限缓存�?
func (c *RedisPermissionCache) buildRolePermissionsKey(roleID string) string {
	return fmt.Sprintf("%srole_permissions:%s", c.config.KeyPrefix, roleID)
}

// 构建权限检查结果缓存键
func (c *RedisPermissionCache) buildCheckResultKey(checkKey string) string {
	return fmt.Sprintf("%scheck_result:%s", c.config.KeyPrefix, checkKey)
}

// 构建用户缓存模式
func (c *RedisPermissionCache) buildUserCachePattern(userID, tenantID string) string {
	return fmt.Sprintf("%suser_*:%s:*", c.config.KeyPrefix, userID)
}

// 序列化数�?
func (c *RedisPermissionCache) serialize(data interface{}) ([]byte, error) {
	switch c.config.SerializationFormat {
	case "json":
		return json.Marshal(data)
	default:
		return json.Marshal(data)
	}
}

// 反序列化数据
func (c *RedisPermissionCache) deserialize(data []byte, target interface{}) error {
	switch c.config.SerializationFormat {
	case "json":
		return json.Unmarshal(data, target)
	default:
		return json.Unmarshal(data, target)
	}
}

// 根据模式删除缓存
func (c *RedisPermissionCache) deleteByPattern(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		
		// 批量删除
		if len(keys) >= c.config.BatchSize {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				c.logger.Error("Failed to delete keys by pattern", zap.String("pattern", pattern), zap.Error(err))
				return err
			}
			keys = keys[:0] // 清空切片
		}
	}

	// 删除剩余的键
	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			c.logger.Error("Failed to delete remaining keys by pattern", zap.String("pattern", pattern), zap.Error(err))
			return err
		}
	}

	if err := iter.Err(); err != nil {
		c.logger.Error("Failed to scan keys by pattern", zap.String("pattern", pattern), zap.Error(err))
		return err
	}

	c.logger.Debug("Cache cleared by pattern", zap.String("pattern", pattern))
	return nil
}

// GetCacheStats 获取缓存统计信息
func (c *RedisPermissionCache) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	info, err := c.client.Info(ctx, "memory").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get redis info: %w", err)
	}

	stats := make(map[string]interface{})
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				stats[parts[0]] = parts[1]
			}
		}
	}

	// 添加自定义统计信�?
	keyCount, err := c.client.DBSize(ctx).Result()
	if err == nil {
		stats["total_keys"] = keyCount
	}

	// 统计权限相关的键数量
	permKeyCount := int64(0)
	iter := c.client.Scan(ctx, 0, c.config.KeyPrefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		permKeyCount++
	}
	stats["permission_keys"] = permKeyCount

	return stats, nil
}
