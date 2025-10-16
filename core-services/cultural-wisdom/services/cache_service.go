package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
)

// CacheService 
type CacheService struct {
	redis         *redis.Client
	logger        *zap.Logger
	searchCacheTTL time.Duration
}

// NewCacheService 
func NewCacheService(redisClient *redis.Client, logger *zap.Logger) *CacheService {
	return &CacheService{
		redis:         redisClient,
		logger:        logger,
		searchCacheTTL: SearchCacheTTL,
	}
}

// 
const (
	WisdomCachePrefix     = "wisdom:"
	WisdomListCachePrefix = "wisdom_list:"
	CategoryCachePrefix   = "category:"
	StatsCachePrefix      = "stats:"
	SearchCachePrefix     = "search:"
	// 
	searchResultsPrefix = "search_results:"
	searchSuggestionsPrefix = "search_suggestions:"
	popularSearchesPrefix = "popular_searches"
)

// 
const (
	WisdomCacheTTL     = 30 * time.Minute
	WisdomListCacheTTL = 10 * time.Minute
	CategoryCacheTTL   = 1 * time.Hour
	StatsCacheTTL      = 15 * time.Minute
	SearchCacheTTL     = 5 * time.Minute
)

// GetWisdom ?
func (c *CacheService) GetWisdom(ctx context.Context, id string) (*models.CulturalWisdom, error) {
	// Redisnil
	if c.redis == nil {
		return nil, fmt.Errorf("redis client not available")
	}

	key := WisdomCachePrefix + id
	data, err := c.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // ?
		}
		c.logger.Error("Failed to get wisdom from cache", zap.Error(err), zap.String("key", key))
		return nil, err
	}

	var wisdom models.CulturalWisdom
	if err := json.Unmarshal([]byte(data), &wisdom); err != nil {
		c.logger.Error("Failed to unmarshal wisdom from cache", zap.Error(err))
		return nil, err
	}

	return &wisdom, nil
}

// SetWisdom ?
func (c *CacheService) SetWisdom(ctx context.Context, wisdom *models.CulturalWisdom) error {
	// Redisnil
	if c.redis == nil {
		return nil
	}
	
	key := WisdomCachePrefix + wisdom.ID
	
	data, err := json.Marshal(wisdom)
	if err != nil {
		c.logger.Error("Failed to marshal wisdom for cache", zap.String("id", wisdom.ID), zap.Error(err))
		return err
	}

	if err := c.redis.Set(ctx, key, data, WisdomCacheTTL).Err(); err != nil {
		c.logger.Error("Failed to set wisdom to cache", zap.String("id", wisdom.ID), zap.Error(err))
		return err
	}

	return nil
}

// GetWisdomList ?
func (c *CacheService) GetWisdomList(ctx context.Context, filter *models.WisdomFilter) ([]models.WisdomSummary, int64, error) {
	// Redisnil
	if c.redis == nil {
		return nil, 0, fmt.Errorf("redis client not available")
	}

	key := c.generateWisdomListKey(filter)
	data, err := c.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, 0, nil // ?
		}
		c.logger.Error("Failed to get wisdom list from cache", zap.Error(err), zap.String("key", key))
		return nil, 0, err
	}

	var result struct {
		Wisdoms []models.WisdomSummary `json:"wisdoms"`
		Total   int64                  `json:"total"`
	}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		c.logger.Error("Failed to unmarshal wisdom list from cache", zap.Error(err))
		return nil, 0, err
	}

	return result.Wisdoms, result.Total, nil
}

// SetWisdomList ?
func (c *CacheService) SetWisdomList(ctx context.Context, filter *models.WisdomFilter, wisdoms []models.WisdomSummary, total int64) error {
	// Redisnil
	if c.redis == nil {
		return nil
	}
	
	key := c.generateWisdomListKey(filter)
	
	listData, err := json.Marshal(wisdoms)
	if err != nil {
		c.logger.Error("Failed to marshal wisdom list for cache", zap.Error(err))
		return err
	}

	totalData, err := json.Marshal(total)
	if err != nil {
		c.logger.Error("Failed to marshal wisdom list total for cache", zap.Error(err))
		return err
	}

	// 
	pipe := c.redis.Pipeline()
	pipe.Set(ctx, key+":list", listData, WisdomListCacheTTL)
	pipe.Set(ctx, key+":total", totalData, WisdomListCacheTTL)

	if _, err := pipe.Exec(ctx); err != nil {
		c.logger.Error("Failed to set wisdom list to cache", zap.Error(err))
		return err
	}

	return nil
}

// GetCategories ?
func (c *CacheService) GetCategories(ctx context.Context) ([]models.Category, error) {
	// Redisnil
	if c.redis == nil {
		return nil, nil
	}
	
	key := CategoryCachePrefix + "all"
	
	data, err := c.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // ?
		}
		c.logger.Error("Failed to get categories from cache", zap.Error(err))
		return nil, err
	}

	var categories []models.Category
	if err := json.Unmarshal([]byte(data), &categories); err != nil {
		c.logger.Error("Failed to unmarshal categories from cache", zap.Error(err))
		return nil, err
	}

	return categories, nil
}

// SetCategories ?
func (c *CacheService) SetCategories(ctx context.Context, categories []models.Category) error {
	// Redisnil
	if c.redis == nil {
		return nil
	}
	
	key := CategoryCachePrefix + "all"
	
	data, err := json.Marshal(categories)
	if err != nil {
		c.logger.Error("Failed to marshal categories for cache", zap.Error(err))
		return err
	}

	if err := c.redis.Set(ctx, key, data, CategoryCacheTTL).Err(); err != nil {
		c.logger.Error("Failed to set categories to cache", zap.Error(err))
		return err
	}

	return nil
}

// GetStats ?
func (c *CacheService) GetStats(ctx context.Context) (*models.WisdomStats, error) {
	// Redisnil
	if c.redis == nil {
		return nil, nil
	}
	
	key := StatsCachePrefix + "all"
	
	data, err := c.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // ?
		}
		c.logger.Error("Failed to get stats from cache", zap.Error(err))
		return nil, err
	}

	var stats models.WisdomStats
	if err := json.Unmarshal([]byte(data), &stats); err != nil {
		c.logger.Error("Failed to unmarshal stats from cache", zap.Error(err))
		return nil, err
	}

	return &stats, nil
}

// SetStats ?
func (c *CacheService) SetStats(ctx context.Context, stats *models.WisdomStats) error {
	// Redisnil
	if c.redis == nil {
		return nil
	}
	
	key := StatsCachePrefix + "all"
	
	data, err := json.Marshal(stats)
	if err != nil {
		c.logger.Error("Failed to marshal stats for cache", zap.Error(err))
		return err
	}

	if err := c.redis.Set(ctx, key, data, StatsCacheTTL).Err(); err != nil {
		c.logger.Error("Failed to set stats to cache", zap.Error(err))
		return err
	}

	return nil
}

// InvalidateWisdom ?
func (c *CacheService) InvalidateWisdom(ctx context.Context, id string) error {
	// Redisnil
	if c.redis == nil {
		return nil
	}
	
	key := WisdomCachePrefix + id
	
	if err := c.redis.Del(ctx, key).Err(); err != nil {
		c.logger.Error("Failed to invalidate wisdom cache", zap.String("id", id), zap.Error(err))
		return err
	}

	return nil
}

// InvalidateWisdomList ?
func (c *CacheService) InvalidateWisdomList(ctx context.Context) error {
	// Redisnil
	if c.redis == nil {
		return nil
	}
	
	pattern := WisdomListCachePrefix + "*"
	
	keys, err := c.redis.Keys(ctx, pattern).Result()
	if err != nil {
		c.logger.Error("Failed to get wisdom list cache keys", zap.Error(err))
		return err
	}

	if len(keys) > 0 {
		if err := c.redis.Del(ctx, keys...).Err(); err != nil {
			c.logger.Error("Failed to invalidate wisdom list cache", zap.Error(err))
			return err
		}
	}

	return nil
}

// InvalidateStats ?
func (c *CacheService) InvalidateStats(ctx context.Context) error {
	// Redisnil
	if c.redis == nil {
		return nil
	}
	
	key := StatsCachePrefix + "all"
	
	if err := c.redis.Del(ctx, key).Err(); err != nil {
		c.logger.Error("Failed to invalidate stats cache", zap.Error(err))
		return err
	}

	return nil
}

// IncrementViewCount 
func (c *CacheService) IncrementViewCount(ctx context.Context, wisdomID string) error {
	// Redisnil
	if c.redis == nil {
		return nil
	}
	
	key := fmt.Sprintf("view_count:%s", wisdomID)
	
	if err := c.redis.Incr(ctx, key).Err(); err != nil {
		c.logger.Error("Failed to increment view count", zap.String("wisdom_id", wisdomID), zap.Error(err))
		return err
	}

	// 
	c.redis.Expire(ctx, key, 24*time.Hour)
	
	return nil
}

// GetViewCount 
func (c *CacheService) GetViewCount(ctx context.Context, wisdomID string) (int64, error) {
	// Redis0
	if c.redis == nil {
		return 0, nil
	}
	
	key := fmt.Sprintf("view_count:%s", wisdomID)
	
	count, err := c.redis.Get(ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // 
		}
		c.logger.Error("Failed to get view count", zap.String("wisdom_id", wisdomID), zap.Error(err))
		return 0, err
	}

	return count, nil
}

// GetSearchResults 
func (c *CacheService) GetSearchResults(ctx context.Context, cacheKey string) ([]*models.Wisdom, error) {
	// Redisnil
	if c.redis == nil {
		return nil, nil
	}
	
	key := searchResultsPrefix + cacheKey
	
	data, err := c.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var results []*models.Wisdom
	if err := json.Unmarshal([]byte(data), &results); err != nil {
		return nil, err
	}

	return results, nil
}

// SetSearchResults 
func (c *CacheService) SetSearchResults(ctx context.Context, cacheKey string, results []*models.Wisdom) error {
	// Redisnil
	if c.redis == nil {
		return nil
	}
	
	key := searchResultsPrefix + cacheKey
	data, err := json.Marshal(results)
	if err != nil {
		return err
	}

	return c.redis.Set(ctx, key, data, c.searchCacheTTL).Err()
}

// GetSearchSuggestions 黺
func (c *CacheService) GetSearchSuggestions(ctx context.Context, query string) ([]string, error) {
	// Redisnil
	if c.redis == nil {
		return nil, nil
	}
	
	key := searchSuggestionsPrefix + query
	
	data, err := c.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var suggestions []string
	if err := json.Unmarshal([]byte(data), &suggestions); err != nil {
		return nil, err
	}

	return suggestions, nil
}

// SetSearchSuggestions 黺
func (c *CacheService) SetSearchSuggestions(ctx context.Context, query string, suggestions []string) error {
	// Redisnil
	if c.redis == nil {
		return nil
	}
	
	key := searchSuggestionsPrefix + query
	data, err := json.Marshal(suggestions)
	if err != nil {
		return err
	}

	return c.redis.Set(ctx, key, data, c.searchCacheTTL).Err()
}

// GetRecommendations 
func (c *CacheService) GetRecommendations(ctx context.Context, cacheKey string) ([]RecommendationItem, error) {
	key := fmt.Sprintf("recommendations:%s", cacheKey)
	data, err := c.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	
	var recommendations []RecommendationItem
	if err := json.Unmarshal([]byte(data), &recommendations); err != nil {
		return nil, err
	}
	
	return recommendations, nil
}

// SetRecommendations 
func (c *CacheService) SetRecommendations(ctx context.Context, cacheKey string, recommendations []RecommendationItem, expiration time.Duration) error {
	key := fmt.Sprintf("recommendations:%s", cacheKey)
	data, err := json.Marshal(recommendations)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, key, data, expiration).Err()
}

// GetPopularSearches 
func (c *CacheService) GetPopularSearches(ctx context.Context) ([]PopularSearch, error) {
	// Redisnil
	if c.redis == nil {
		return nil, nil
	}
	
	data, err := c.redis.Get(ctx, popularSearchesPrefix).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var searches []PopularSearch
	if err := json.Unmarshal([]byte(data), &searches); err != nil {
		return nil, err
	}

	return searches, nil
}

// SetPopularSearches 
func (c *CacheService) SetPopularSearches(ctx context.Context, searches []PopularSearch) error {
	// Redisnil
	if c.redis == nil {
		return nil
	}
	
	data, err := json.Marshal(searches)
	if err != nil {
		return err
	}

	return c.redis.Set(ctx, popularSearchesPrefix, data, c.searchCacheTTL).Err()
}

// DeletePattern 
func (c *CacheService) DeletePattern(ctx context.Context, pattern string) error {
	// Redisnil
	if c.redis == nil {
		return nil
	}
	
	keys, err := c.redis.Keys(ctx, pattern).Result()
	if err != nil {
		c.logger.Error("Failed to get keys for pattern", zap.String("pattern", pattern), zap.Error(err))
		return err
	}

	if len(keys) > 0 {
		if err := c.redis.Del(ctx, keys...).Err(); err != nil {
			c.logger.Error("Failed to delete keys", zap.String("pattern", pattern), zap.Error(err))
			return err
		}
	}

	return nil
}

// PopularSearch 
type PopularSearch struct {
	Query string `json:"query"`
	Count int64  `json:"count"`
	Rank  int    `json:"rank"`
}

// generateWisdomListKey ?
func (c *CacheService) generateWisdomListKey(filter *models.WisdomFilter) string {
	key := WisdomListCachePrefix
	
	if filter.CategoryID != "" {
		key += fmt.Sprintf("cat_%s_", filter.CategoryID)
	}
	if filter.School != "" {
		key += fmt.Sprintf("school_%s_", filter.School)
	}
	if filter.Status != "" {
		key += fmt.Sprintf("status_%s_", filter.Status)
	}
	if filter.Difficulty != nil && len(filter.Difficulty) > 0 {
		key += fmt.Sprintf("diff_%v_", filter.Difficulty)
	}
	if filter.SearchQuery != "" {
		key += fmt.Sprintf("q_%s_", filter.SearchQuery)
	}
	
	key += fmt.Sprintf("page_%d_size_%d_sort_%s_%s", 
		filter.Page, filter.Size, filter.SortBy, filter.SortOrder)
	
	return key
}

