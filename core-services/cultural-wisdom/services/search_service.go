package services

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
)

// SearchService 搜索服务
type SearchService struct {
	db    *gorm.DB
	cache *CacheService
}

// NewSearchService 创建搜索服务实例
func NewSearchService(db *gorm.DB, cache *CacheService) *SearchService {
	return &SearchService{
		db:    db,
		cache: cache,
	}
}

// SearchWisdom 搜索智慧内容
func (s *SearchService) SearchWisdom(ctx context.Context, query string, limit int, offset int) ([]*models.Wisdom, error) {
	if query == "" {
		return nil, fmt.Errorf("搜索关键词不能为空")
	}

	// 尝试从缓存获取搜索结果
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

	// 转换为搜索结果
	results := make([]*models.Wisdom, len(wisdoms))
	for i, wisdom := range wisdoms {
		results[i] = &models.Wisdom{
			ID:       wisdom.ID,
			Title:    wisdom.Title,
			Content:  wisdom.Content,
			Author:   wisdom.Author,
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

// SearchRequest 搜索请求结构体
type SearchRequest struct {
	Query      string   `json:"query"`
	CategoryID string   `json:"category_id"`
	School     string   `json:"school"`
	Tags       []string `json:"tags"`
	Page       int      `json:"page"`
	Size       int      `json:"size"`
}

// SemanticSearchRequest 语义搜索请求结构体
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
		return nil, fmt.Errorf("搜索关键词不能为空")
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

// SemanticSearch 语义搜索（需要AI集成）
func (s *SearchService) SemanticSearch(ctx context.Context, req *SemanticSearchRequest) ([]*models.Wisdom, error) {
	if req.Query == "" {
		return nil, fmt.Errorf("搜索关键词不能为空")
	}

	// TODO: 集成AI服务进行语义搜索
	// 目前先使用关键词搜索作为替代
	offset := (req.Page - 1) * req.Size
	return s.SearchWisdom(ctx, req.Query, req.Size, offset)
}

// GetSearchSuggestions 获取搜索建议
func (s *SearchService) GetSearchSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	if query == "" {
		return []string{}, nil
	}

	// 尝试从缓存获取
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
	
	// 去重并限制数量
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
	// 尝试从缓存获取
	if searches, err := s.cache.GetPopularSearches(ctx); err == nil && searches != nil {
		if len(searches) > limit {
			searches = searches[:limit]
		}
		return searches, nil
	}

	// 模拟热门搜索数据
	popularSearches := []PopularSearch{
		{Query: "道德经", Count: 1500},
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
			return nil, fmt.Errorf("智慧内容不存在")
		}
		return nil, fmt.Errorf("查询智慧内容失败: %w", err)
	}

	return &wisdom, nil
}

// GetCategories 获取分类列表
func (s *SearchService) GetCategories(ctx context.Context) ([]models.Category, error) {
	// 尝试从缓存获取
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

// SearchByCategory 按分类搜索
func (s *SearchService) SearchByCategory(ctx context.Context, category string, limit int, offset int) ([]*models.Wisdom, error) {
	if category == "" {
		return nil, fmt.Errorf("分类不能为空")
	}

	var wisdoms []models.CulturalWisdom
	
	// 按分类查询
	if err := s.db.WithContext(ctx).
		Where("category = ?", category).
		Limit(limit).
		Offset(offset).
		Find(&wisdoms).Error; err != nil {
		return nil, fmt.Errorf("按分类搜索失败: %w", err)
	}

	// 转换为搜索结果
	results := make([]*models.Wisdom, len(wisdoms))
	for i, wisdom := range wisdoms {
		results[i] = &models.Wisdom{
			ID:       wisdom.ID,
			Title:    wisdom.Title,
			Content:  wisdom.Content,
			Author:   wisdom.Author,
			Category: wisdom.Category,
			Tags:     wisdom.Tags,
		}
	}

	return results, nil
}

// AdvancedSearchRequest 高级搜索请求结构体
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

// SearchFilters 搜索过滤器
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

	// 作者过滤
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
		if req.Query != "" {
			// 简单的相关性排序：标题匹配优先
			orderClause = "CASE WHEN title ILIKE '%" + req.Query + "%' THEN 1 ELSE 2 END, view_count DESC"
		} else {
			orderClause = "view_count DESC"
		}
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

	// 获取作者
	var authors []string
	if err := s.db.WithContext(ctx).Model(&models.CulturalWisdom{}).
		Select("DISTINCT author").Where("author != ''").
		Order("author ASC").Limit(100).Pluck("author", &authors).Error; err != nil {
		return nil, fmt.Errorf("获取作者失败: %w", err)
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

	// 作者分面
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
		return nil, fmt.Errorf("获取作者分面失败: %w", err)
	}
	facets["authors"] = authorFacets

	return facets, nil
}

// ... existing code ...