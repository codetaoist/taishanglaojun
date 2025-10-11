package cache

import (
	"sync"
	"time"
)

// MemoryCache 内存缓存实现
type MemoryCache struct {
	items   map[string]*cacheItem
	mutex   sync.RWMutex
	janitor *janitor
}

// cacheItem 缓存�?
type cacheItem struct {
	value      interface{}
	expiration int64
}

// janitor 清理�?
type janitor struct {
	interval time.Duration
	stop     chan bool
}

// NewMemoryCache 创建内存缓存
func NewMemoryCache(cleanupInterval time.Duration) *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]*cacheItem),
	}

	if cleanupInterval > 0 {
		cache.janitor = &janitor{
			interval: cleanupInterval,
			stop:     make(chan bool),
		}
		go cache.janitor.run(cache)
	}

	return cache
}

// Get 获取缓存�?
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	// 检查是否过�?
	if item.expiration > 0 && time.Now().UnixNano() > item.expiration {
		return nil, false
	}

	return item.value, true
}

// Set 设置缓存�?
func (c *MemoryCache) Set(key string, value interface{}, expiry time.Duration) {
	var expiration int64
	if expiry > 0 {
		expiration = time.Now().Add(expiry).UnixNano()
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items[key] = &cacheItem{
		value:      value,
		expiration: expiration,
	}
}

// Delete 删除缓存�?
func (c *MemoryCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.items, key)
}

// Clear 清空缓存
func (c *MemoryCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.items = make(map[string]*cacheItem)
}

// Size 获取缓存大小
func (c *MemoryCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.items)
}

// Keys 获取所有键
func (c *MemoryCache) Keys() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	keys := make([]string, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}
	return keys
}

// DeleteExpired 删除过期�?
func (c *MemoryCache) DeleteExpired() {
	now := time.Now().UnixNano()
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for key, item := range c.items {
		if item.expiration > 0 && now > item.expiration {
			delete(c.items, key)
		}
	}
}

// Stop 停止清理�?
func (c *MemoryCache) Stop() {
	if c.janitor != nil {
		c.janitor.stop <- true
	}
}

// run 运行清理�?
func (j *janitor) run(cache *MemoryCache) {
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cache.DeleteExpired()
		case <-j.stop:
			return
		}
	}
}
