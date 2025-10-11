package services

import (
	"context"
	"fmt"
	"strings"
	"math"
	"sort"
	"sync"
	"time"

	"gorm.io/gorm"
	"go.uber.org/zap"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/services"
)

// SearchService 搜索服务
type SearchService struct {
	db        *gorm.DB
	cache     *CacheService
	aiService *services.AIService
	logger    *zap.Logger
	
	// 向量搜索优化
	vectorCache    map[string][]float32 // 向量缓存
	vectorCacheMu  sync.RWMutex         // 向量缓存�?
	indexedWisdoms []IndexedWisdom      // 预索引的智慧内容
	indexMu        sync.RWMutex         // 索引�?
	lastIndexTime  time.Time            // 最后索引时�?
}

// IndexedWisdom 预索引的智慧内容
type IndexedWisdom struct {
	ID       string
	Title    string
	Content  string
	Author   string
	School   string
	Category string
	Tags     string
	Vector   []float32
	Score    float32 // 内容质量评分
}

// NewSearchService 创建搜索服务实例
func NewSearchService(db *gorm.DB, cache *CacheService, aiService *services.AIService, logger *zap.Logger) *SearchService {
	service := &SearchService{
		db:          db,
		cache:       cache,
		aiService:   aiService,
		logger:      logger,
		vectorCache: make(map[string][]float32),
	}
	
	// 启动后台索引更新
	go service.backgroundIndexUpdate()
	
	return service
}

// backgroundIndexUpdate 后台索引更新
func (s *SearchService) backgroundIndexUpdate() {
	ticker := time.NewTicker(30 * time.Minute) // �?0分钟更新一次索�?
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if err := s.updateVectorIndex(); err != nil {
				s.logger.Error("Failed to update vector index", zap.Error(err))
			}
		}
	}
}

// updateVectorIndex 更新向量索引
func (s *SearchService) updateVectorIndex() error {
	// 获取所有已发布的智慧内�?
	var wisdoms []models.CulturalWisdom
	if err := s.db.Where("status = ?", "published").Find(&wisdoms).Error; err != nil {
		return fmt.Errorf("failed to fetch wisdoms: %w", err)
	}

	var indexedWisdoms []IndexedWisdom
	var wg sync.WaitGroup
	
	for _, wisdom := range wisdoms {
		wg.Add(1)
		go func(wisdom models.CulturalWisdom) {
			defer wg.Done()
			
			// 生成向量表示
			vector, err := s.aiService.GetEmbedding(context.Background(), wisdom.Content)
			if err != nil {
				s.logger.Error("Failed to generate embedding", zap.Error(err))
				return
			}
			
			// 计算内容质量评分
			score := s.calculateContentScore(wisdom)
			
			// 修复类型错误 - �?StringSlice 转换为字符串
			tagsStr := ""
			if len(wisdom.Tags) > 0 {
				tagsStr = strings.Join(wisdom.Tags, ",")
			}
			
			indexed := IndexedWisdom{
				ID:       wisdom.ID,
				Title:    wisdom.Title,
				Content:  wisdom.Content,
				Author:   wisdom.Author,
				School:   wisdom.School,
				Category: wisdom.Category,
				Tags:     tagsStr,
				Vector:   vector,
				Score:    score,
			}
			
			s.indexMu.Lock()
			indexedWisdoms = append(indexedWisdoms, indexed)
			s.indexMu.Unlock()
		}(wisdom)
	}
	
	wg.Wait()
	
	// 更新索引
	s.indexMu.Lock()
	s.indexedWisdoms = indexedWisdoms
	s.lastIndexTime = time.Now()
	s.indexMu.Unlock()
	
	s.logger.Info("Vector index updated", zap.Int("count", len(indexedWisdoms)))
	return nil
}

// calculateContentScore 计算内容质量评分
func (s *SearchService) calculateContentScore(wisdom models.CulturalWisdom) float32 {
	score := float32(1.0)
	
	// 基于内容长度
	contentLength := len(wisdom.Content)
	if contentLength > 100 {
		score += 0.2
	}
	if contentLength > 500 {
		score += 0.3
	}
	
	// 基于标题质量
	if len(wisdom.Title) > 10 {
		score += 0.1
	}
	
	// 基于作者信�?
	if wisdom.Author != "" {
		score += 0.2
	}
	
	// 基于学派信息
	if wisdom.School != "" {
		score += 0.1
	}
	
	// 基于标签数量 - 修复类型错误
	tagCount := len(wisdom.Tags)
	if tagCount > 2 {
		score += 0.1
	}
	
	return score
}

// SearchWisdom 搜索智慧内容
func (s *SearchService) SearchWisdom(ctx context.Context, query string, limit int, offset int) ([]*models.Wisdom, error) {
	if query == "" {
		return nil, fmt.Errorf("搜索关键词不能为�?)
	}

	// 尝试从缓存获取搜索结�?
	cacheKey := fmt.Sprintf("search:%s:%d:%d", query, limit, offset)
	if s.cache != nil {
		if results, err := s.cache.GetSearchResults(ctx, cacheKey); err == nil && results != nil {
			return results, nil
		}
	}

	var wisdoms []models.CulturalWisdom
	
	// 构建搜索查询
	searchQuery := s.db.WithContext(ctx).Model(&models.CulturalWisdom{})
	
	// 在标题、内容、作者、标签中搜索
	searchQuery = searchQuery.Where(
		"title LIKE ? OR content LIKE ? OR author LIKE ? OR tags LIKE ?",
		"%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%",
	)
	
	// 分页
	if err := searchQuery.Limit(limit).Offset(offset).Find(&wisdoms).Error; err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}

	// 转换为搜索结�?
	results := make([]*models.Wisdom, len(wisdoms))
	for i, wisdom := range wisdoms {
		results[i] = &models.Wisdom{
			ID:       wisdom.ID,
			Title:    wisdom.Title,
			Content:  wisdom.Content,
			Author:   wisdom.Author,
			Source:   wisdom.School, // 使用School字段作为Source
			Category: wisdom.Category,
			Tags:     wisdom.Tags,
		}
	}

	// 缓存搜索结果
	if s.cache != nil {
		s.cache.SetSearchResults(ctx, cacheKey, results)
	}

	return results, nil
}

// SearchRequest 搜索请求结构�?
type SearchRequest struct {
	Query      string   `json:"query"`
	CategoryID string   `json:"category_id"`
	School     string   `json:"school"`
	Tags       []string `json:"tags"`
	Page       int      `json:"page"`
	Size       int      `json:"size"`
}

// SemanticSearchRequest 语义搜索请求结构�?
type SemanticSearchRequest struct {
	Query      string   `json:"query"`
	UserID     string   `json:"user_id"`
	CategoryID string   `json:"category_id"`
	School     string   `json:"school"`
	Tags       []string `json:"tags"`
	Threshold  float32  `json:"threshold"`
	Page       int      `json:"page"`
	Size       int      `json:"size"`
}

// FullTextSearch 全文搜索
func (s *SearchService) FullTextSearch(ctx context.Context, req *SearchRequest) (*models.SearchResult, error) {
	if req.Query == "" {
		return nil, fmt.Errorf("搜索关键词不能为�?)
	}

	var wisdoms []models.CulturalWisdom
	query := s.db.WithContext(ctx).Model(&models.CulturalWisdom{})

	// 构建搜索条件
	searchCondition := "title LIKE ? OR content LIKE ? OR author LIKE ?"
	searchArgs := []interface{}{
		"%" + req.Query + "%",
		"%" + req.Query + "%", 
		"%" + req.Query + "%",
	}

	// 添加分类过滤
	if req.CategoryID != "" {
		query = query.Where("category = ?", req.CategoryID)
	}

	// 添加学派过滤
	if req.School != "" {
		query = query.Where("school = ?", req.School)
	}

	// 添加标签过滤
	if len(req.Tags) > 0 {
		for _, tag := range req.Tags {
			query = query.Where("tags LIKE ?", "%"+tag+"%")
		}
	}

	// 应用搜索条件
	query = query.Where(searchCondition, searchArgs...)

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("计算搜索结果总数失败: %w", err)
	}

	// 分页查询
	offset := (req.Page - 1) * req.Size
	if err := query.Limit(req.Size).Offset(offset).Find(&wisdoms).Error; err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}

	// 转换为Wisdom结构
	wisdomResults := make([]models.Wisdom, len(wisdoms))
	for i, wisdom := range wisdoms {
		wisdomResults[i] = models.Wisdom{
			ID:       wisdom.ID,
			Title:    wisdom.Title,
			Content:  wisdom.Content,
			Author:   wisdom.Author,
			Category: wisdom.Category,
			Tags:     wisdom.Tags,
		}
	}

	return &models.SearchResult{
		Wisdoms:  wisdomResults,
		Total:    int(total),
		Page:     req.Page,
		PageSize: req.Size,
	}, nil
}

// SemanticSearch 语义搜索（需要AI集成�?
func (s *SearchService) SemanticSearch(ctx context.Context, req *SemanticSearchRequest) ([]*models.Wisdom, error) {
	if req.Query == "" {
		return nil, fmt.Errorf("搜索关键词不能为�?)
	}

	// 使用AI服务进行语义搜索
	if s.aiService != nil {
		// 获取查询文本的向量表�?
		queryVector, err := s.aiService.GetEmbedding(ctx, req.Query)
		if err != nil {
			s.logger.Warn("Failed to get query embedding, falling back to keyword search", 
				zap.Error(err), zap.String("query", req.Query))
		} else {
			// 执行向量相似度搜�?
			results, err := s.vectorSearch(ctx, queryVector, req.Size, (req.Page-1)*req.Size)
			if err != nil {
				s.logger.Warn("Vector search failed, falling back to keyword search", 
					zap.Error(err))
			} else {
				return results, nil
			}
		}
	}

	// 回退到关键词搜索
	offset := (req.Page - 1) * req.Size
	return s.SearchWisdom(ctx, req.Query, req.Size, offset)
}

// GetSearchSuggestions 获取搜索建议
func (s *SearchService) GetSearchSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	if query == "" {
		return []string{}, nil
	}

	// 尝试从缓存获�?
	if s.cache != nil {
		if suggestions, err := s.cache.GetSearchSuggestions(ctx, query); err == nil && suggestions != nil {
			return suggestions, nil
		}
	}

	var suggestions []string
	
	// 从数据库获取相关的标题和标签作为建议
	var titles []string
	if err := s.db.WithContext(ctx).
		Model(&models.CulturalWisdom{}).
		Select("DISTINCT title").
		Where("title LIKE ?", "%"+query+"%").
		Limit(limit/2).
		Pluck("title", &titles).Error; err != nil {
		return nil, fmt.Errorf("获取标题建议失败: %w", err)
	}
	
	suggestions = append(suggestions, titles...)
	
	// 获取标签建议
	var tags []string
	if err := s.db.WithContext(ctx).
		Model(&models.CulturalWisdom{}).
		Select("DISTINCT tags").
		Where("tags LIKE ?", "%"+query+"%").
		Limit(limit/2).
		Pluck("tags", &tags).Error; err != nil {
		return nil, fmt.Errorf("获取标签建议失败: %w", err)
	}
	
	// 解析标签字符串并添加到建议中
	for _, tagStr := range tags {
		if tagStr != "" {
			tagList := strings.Split(tagStr, ",")
			for _, tag := range tagList {
				tag = strings.TrimSpace(tag)
				if strings.Contains(strings.ToLower(tag), strings.ToLower(query)) {
					suggestions = append(suggestions, tag)
				}
			}
		}
	}
	
	// 去重并限制数�?
	uniqueSuggestions := make([]string, 0)
	seen := make(map[string]bool)
	for _, suggestion := range suggestions {
		if !seen[suggestion] && len(uniqueSuggestions) < limit {
			uniqueSuggestions = append(uniqueSuggestions, suggestion)
			seen[suggestion] = true
		}
	}

	// 缓存结果
	if s.cache != nil {
		s.cache.SetSearchSuggestions(ctx, query, uniqueSuggestions)
	}

	return uniqueSuggestions, nil
}

// GetPopularSearches 获取热门搜索
func (s *SearchService) GetPopularSearches(ctx context.Context, limit int) ([]PopularSearch, error) {
	// 尝试从缓存获�?
	if searches, err := s.cache.GetPopularSearches(ctx); err == nil && searches != nil {
		if len(searches) > limit {
			searches = searches[:limit]
		}
		return searches, nil
	}

	// 模拟热门搜索数据
	popularSearches := []PopularSearch{
		{Query: "道德�?, Count: 1500},
		{Query: "论语", Count: 1200},
		{Query: "孔子", Count: 1000},
		{Query: "老子", Count: 950},
		{Query: "庄子", Count: 800},
		{Query: "孟子", Count: 750},
		{Query: "易经", Count: 700},
		{Query: "诗经", Count: 650},
		{Query: "春秋", Count: 600},
		{Query: "礼记", Count: 550},
	}

	if len(popularSearches) > limit {
		popularSearches = popularSearches[:limit]
	}

	// 缓存结果
	s.cache.SetPopularSearches(ctx, popularSearches)

	return popularSearches, nil
}

// GetWisdomByID 根据ID获取智慧内容
func (s *SearchService) GetWisdomByID(ctx context.Context, id string) (*models.CulturalWisdom, error) {
	if id == "" {
		return nil, fmt.Errorf("ID不能为空")
	}

	var wisdom models.CulturalWisdom
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&wisdom).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("智慧内容不存�?)
		}
		return nil, fmt.Errorf("查询智慧内容失败: %w", err)
	}

	return &wisdom, nil
}

// GetCategories 获取分类列表
func (s *SearchService) GetCategories(ctx context.Context) ([]models.Category, error) {
	// 尝试从缓存获�?
	if s.cache != nil {
		if categories, err := s.cache.GetCategories(ctx); err == nil && categories != nil {
			return categories, nil
		}
	}

	var categories []models.Category
	if err := s.db.Where("is_active = ?", true).Order("sort_order ASC, name ASC").Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	// 缓存结果
	if s.cache != nil {
		s.cache.SetCategories(ctx, categories)
	}

	return categories, nil
}

// SearchByCategory 按分类搜�?
func (s *SearchService) SearchByCategory(ctx context.Context, category string, limit int, offset int) ([]*models.Wisdom, error) {
	if category == "" {
		return nil, fmt.Errorf("分类不能为空")
	}

	var wisdoms []models.CulturalWisdom
	
	// 按分类查�?
	if err := s.db.WithContext(ctx).
		Where("category = ?", category).
		Limit(limit).
		Offset(offset).
		Find(&wisdoms).Error; err != nil {
		return nil, fmt.Errorf("按分类搜索失�? %w", err)
	}

	// 转换为搜索结�?
	results := make([]*models.Wisdom, len(wisdoms))
	for i, wisdom := range wisdoms {
		results[i] = &models.Wisdom{
			ID:       wisdom.ID,
			Title:    wisdom.Title,
			Content:  wisdom.Content,
			Author:   wisdom.Author,
			Source:   wisdom.School, // 使用School字段作为Source
			Category: wisdom.Category,
			Tags:     wisdom.Tags,
		}
	}

	return results, nil
}

// AdvancedSearchRequest 高级搜索请求结构�?
type AdvancedSearchRequest struct {
	Query       string    `json:"query"`
	CategoryIDs []int     `json:"category_ids"`
	TagIDs      []int     `json:"tag_ids"`
	Schools     []string  `json:"schools"`
	Authors     []string  `json:"authors"`
	DateFrom    *string   `json:"date_from"`
	DateTo      *string   `json:"date_to"`
	SortBy      string    `json:"sort_by"`      // created_at, view_count, like_count, relevance
	SortOrder   string    `json:"sort_order"`   // asc, desc
	Page        int       `json:"page"`
	Size        int       `json:"size"`
}

// SearchFilters 搜索过滤�?
type SearchFilters struct {
	Categories []models.Category `json:"categories"`
	Tags       []models.WisdomTag `json:"tags"`
	Schools    []string          `json:"schools"`
	Authors    []string          `json:"authors"`
}

// AdvancedSearch 高级搜索
func (s *SearchService) AdvancedSearch(ctx context.Context, req *AdvancedSearchRequest) (*models.SearchResult, error) {
	var wisdoms []models.CulturalWisdom
	query := s.db.WithContext(ctx).Model(&models.CulturalWisdom{})

	// 基础文本搜索
	if req.Query != "" {
		searchCondition := "title ILIKE ? OR content ILIKE ? OR author ILIKE ?"
		searchArgs := []interface{}{
			"%" + req.Query + "%",
			"%" + req.Query + "%", 
			"%" + req.Query + "%",
		}
		query = query.Where(searchCondition, searchArgs...)
	}

	// 分类过滤
	if len(req.CategoryIDs) > 0 {
		query = query.Joins("JOIN wisdom_categories wc ON cultural_wisdom.category_id = wc.id").
			Where("wc.id IN ?", req.CategoryIDs)
	}

	// 标签过滤
	if len(req.TagIDs) > 0 {
		query = query.Joins("JOIN wisdom_tag_relations wtr ON cultural_wisdom.id = wtr.wisdom_id").
			Where("wtr.tag_id IN ?", req.TagIDs)
	}

	// 学派过滤
	if len(req.Schools) > 0 {
		query = query.Joins("JOIN wisdom_schools ws ON cultural_wisdom.school_id = ws.id").
			Where("ws.name IN ?", req.Schools)
	}

	// 作者过�?
	if len(req.Authors) > 0 {
		query = query.Where("author IN ?", req.Authors)
	}

	// 日期范围过滤
	if req.DateFrom != nil {
		query = query.Where("created_at >= ?", *req.DateFrom)
	}
	if req.DateTo != nil {
		query = query.Where("created_at <= ?", *req.DateTo)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("计算搜索结果总数失败: %w", err)
	}

	// 排序
	orderClause := "created_at DESC"
	switch req.SortBy {
	case "view_count":
		orderClause = "view_count " + req.SortOrder
	case "like_count":
		orderClause = "like_count " + req.SortOrder
	case "created_at":
		orderClause = "created_at " + req.SortOrder
	case "relevance":
		orderClause = "relevance " + req.SortOrder
	}

	// 分页查询
	offset := (req.Page - 1) * req.Size
	if err := query.Order(orderClause).Limit(req.Size).Offset(offset).Find(&wisdoms).Error; err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}

	// 转换为Wisdom结构
	wisdomResults := make([]models.Wisdom, len(wisdoms))
	for i, wisdom := range wisdoms {
		wisdomResults[i] = models.Wisdom{
			ID:       wisdom.ID,
			Title:    wisdom.Title,
			Content:  wisdom.Content,
			Author:   wisdom.Author,
			Category: wisdom.Category,
			Tags:     wisdom.Tags,
		}
	}

	return &models.SearchResult{
		Wisdoms:  wisdomResults,
		Total:    int(total),
		Page:     req.Page,
		PageSize: req.Size,
	}, nil
}

// GetSearchFilters 获取搜索过滤器选项
func (s *SearchService) GetSearchFilters(ctx context.Context) (*SearchFilters, error) {
	filters := &SearchFilters{}

	// 获取分类
	var categories []models.Category
	if err := s.db.WithContext(ctx).Where("is_active = ?", true).Order("sort_order ASC").Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("获取分类失败: %w", err)
	}
	filters.Categories = categories

	// 获取标签
	var tags []models.WisdomTag
	if err := s.db.WithContext(ctx).Where("is_active = ? AND usage_count > 0", true).
		Order("usage_count DESC").Limit(50).Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("获取标签失败: %w", err)
	}
	filters.Tags = tags

	// 获取学派
	var schools []string
	if err := s.db.WithContext(ctx).Model(&models.WisdomSchool{}).
		Where("is_active = ?", true).Order("name ASC").Pluck("name", &schools).Error; err != nil {
		return nil, fmt.Errorf("获取学派失败: %w", err)
	}
	filters.Schools = schools

	// 获取作�?
	var authors []string
	if err := s.db.WithContext(ctx).Model(&models.CulturalWisdom{}).
		Select("DISTINCT author").Where("author != ''").
		Order("author ASC").Limit(100).Pluck("author", &authors).Error; err != nil {
		return nil, fmt.Errorf("获取作者失�? %w", err)
	}
	filters.Authors = authors

	return filters, nil
}

// SearchWithFacets 带分面的搜索
func (s *SearchService) SearchWithFacets(ctx context.Context, req *SearchRequest) (*models.SearchResultWithFacets, error) {
	// 执行基础搜索
	searchResult, err := s.FullTextSearch(ctx, req)
	if err != nil {
		return nil, err
	}

	// 获取分面统计
	facets, err := s.getSearchFacets(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("获取分面统计失败: %w", err)
	}

	return &models.SearchResultWithFacets{
		SearchResult: *searchResult,
		Facets:       facets,
	}, nil
}

// getSearchFacets 获取搜索分面统计
func (s *SearchService) getSearchFacets(ctx context.Context, req *SearchRequest) (map[string]interface{}, error) {
	facets := make(map[string]interface{})

	// 构建基础查询
	baseQuery := s.db.WithContext(ctx).Model(&models.CulturalWisdom{})
	if req.Query != "" {
		searchCondition := "title ILIKE ? OR content ILIKE ? OR author ILIKE ?"
		searchArgs := []interface{}{
			"%" + req.Query + "%",
			"%" + req.Query + "%", 
			"%" + req.Query + "%",
		}
		baseQuery = baseQuery.Where(searchCondition, searchArgs...)
	}

	// 分类分面
	var categoryFacets []struct {
		CategoryID int    `json:"category_id"`
		Name       string `json:"name"`
		Count      int64  `json:"count"`
	}
	if err := baseQuery.Select("wc.id as category_id, wc.name, COUNT(*) as count").
		Joins("JOIN wisdom_categories wc ON cultural_wisdom.category_id = wc.id").
		Group("wc.id, wc.name").
		Order("count DESC").
		Scan(&categoryFacets).Error; err != nil {
		return nil, fmt.Errorf("获取分类分面失败: %w", err)
	}
	facets["categories"] = categoryFacets

	// 标签分面
	var tagFacets []struct {
		TagID int    `json:"tag_id"`
		Name  string `json:"name"`
		Count int64  `json:"count"`
	}
	if err := baseQuery.Select("wt.id as tag_id, wt.name, COUNT(*) as count").
		Joins("JOIN wisdom_tag_relations wtr ON cultural_wisdom.id = wtr.wisdom_id").
		Joins("JOIN wisdom_tags wt ON wtr.tag_id = wt.id").
		Group("wt.id, wt.name").
		Order("count DESC").
		Limit(20).
		Scan(&tagFacets).Error; err != nil {
		return nil, fmt.Errorf("获取标签分面失败: %w", err)
	}
	facets["tags"] = tagFacets

	// 作者分�?
	var authorFacets []struct {
		Author string `json:"author"`
		Count  int64  `json:"count"`
	}
	if err := baseQuery.Select("author, COUNT(*) as count").
		Where("author != ''").
		Group("author").
		Order("count DESC").
		Limit(10).
		Scan(&authorFacets).Error; err != nil {
		return nil, fmt.Errorf("获取作者分面失�? %w", err)
	}
	facets["authors"] = authorFacets

	return facets, nil
}

// vectorSearch 执行向量相似度搜�?
func (s *SearchService) vectorSearch(ctx context.Context, queryVector []float32, limit int, offset int) ([]*models.Wisdom, error) {
	// 获取所有智慧内�?
	var wisdoms []models.CulturalWisdom
	if err := s.db.WithContext(ctx).Find(&wisdoms).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch wisdom content: %w", err)
	}

	// 计算相似度并排序
	type wisdomScore struct {
		wisdom *models.CulturalWisdom
		score  float32
	}

	var scores []wisdomScore
	for _, wisdom := range wisdoms {
		// 如果智慧内容没有向量表示，则生成一�?
		var contentVector []float32
		if wisdom.Vector != nil && len(wisdom.Vector) > 0 {
			contentVector = wisdom.Vector
		} else {
			// 为内容生成向量表�?
			vector, err := s.aiService.GetEmbedding(ctx, wisdom.Content)
			if err != nil {
				s.logger.Warn("Failed to generate embedding for wisdom content", 
					zap.Error(err), zap.String("wisdom_id", wisdom.ID))
				continue
			}
			contentVector = vector
			
			// 保存向量到数据库（异步）
			go func(w models.CulturalWisdom, v []float32) {
				w.Vector = v
				s.db.Save(&w)
			}(wisdom, vector)
		}

		// 计算余弦相似�?
		similarity := s.cosineSimilarity(queryVector, contentVector)
		scores = append(scores, wisdomScore{
			wisdom: &wisdom,
			score:  similarity,
		})
	}

	// 按相似度排序
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[i].score < scores[j].score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	// 分页处理
	start := offset
	end := offset + limit
	if start >= len(scores) {
		return []*models.Wisdom{}, nil
	}
	if end > len(scores) {
		end = len(scores)
	}

	// 转换结果
	var results []*models.Wisdom
	for i := start; i < end; i++ {
		wisdom := scores[i].wisdom
		// 修复类型错误 - �?StringSlice 转换�?[]string
		tagsSlice := []string(wisdom.Tags)
		results = append(results, &models.Wisdom{
			ID:          wisdom.ID,
			Title:       wisdom.Title,
			Content:     wisdom.Content,
			Author:      wisdom.Author,
			Source:      wisdom.School, // 使用School字段作为Source
			Category:    wisdom.Category,
			Tags:        tagsSlice,
		})
	}

	return results, nil
}

// cosineSimilarity 计算两个向量的余弦相似度
func (s *SearchService) cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float32
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

// EnhancedSemanticSearch 增强的语义搜�?
func (s *SearchService) EnhancedSemanticSearch(ctx context.Context, req *SemanticSearchRequest) (*EnhancedSearchResult, error) {
	if req.Query == "" {
		return nil, fmt.Errorf("搜索关键词不能为�?)
	}

	// 获取查询向量
	queryVector, err := s.getOrGenerateVector(ctx, req.Query)
	if err != nil {
		s.logger.Warn("Failed to get query embedding, falling back to keyword search", 
			zap.Error(err), zap.String("query", req.Query))
		// 回退到关键词搜索
		results, err := s.SearchWisdom(ctx, req.Query, req.Size, (req.Page-1)*req.Size)
		if err != nil {
			return nil, err
		}
		return &EnhancedSearchResult{
			Results:    results,
			Total:      len(results),
			SearchType: "keyword",
			QueryTime:  time.Since(time.Now()),
		}, nil
	}

	startTime := time.Now()
	
	// 使用预索引进行快速搜�?
	results, err := s.fastVectorSearch(ctx, queryVector, req)
	if err != nil {
		return nil, err
	}
	
	queryTime := time.Since(startTime)
	
	return &EnhancedSearchResult{
		Results:     results.Results,
		Total:       results.Total,
		SearchType:  "semantic",
		QueryTime:   queryTime,
		Suggestions: results.Suggestions,
		Facets:      results.Facets,
	}, nil
}

// scoredResult 用于搜索结果评分
type scoredResult struct {
	wisdom     IndexedWisdom
	similarity float32
	relevance  float32 // 综合相关性评�?
}

// FacetItem represents a facet item with count
type FacetItem struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// fastVectorSearch 快速向量搜�?
func (s *SearchService) fastVectorSearch(ctx context.Context, queryVector []float32, req *SemanticSearchRequest) (*DetailedSearchResult, error) {
	// 获取预索引的智慧内容
	s.indexMu.RLock()
	indexedWisdoms := make([]IndexedWisdom, len(s.indexedWisdoms))
	copy(indexedWisdoms, s.indexedWisdoms)
	s.indexMu.RUnlock()
	
	// 如果索引为空，尝试更�?
	if len(indexedWisdoms) == 0 {
		if err := s.updateVectorIndex(); err != nil {
			return nil, fmt.Errorf("failed to update vector index: %w", err)
		}
		
		s.indexMu.RLock()
		indexedWisdoms = s.indexedWisdoms
		s.indexMu.RUnlock()
	}
	
	var scoredResults []scoredResult
	
	for _, wisdom := range indexedWisdoms {
		// 应用过滤条件
		if !s.matchesFilters(wisdom, req) {
			continue
		}
		
		similarity := s.cosineSimilarity(queryVector, wisdom.Vector)
		
		// 如果设置了阈值，过滤低相似度结果
		if req.Threshold > 0 && similarity < req.Threshold {
			continue
		}
		
		// 计算综合相关性评�?
		relevance := s.calculateRelevanceScore(similarity, wisdom, req.Query)
		
		scoredResults = append(scoredResults, scoredResult{
			wisdom:     wisdom,
			similarity: similarity,
			relevance:  relevance,
		})
	}
	
	// 按相关性排�?
	sort.Slice(scoredResults, func(i, j int) bool {
		return scoredResults[i].relevance > scoredResults[j].relevance
	})
	
	// 分页处理
	total := len(scoredResults)
	start := (req.Page - 1) * req.Size
	end := start + req.Size
	
	if start >= total {
		return &DetailedSearchResult{
			Results: []*models.Wisdom{},
			Total:   total,
		}, nil
	}
	
	if end > total {
		end = total
	}
	
	// 转换结果
	var results []*models.Wisdom
	for i := start; i < end; i++ {
		wisdom := scoredResults[i].wisdom
		// 修复类型错误 - 将字符串转换�?[]string
		tagsSlice := []string{}
		if wisdom.Tags != "" {
			tagsSlice = strings.Split(wisdom.Tags, ",")
		}
		results = append(results, &models.Wisdom{
			ID:       wisdom.ID,
			Title:    wisdom.Title,
			Content:  wisdom.Content,
			Author:   wisdom.Author,
			Source:   wisdom.School,
			Category: wisdom.Category,
			Tags:     tagsSlice,
		})
	}
	
	// 生成搜索建议和分�?
	suggestions := s.generateSearchSuggestions(req.Query, scoredResults[:min(10, len(scoredResults))])
	facets := s.generateSearchFacets(scoredResults)
	
	return &DetailedSearchResult{
		Results:     results,
		Total:       total,
		Suggestions: suggestions,
		Facets:      convertFacetsToInterface(facets),
	}, nil
}

// generateSearchSuggestions 生成搜索建议
func (s *SearchService) generateSearchSuggestions(query string, results []scoredResult) []string {
	suggestions := make([]string, 0, 5)
	
	// 基于搜索结果生成相关建议
	keywordMap := make(map[string]int)
	
	for _, result := range results {
		// 从标题和内容中提取关键词
		titleWords := strings.Fields(strings.ToLower(result.wisdom.Title))
		contentWords := strings.Fields(strings.ToLower(result.wisdom.Content))
		
		for _, word := range append(titleWords, contentWords...) {
			if len(word) > 2 && !strings.Contains(query, word) {
				keywordMap[word]++
			}
		}
	}
	
	// 按频率排序并选择�?个作为建�?
	type wordFreq struct {
		word string
		freq int
	}
	
	var wordFreqs []wordFreq
	for word, freq := range keywordMap {
		wordFreqs = append(wordFreqs, wordFreq{word: word, freq: freq})
	}
	
	// 简单排�?
	for i := 0; i < len(wordFreqs)-1; i++ {
		for j := i + 1; j < len(wordFreqs); j++ {
			if wordFreqs[i].freq < wordFreqs[j].freq {
				wordFreqs[i], wordFreqs[j] = wordFreqs[j], wordFreqs[i]
			}
		}
	}
	
	// 生成建议
	for i := 0; i < min(5, len(wordFreqs)); i++ {
		suggestions = append(suggestions, query+" "+wordFreqs[i].word)
	}
	
	return suggestions
}

// generateSearchFacets 生成搜索分面
func (s *SearchService) generateSearchFacets(results []scoredResult) map[string][]FacetItem {
	facets := make(map[string][]FacetItem)
	
	// 统计分类
	categoryCount := make(map[string]int)
	schoolCount := make(map[string]int)
	authorCount := make(map[string]int)
	
	for _, result := range results {
		categoryCount[result.wisdom.Category]++
		schoolCount[result.wisdom.School]++
		authorCount[result.wisdom.Author]++
	}
	
	// 生成分类分面
	var categoryFacets []FacetItem
	for category, count := range categoryCount {
		categoryFacets = append(categoryFacets, FacetItem{
			Value: category,
			Count: count,
		})
	}
	facets["category"] = categoryFacets
	
	// 生成学派分面
	var schoolFacets []FacetItem
	for school, count := range schoolCount {
		schoolFacets = append(schoolFacets, FacetItem{
			Value: school,
			Count: count,
		})
	}
	facets["school"] = schoolFacets
	
	// 生成作者分�?
	var authorFacets []FacetItem
	for author, count := range authorCount {
		authorFacets = append(authorFacets, FacetItem{
			Value: author,
			Count: count,
		})
	}
	facets["author"] = authorFacets
	
	return facets
}

// matchesFilters 检查是否匹配过滤条�?
func (s *SearchService) matchesFilters(wisdom IndexedWisdom, req *SemanticSearchRequest) bool {
	// 分类过滤
	if req.CategoryID != "" && wisdom.Category != req.CategoryID {
		return false
	}
	
	// 学派过滤
	if req.School != "" && wisdom.School != req.School {
		return false
	}
	
	// 标签过滤
	if len(req.Tags) > 0 {
		wisdomTags := strings.Split(wisdom.Tags, ",")
		hasMatchingTag := false
		for _, reqTag := range req.Tags {
			for _, wisdomTag := range wisdomTags {
				if strings.TrimSpace(wisdomTag) == reqTag {
					hasMatchingTag = true
					break
				}
			}
			if hasMatchingTag {
				break
			}
		}
		if !hasMatchingTag {
			return false
		}
	}
	
	return true
}

// calculateRelevanceScore 计算综合相关性评�?
func (s *SearchService) calculateRelevanceScore(similarity float32, wisdom IndexedWisdom, query string) float32 {
	relevance := similarity * 0.7 // 语义相似度权�?0%
	
	// 内容质量评分权重20%
	relevance += wisdom.Score * 0.2
	
	// 关键词匹配加分权�?0%
	keywordScore := s.calculateKeywordMatchScore(wisdom, query)
	relevance += keywordScore * 0.1
	
	return relevance
}

// calculateKeywordMatchScore 计算关键词匹配评�?
func (s *SearchService) calculateKeywordMatchScore(wisdom IndexedWisdom, query string) float32 {
	queryLower := strings.ToLower(query)
	titleLower := strings.ToLower(wisdom.Title)
	contentLower := strings.ToLower(wisdom.Content)
	
	score := float32(0)
	
	// 标题匹配
	if strings.Contains(titleLower, queryLower) {
		score += 1.0
	}
	
	// 内容匹配
	if strings.Contains(contentLower, queryLower) {
		score += 0.5
	}
	
	// 作者匹�?
	if strings.Contains(strings.ToLower(wisdom.Author), queryLower) {
		score += 0.3
	}
	
	return score
}

// getOrGenerateVector 获取或生成向�?
func (s *SearchService) getOrGenerateVector(ctx context.Context, text string) ([]float32, error) {
	// 检查缓�?
	s.vectorCacheMu.RLock()
	if vector, exists := s.vectorCache[text]; exists {
		s.vectorCacheMu.RUnlock()
		return vector, nil
	}
	s.vectorCacheMu.RUnlock()
	
	// 生成向量
	vector, err := s.aiService.GetEmbedding(ctx, text)
	if err != nil {
		return nil, err
	}
	
	// 缓存向量
	s.vectorCacheMu.Lock()
	s.vectorCache[text] = vector
	s.vectorCacheMu.Unlock()
	
	return vector, nil
}

// EnhancedSearchResult 增强的搜索结�?
type EnhancedSearchResult struct {
	Results     []*models.Wisdom       `json:"results"`
	Total       int                    `json:"total"`
	SearchType  string                 `json:"search_type"` // semantic, keyword, hybrid
	QueryTime   time.Duration          `json:"query_time"`
	Suggestions []string               `json:"suggestions"`
	Facets      map[string]interface{} `json:"facets"`
}

// DetailedSearchResult 详细搜索结果
type DetailedSearchResult struct {
	Results     []*models.Wisdom       `json:"results"`
	Total       int                    `json:"total"`
	Suggestions []string               `json:"suggestions"`
	Facets      map[string]interface{} `json:"facets"`
}

// convertFacetsToInterface converts map[string][]FacetItem to map[string]interface{}
func convertFacetsToInterface(facets map[string][]FacetItem) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range facets {
		result[key] = value
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
