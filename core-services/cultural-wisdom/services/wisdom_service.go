package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
)

// WisdomService 智慧服务
type WisdomService struct {
	db    *gorm.DB
	cache *CacheService
}

// NewWisdomService 创建智慧服务实例
func NewWisdomService(db *gorm.DB, cache *CacheService) *WisdomService {
	return &WisdomService{
		db:    db,
		cache: cache,
	}
}

// GetWisdomList 获取智慧列表
func (s *WisdomService) GetWisdomList(ctx context.Context, filter *models.WisdomFilter) ([]models.WisdomSummary, int64, error) {
	// 尝试从缓存获取
	if s.cache != nil {
		if wisdoms, total, err := s.cache.GetWisdomList(ctx, filter); err == nil && wisdoms != nil {
			return wisdoms, total, nil
		}
	}

	query := s.db.Model(&models.CulturalWisdom{})

	// 应用过滤条件
	if filter.CategoryID != "" {
		query = query.Where("category = ?", filter.CategoryID)
	}
	if filter.School != "" {
		query = query.Where("school = ?", filter.School)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.AuthorID != "" {
		query = query.Where("author_id = ?", filter.AuthorID)
	}
	if filter.SearchQuery != "" {
		searchPattern := "%" + filter.SearchQuery + "%"
		query = query.Where("title ILIKE ? OR content ILIKE ? OR author ILIKE ?", 
			searchPattern, searchPattern, searchPattern)
	}
	if len(filter.Tags) > 0 {
		for _, tag := range filter.Tags {
			query = query.Where("tags ILIKE ?", "%"+tag+"%")
		}
	}
	if len(filter.Difficulty) > 0 {
		query = query.Where("difficulty IN ?", filter.Difficulty)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count wisdoms: %w", err)
	}

	// 应用排序
	sortBy := "created_at"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortOrder := "DESC"
	if filter.SortOrder != "" {
		sortOrder = filter.SortOrder
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// 应用分页
	offset := (filter.Page - 1) * filter.Size
	query = query.Offset(offset).Limit(filter.Size)

	var wisdoms []models.CulturalWisdom
	if err := query.Find(&wisdoms).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get wisdoms: %w", err)
	}

	// 转换为WisdomSummary格式
	summaries := make([]models.WisdomSummary, len(wisdoms))
	for i, wisdom := range wisdoms {
		summaries[i] = models.WisdomSummary{
			ID:        wisdom.ID,
			Title:     wisdom.Title,
			Summary:   wisdom.Summary,
			Tags:      wisdom.Tags,
			ViewCount: wisdom.ViewCount,
			LikeCount: wisdom.LikeCount,
			CreatedAt: wisdom.CreatedAt,
		}
	}

	// 缓存结果
	if s.cache != nil {
		s.cache.SetWisdomList(ctx, filter, summaries, total)
	}

	return summaries, total, nil
}

// GetWisdomByID 根据ID获取智慧内容
func (s *WisdomService) GetWisdomByID(ctx context.Context, id string) (*models.CulturalWisdom, error) {
	// 尝试从缓存获取
	if s.cache != nil {
		if wisdom, err := s.cache.GetWisdom(ctx, id); err == nil && wisdom != nil {
			return wisdom, nil
		}
	}

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
	
	// 缓存结果
	if s.cache != nil {
		s.cache.SetWisdom(ctx, &wisdom)
	}

	return &wisdom, nil
}

// IncrementViewCount 增加浏览次数
func (s *WisdomService) IncrementViewCount(ctx context.Context, id string) error {
	// 使用Redis计数器增加浏览次数
	if s.cache != nil {
		if err := s.cache.IncrementViewCount(ctx, id); err == nil {
			return nil
		}
	}

	// 如果Redis不可用，直接更新数据库
	if err := s.db.Model(&models.CulturalWisdom{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error; err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	// 使缓存失效
	if s.cache != nil {
		s.cache.InvalidateWisdom(ctx, id)
	}

	return nil
}

// CreateWisdom 创建智慧内容
func (s *WisdomService) CreateWisdom(ctx context.Context, wisdom *models.CulturalWisdom) (*models.CulturalWisdom, error) {
	if wisdom.Title == "" {
		return nil, fmt.Errorf("标题不能为空")
	}
	if wisdom.Content == "" {
		return nil, fmt.Errorf("内容不能为空")
	}

	// 设置创建时间
	now := time.Now()
	wisdom.CreatedAt = now
	wisdom.UpdatedAt = now
	
	// 生成ID（这里简化处理，实际应该使用UUID）
	wisdom.ID = fmt.Sprintf("wisdom_%d", now.Unix())

	if err := s.db.WithContext(ctx).Create(wisdom).Error; err != nil {
		return nil, fmt.Errorf("创建智慧内容失败: %w", err)
	}

	// 使相关缓存失效
	if s.cache != nil {
		s.cache.InvalidateWisdomList(ctx)
		s.cache.InvalidateStats(ctx)
	}

	return wisdom, nil
}

// UpdateWisdom 更新智慧内容
func (s *WisdomService) UpdateWisdom(ctx context.Context, id string, req *models.UpdateWisdomRequest, userID string) (*models.CulturalWisdom, error) {
	// 检查记录是否存在
	var wisdom models.CulturalWisdom
	if err := s.db.Where("id = ?", id).First(&wisdom).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("wisdom not found")
		}
		return nil, fmt.Errorf("failed to get wisdom: %w", err)
	}

	// 检查权限（只有作者可以更新）
	if wisdom.AuthorID != userID {
		return nil, fmt.Errorf("permission denied: only author can update")
	}

	// 更新字段
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Summary != "" {
		updates["summary"] = req.Summary
	}
	if req.CategoryID != "" {
		updates["category"] = req.CategoryID
	}
	if req.Tags != nil {
		updates["tags"] = req.Tags
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if err := s.db.Model(&wisdom).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update wisdom: %w", err)
	}

	// 重新获取更新后的记录
	if err := s.db.Where("id = ?", id).First(&wisdom).Error; err != nil {
		return nil, fmt.Errorf("failed to get updated wisdom: %w", err)
	}

	// 使相关缓存失效
	if s.cache != nil {
		s.cache.InvalidateWisdom(ctx, id)
		s.cache.InvalidateWisdomList(ctx)
		s.cache.InvalidateStats(ctx)
	}

	return &wisdom, nil
}

// DeleteWisdom 删除智慧内容
func (s *WisdomService) DeleteWisdom(ctx context.Context, id string, userID string) error {
	// 检查记录是否存在
	var wisdom models.CulturalWisdom
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&wisdom).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("智慧内容不存在")
		}
		return fmt.Errorf("查询智慧内容失败: %w", err)
	}

	// 检查权限（只有作者可以删除）
	if wisdom.AuthorID != userID {
		return fmt.Errorf("权限不足：只有作者可以删除")
	}

	// 删除记录
	if err := s.db.WithContext(ctx).Delete(&wisdom).Error; err != nil {
		return fmt.Errorf("删除智慧内容失败: %w", err)
	}

	// 使相关缓存失效
	if s.cache != nil {
		s.cache.InvalidateWisdom(ctx, id)
		s.cache.InvalidateWisdomList(ctx)
		s.cache.InvalidateStats(ctx)
	}

	return nil
}

// BatchDeleteWisdom 批量删除智慧内容
func (s *WisdomService) BatchDeleteWisdom(ctx context.Context, ids []string, userID string) (int, error) {
	if len(ids) == 0 {
		return 0, fmt.Errorf("删除ID列表不能为空")
	}

	deletedCount := 0
	
	// 逐个删除，确保权限检查
	for _, id := range ids {
		// 检查记录是否存在和权限
		var wisdom models.CulturalWisdom
		if err := s.db.WithContext(ctx).Where("id = ?", id).First(&wisdom).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				continue // 跳过不存在的记录
			}
			return deletedCount, fmt.Errorf("查询智慧内容失败: %w", err)
		}

		// 检查权限（只有作者可以删除）
		if wisdom.AuthorID != userID {
			continue // 跳过无权限的记录
		}

		// 删除记录
		if err := s.db.WithContext(ctx).Delete(&wisdom).Error; err != nil {
			return deletedCount, fmt.Errorf("删除智慧内容失败: %w", err)
		}

		deletedCount++

		// 使相关缓存失效
		if s.cache != nil {
			s.cache.InvalidateWisdom(ctx, id)
		}
	}

	// 使列表缓存失效
	if s.cache != nil && deletedCount > 0 {
		s.cache.InvalidateWisdomList(ctx)
		s.cache.InvalidateStats(ctx)
	}

	return deletedCount, nil
}

// AdvancedSearchWisdom 高级搜索智慧内容
func (s *WisdomService) AdvancedSearchWisdom(ctx context.Context, filter *models.WisdomFilter) ([]models.WisdomSummary, int64, error) {
	if filter == nil {
		return nil, 0, fmt.Errorf("搜索过滤器不能为空")
	}

	// 构建查询
	query := s.db.WithContext(ctx).Model(&models.CulturalWisdom{})

	// 关键词搜索
	if filter.SearchQuery != "" {
		query = query.Where(
			"title LIKE ? OR content LIKE ? OR author LIKE ? OR tags LIKE ?",
			"%"+filter.SearchQuery+"%", "%"+filter.SearchQuery+"%", 
			"%"+filter.SearchQuery+"%", "%"+filter.SearchQuery+"%",
		)
	}

	// 分类过滤
	if filter.CategoryID != "" {
		query = query.Where("category = ?", filter.CategoryID)
	}

	// 学派过滤
	if filter.School != "" {
		query = query.Where("school = ?", filter.School)
	}

	// 作者过滤
	if filter.AuthorID != "" {
		query = query.Where("author_id = ?", filter.AuthorID)
	}

	// 状态过滤
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	// 标签过滤
	if len(filter.Tags) > 0 {
		for _, tag := range filter.Tags {
			query = query.Where("JSON_CONTAINS(tags, ?)", fmt.Sprintf(`"%s"`, tag))
		}
	}

	// 难度过滤
	if len(filter.Difficulty) > 0 {
		difficulties := make([]string, len(filter.Difficulty))
		for i, d := range filter.Difficulty {
			difficulties[i] = strconv.Itoa(d)
		}
		query = query.Where("difficulty IN ?", difficulties)
	}

	// 日期范围过滤
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", filter.DateFrom)
	}
	if filter.DateTo != nil {
		query = query.Where("created_at <= ?", filter.DateTo)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计搜索结果失败: %w", err)
	}

	// 排序
	orderClause := fmt.Sprintf("%s %s", filter.SortBy, filter.SortOrder)
	query = query.Order(orderClause)

	// 分页
	offset := (filter.Page - 1) * filter.Size
	query = query.Offset(offset).Limit(filter.Size)

	// 执行查询
	var wisdoms []models.CulturalWisdom
	if err := query.Find(&wisdoms).Error; err != nil {
		return nil, 0, fmt.Errorf("执行搜索查询失败: %w", err)
	}

	// 转换为摘要格式
	summaries := make([]models.WisdomSummary, len(wisdoms))
	for i, wisdom := range wisdoms {
		difficulty, _ := strconv.Atoi(wisdom.Difficulty)
		summaries[i] = models.WisdomSummary{
			ID:         wisdom.ID,
			Title:      wisdom.Title,
			Summary:    wisdom.Summary,
			Category:   models.Category{Name: wisdom.Category},
			Tags:       []string(wisdom.Tags),
			Difficulty: difficulty,
			ViewCount:  wisdom.ViewCount,
			LikeCount:  wisdom.LikeCount,
			CreatedAt:  wisdom.CreatedAt,
		}
	}

	return summaries, total, nil
}

// GetWisdomStats 获取智慧统计信息
func (s *WisdomService) GetWisdomStats(ctx context.Context) (*models.WisdomStats, error) {
	// 尝试从缓存获取
	if s.cache != nil {
		if stats, err := s.cache.GetStats(ctx); err == nil && stats != nil {
			return stats, nil
		}
	}

	stats := &models.WisdomStats{
		CategoryStats:   make(map[string]int64),
		SchoolStats:     make(map[string]int64),
		DifficultyStats: make(map[int]int64),
		MonthlyStats:    []models.MonthlyCount{},
	}

	// 获取总数
	if err := s.db.Model(&models.CulturalWisdom{}).Count(&stats.TotalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// 获取已发布数量
	if err := s.db.Model(&models.CulturalWisdom{}).Where("status = ?", "published").Count(&stats.PublishedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get published count: %w", err)
	}

	// 获取草稿数量
	if err := s.db.Model(&models.CulturalWisdom{}).Where("status = ?", "draft").Count(&stats.DraftCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get draft count: %w", err)
	}

	// 获取分类统计
	var categoryStats []struct {
		Category string
		Count    int64
	}
	if err := s.db.Model(&models.CulturalWisdom{}).
		Select("category, count(*) as count").
		Group("category").
		Scan(&categoryStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get category stats: %w", err)
	}
	for _, stat := range categoryStats {
		stats.CategoryStats[stat.Category] = stat.Count
	}

	// 获取学派统计
	var schoolStats []struct {
		School string
		Count  int64
	}
	if err := s.db.Model(&models.CulturalWisdom{}).
		Select("school, count(*) as count").
		Group("school").
		Scan(&schoolStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get school stats: %w", err)
	}
	for _, stat := range schoolStats {
		stats.SchoolStats[stat.School] = stat.Count
	}

	// 获取难度统计
	var difficultyStats []struct {
		Difficulty string
		Count      int64
	}
	if err := s.db.Model(&models.CulturalWisdom{}).
		Select("difficulty, count(*) as count").
		Group("difficulty").
		Scan(&difficultyStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get difficulty stats: %w", err)
	}
	for _, stat := range difficultyStats {
		// 将difficulty字符串转换为int
		var diffLevel int
		switch stat.Difficulty {
		case "easy":
			diffLevel = 1
		case "medium":
			diffLevel = 2
		case "hard":
			diffLevel = 3
		default:
			diffLevel = 2
		}
		stats.DifficultyStats[diffLevel] = stat.Count
	}

	// 获取月度统计（最近12个月）
	var monthlyStats []struct {
		Year  int
		Month int
		Count int64
	}
	if err := s.db.Model(&models.CulturalWisdom{}).
		Select("EXTRACT(year FROM created_at) as year, EXTRACT(month FROM created_at) as month, count(*) as count").
		Where("created_at >= ?", time.Now().AddDate(0, -12, 0)).
		Group("year, month").
		Order("year, month").
		Scan(&monthlyStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get monthly stats: %w", err)
	}
	for _, stat := range monthlyStats {
		stats.MonthlyStats = append(stats.MonthlyStats, models.MonthlyCount{
			Year:  stat.Year,
			Month: stat.Month,
			Count: stat.Count,
		})
	}

	// 缓存结果
	if s.cache != nil {
		s.cache.SetStats(ctx, stats)
	}

	return stats, nil
}