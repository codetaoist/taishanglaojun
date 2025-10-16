package middleware

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CacheItem 缓存项
type CacheItem struct {
	Data      []byte    `json:"data"`
	Headers   map[string]string `json:"headers"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// CacheManager 缓存管理器
type CacheManager struct {
	cache  map[string]*CacheItem
	mutex  sync.RWMutex
	logger *zap.Logger
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(logger *zap.Logger) *CacheManager {
	cm := &CacheManager{
		cache:  make(map[string]*CacheItem),
		logger: logger,
	}
	
	// 启动清理协程
	go cm.cleanupExpiredItems()
	
	return cm
}

// CacheConfig 缓存配置
type CacheConfig struct {
	TTL        time.Duration // 缓存时间
	KeyPrefix  string        // 键前缀
	SkipAuth   bool          // 是否跳过认证信息
	VaryBy     []string      // 变化依据（如用户ID、角色等）
}

// CacheMiddleware 缓存中间件
func (cm *CacheManager) CacheMiddleware(config CacheConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只缓存GET请求
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}
		
		// 生成缓存键
		cacheKey := cm.generateCacheKey(c, config)
		
		// 尝试从缓存获取
		if item := cm.get(cacheKey); item != nil {
			cm.logger.Debug("Cache hit", zap.String("key", cacheKey))
			
			// 设置响应头
			for key, value := range item.Headers {
				c.Header(key, value)
			}
			
			// 添加缓存标识头
			c.Header("X-Cache", "HIT")
			c.Header("X-Cache-Created", item.CreatedAt.Format(time.RFC3339))
			c.Header("X-Cache-Expires", item.ExpiresAt.Format(time.RFC3339))
			
			// 返回缓存数据
			c.Data(http.StatusOK, "application/json", item.Data)
			c.Abort()
			return
		}
		
		// 缓存未命中，继续处理请求
		cm.logger.Debug("Cache miss", zap.String("key", cacheKey))
		
		// 创建响应写入器来捕获响应
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:          &bytes.Buffer{},
		}
		c.Writer = writer
		
		c.Next()
		
		// 如果响应成功，缓存结果
		if c.Writer.Status() == http.StatusOK && writer.body.Len() > 0 {
			// 收集响应头
			headers := make(map[string]string)
			for key, values := range c.Writer.Header() {
				if len(values) > 0 {
					headers[key] = values[0]
				}
			}
			
			// 添加缓存标识头
			headers["X-Cache"] = "MISS"
			
			// 存储到缓存
			item := &CacheItem{
				Data:      writer.body.Bytes(),
				Headers:   headers,
				ExpiresAt: time.Now().Add(config.TTL),
				CreatedAt: time.Now(),
			}
			
			cm.set(cacheKey, item)
			cm.logger.Debug("Response cached", 
				zap.String("key", cacheKey),
				zap.Int("size", len(item.Data)),
				zap.Duration("ttl", config.TTL),
			)
		}
	}
}

// responseWriter 响应写入器
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

// generateCacheKey 生成缓存键
func (cm *CacheManager) generateCacheKey(c *gin.Context, config CacheConfig) string {
	var keyParts []string
	
	// 添加前缀
	if config.KeyPrefix != "" {
		keyParts = append(keyParts, config.KeyPrefix)
	}
	
	// 添加路径
	keyParts = append(keyParts, c.Request.URL.Path)
	
	// 添加查询参数
	if c.Request.URL.RawQuery != "" {
		keyParts = append(keyParts, c.Request.URL.RawQuery)
	}
	
	// 添加变化依据
	for _, vary := range config.VaryBy {
		if value := c.GetString(vary); value != "" {
			keyParts = append(keyParts, fmt.Sprintf("%s=%s", vary, value))
		}
	}
	
	// 如果不跳过认证，添加用户信息
	if !config.SkipAuth {
		if userID := c.GetString("user_id"); userID != "" {
			keyParts = append(keyParts, fmt.Sprintf("user=%s", userID))
		}
		if role := c.GetString("role"); role != "" {
			keyParts = append(keyParts, fmt.Sprintf("role=%s", role))
		}
	}
	
	// 生成MD5哈希
	key := strings.Join(keyParts, "|")
	hash := md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])
}

// get 从缓存获取
func (cm *CacheManager) get(key string) *CacheItem {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	item, exists := cm.cache[key]
	if !exists {
		return nil
	}
	
	// 检查是否过期
	if time.Now().After(item.ExpiresAt) {
		delete(cm.cache, key)
		return nil
	}
	
	return item
}

// set 设置缓存
func (cm *CacheManager) set(key string, item *CacheItem) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	cm.cache[key] = item
}

// delete 删除缓存
func (cm *CacheManager) Delete(key string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	delete(cm.cache, key)
}

// clear 清空缓存
func (cm *CacheManager) Clear() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	cm.cache = make(map[string]*CacheItem)
}

// cleanupExpiredItems 清理过期项
func (cm *CacheManager) cleanupExpiredItems() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		cm.mutex.Lock()
		now := time.Now()
		expiredKeys := make([]string, 0)
		
		for key, item := range cm.cache {
			if now.After(item.ExpiresAt) {
				expiredKeys = append(expiredKeys, key)
			}
		}
		
		for _, key := range expiredKeys {
			delete(cm.cache, key)
		}
		
		if len(expiredKeys) > 0 {
			cm.logger.Debug("Cleaned up expired cache items", 
				zap.Int("count", len(expiredKeys)),
				zap.Int("remaining", len(cm.cache)),
			)
		}
		
		cm.mutex.Unlock()
	}
}

// GetStats 获取缓存统计
func (cm *CacheManager) GetStats() map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	totalSize := 0
	expiredCount := 0
	now := time.Now()
	
	for _, item := range cm.cache {
		totalSize += len(item.Data)
		if now.After(item.ExpiresAt) {
			expiredCount++
		}
	}
	
	return map[string]interface{}{
		"total_items":    len(cm.cache),
		"expired_items":  expiredCount,
		"total_size":     totalSize,
		"average_size":   totalSize / max(len(cm.cache), 1),
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}