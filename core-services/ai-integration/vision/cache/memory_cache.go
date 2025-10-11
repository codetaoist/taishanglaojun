package cache

import (
	"sync"
	"time"
)

// MemoryCache 内存缓存实现
type MemoryCache struct {
	data   map[string]*cacheItem
	mutex  sync.RWMutex
	ticker *time.Ticker
	done   chan bool
}

// cacheItem 缓存�?
type cacheItem struct {
	value      interface{}
	expiration time.Time
}

// NewMemoryCache 创建内存缓存
func NewMemoryCache(cleanupInterval time.Duration) *MemoryCache {
	cache := &MemoryCache{
		data: make(map[string]*cacheItem),
		done: make(chan bool),
	}

	if cleanupInterval > 0 {
		cache.ticker = time.NewTicker(cleanupInterval)
		go cache.cleanup()
	}

	return cache
}

// Get 获取缓存�?
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return nil, false
	}

	// 检查是否过�?
	if time.Now().After(item.expiration) {
		// 异步删除过期�?
		go func() {
			c.mutex.Lock()
			delete(c.data, key)
			c.mutex.Unlock()
		}()
		return nil, false
	}

	return item.value, true
}

// Set 设置缓存�?
func (c *MemoryCache) Set(key string, value interface{}, expiry time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	expiration := time.Now().Add(expiry)
	c.data[key] = &cacheItem{
		value:      value,
		expiration: expiration,
	}
}

// Delete 删除缓存�?
func (c *MemoryCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.data, key)
}

// Clear 清空缓存
func (c *MemoryCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data = make(map[string]*cacheItem)
}

// Size 获取缓存大小
func (c *MemoryCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.data)
}

// Keys 获取所有键
func (c *MemoryCache) Keys() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	keys := make([]string, 0, len(c.data))
	for key := range c.data {
		keys = append(keys, key)
	}
	return keys
}

// Close 关闭缓存
func (c *MemoryCache) Close() {
	if c.ticker != nil {
		c.ticker.Stop()
	}
	close(c.done)
}

// cleanup 清理过期�?
func (c *MemoryCache) cleanup() {
	for {
		select {
		case <-c.ticker.C:
			c.removeExpired()
		case <-c.done:
			return
		}
	}
}

// removeExpired 移除过期�?
func (c *MemoryCache) removeExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, item := range c.data {
		if now.After(item.expiration) {
			delete(c.data, key)
		}
	}
}

// Stats 获取缓存统计信息
func (c *MemoryCache) Stats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	now := time.Now()
	expired := 0
	for _, item := range c.data {
		if now.After(item.expiration) {
			expired++
		}
	}

	return CacheStats{
		TotalItems:   len(c.data),
		ExpiredItems: expired,
		ActiveItems:  len(c.data) - expired,
	}
}

// CacheStats 缓存统计信息
type CacheStats struct {
	TotalItems   int `json:"total_items"`
	ExpiredItems int `json:"expired_items"`
	ActiveItems  int `json:"active_items"`
}
