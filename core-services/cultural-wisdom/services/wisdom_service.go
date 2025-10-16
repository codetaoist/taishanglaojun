package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
	"gorm.io/gorm"
)

// WisdomService 智慧服务
// 提供对文化智慧的 CRUD 操作和搜索功能
type WisdomService struct {
	db    *gorm.DB
	cache *CacheService
}

// NewWisdomService 创建新的智慧服务实例
func NewWisdomService(db *gorm.DB, cache *CacheService) *WisdomService {
	return &WisdomService{
		db:    db,
		cache: cache,
	}
}

// GetWisdomList 获取智慧列表
// 根据筛选条件查询文化智慧列表，支持分页和排序
func (s *WisdomService) GetWisdomList(ctx context.Context, filter *models.WisdomFilter) ([]models.WisdomSummary, int64, error) {
	// 从缓存中获取智慧列表
	if s.cache != nil {
		if wisdoms, total, err := s.cache.GetWisdomList(ctx, filter); err == nil && wisdoms != nil {
			return wisdoms, total, nil
		}
	}

	query := s.db.Model(&models.CulturalWisdom{})

	//
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

	//
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count wisdoms: %w", err)
	}

	//
	sortBy := "created_at"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortOrder := "DESC"
	if filter.SortOrder != "" {
		sortOrder = filter.SortOrder
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	//
	offset := (filter.Page - 1) * filter.Size
	query = query.Offset(offset).Limit(filter.Size)

	var wisdoms []models.CulturalWisdom
	if err := query.Find(&wisdoms).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get wisdoms: %w", err)
	}

	// WisdomSummary
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

	//
	if s.cache != nil {
		s.cache.SetWisdomList(ctx, filter, summaries, total)
	}

	return summaries, total, nil
}

// GetWisdomByID 获取智慧内容
// 根据 ID 查询指定的文化智慧内容
func (s *WisdomService) GetWisdomByID(ctx context.Context, id string) (*models.CulturalWisdom, error) {
	// ?
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
		return nil, fmt.Errorf(": %w", err)
	}

	//
	if s.cache != nil {
		s.cache.SetWisdom(ctx, &wisdom)
	}

	return &wisdom, nil
}

// IncrementViewCount 增加视图计数
// 增加指定文化智慧的视图计数，支持缓存更新
func (s *WisdomService) IncrementViewCount(ctx context.Context, id string) error {
	// Redis?
	if s.cache != nil {
		if err := s.cache.IncrementViewCount(ctx, id); err == nil {
			return nil
		}
	}

	// Redis?
	if err := s.db.Model(&models.CulturalWisdom{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error; err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	// ?
	if s.cache != nil {
		s.cache.InvalidateWisdom(ctx, id)
	}

	return nil
}

// CreateWisdom 创建新的文化智慧
// 验证输入参数，创建新的文化智慧记录，并更新缓存
func (s *WisdomService) CreateWisdom(ctx context.Context, wisdom *models.CulturalWisdom) (*models.CulturalWisdom, error) {
	if wisdom.Title == "" {
		return nil, fmt.Errorf("标题不能为空")
	}
	if wisdom.Content == "" {
		return nil, fmt.Errorf("内容不能为空")
	}
	if wisdom.Summary == "" {
		return nil, fmt.Errorf("摘要不能为空")
	}

	//
	now := time.Now()
	wisdom.CreatedAt = now
	wisdom.UpdatedAt = now

	// IDUUID?
	wisdom.ID = fmt.Sprintf("wisdom_%d", now.Unix())

	if err := s.db.WithContext(ctx).Create(wisdom).Error; err != nil {
		return nil, fmt.Errorf("failed to create wisdom: %w", err)
	}

	// ?
	if s.cache != nil {
		s.cache.InvalidateWisdomList(ctx)
		s.cache.InvalidateStats(ctx)
	}

	return wisdom, nil
}

// UpdateWisdom 更新文化智慧
// 验证输入参数，更新指定 ID 的文化智慧记录，并更新缓存
func (s *WisdomService) UpdateWisdom(ctx context.Context, id string, req *models.UpdateWisdomRequest, userID string) (*models.CulturalWisdom, error) {
	// ?
	var wisdom models.CulturalWisdom
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&wisdom).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("智慧内容不存在")
		}
		return nil, fmt.Errorf("failed to get wisdom: %w", err)
	}

	//
	if wisdom.AuthorID != userID {
		return nil, fmt.Errorf("permission denied: only author can update")
	}

	//
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

	//
	if err := s.db.Where("id = ?", id).First(&wisdom).Error; err != nil {
		return nil, fmt.Errorf("failed to get updated wisdom: %w", err)
	}

	// ?
	if s.cache != nil {
		s.cache.InvalidateWisdom(ctx, id)
		s.cache.InvalidateWisdomList(ctx)
		s.cache.InvalidateStats(ctx)
	}

	return &wisdom, nil
}

// DeleteWisdom 删除文化智慧
// 验证用户权限，删除指定 ID 的文化智慧记录，并更新缓存
func (s *WisdomService) DeleteWisdom(ctx context.Context, id string, userID string) error {
	// ?
	var wisdom models.CulturalWisdom
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&wisdom).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("智慧内容不存在")
		}
		return fmt.Errorf("failed to get wisdom: %w", err)
	}

	//
	if wisdom.AuthorID != userID {
		return fmt.Errorf("无权限删除此内容")
	}

	//
	if err := s.db.WithContext(ctx).Delete(&wisdom).Error; err != nil {
		return fmt.Errorf(": %w", err)
	}

	// ?
	if s.cache != nil {
		s.cache.InvalidateWisdom(ctx, id)
		s.cache.InvalidateWisdomList(ctx)
		s.cache.InvalidateStats(ctx)
	}

	return nil
}

// BatchDeleteWisdom 批量删除文化智慧
// 验证用户权限，批量删除指定 ID 的文化智慧记录，并更新缓存
func (s *WisdomService) BatchDeleteWisdom(ctx context.Context, ids []string, userID string) (int, error) {
	if len(ids) == 0 {
		return 0, fmt.Errorf("ID列表不能为空")
	}

	deletedCount := 0

	// ?
	for _, id := range ids {
		//
		var wisdom models.CulturalWisdom
		if err := s.db.WithContext(ctx).Where("id = ?", id).First(&wisdom).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				continue //
			}
			return deletedCount, fmt.Errorf(": %w", err)
		}

		//
		if wisdom.AuthorID != userID {
			continue //
		}

		//
		if err := s.db.WithContext(ctx).Delete(&wisdom).Error; err != nil {
			return deletedCount, fmt.Errorf(": %w", err)
		}

		deletedCount++

		// ?
		if s.cache != nil {
			s.cache.InvalidateWisdom(ctx, id)
		}
	}

	// ?
	if s.cache != nil && deletedCount > 0 {
		s.cache.InvalidateWisdomList(ctx)
		s.cache.InvalidateStats(ctx)
	}

	return deletedCount, nil
}

// AdvancedSearchWisdom 高级搜索文化智慧
// 基于搜索过滤器，返回符合条件的文化智慧摘要列表和总数
func (s *WisdomService) AdvancedSearchWisdom(ctx context.Context, filter *models.WisdomFilter) ([]models.WisdomSummary, int64, error) {
	if filter == nil {
		return nil, 0, fmt.Errorf("搜索过滤器不能为空")
	}

	//
	query := s.db.WithContext(ctx).Model(&models.CulturalWisdom{})

	// ?
	if filter.SearchQuery != "" {
		query = query.Where(
			"title LIKE ? OR content LIKE ? OR author LIKE ? OR tags LIKE ?",
			"%"+filter.SearchQuery+"%", "%"+filter.SearchQuery+"%",
			"%"+filter.SearchQuery+"%", "%"+filter.SearchQuery+"%",
		)
	}

	//
	if filter.CategoryID != "" {
		query = query.Where("category = ?", filter.CategoryID)
	}

	//
	if filter.School != "" {
		query = query.Where("school = ?", filter.School)
	}

	// ?
	if filter.AuthorID != "" {
		query = query.Where("author_id = ?", filter.AuthorID)
	}

	// ?
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	//
	if len(filter.Tags) > 0 {
		for _, tag := range filter.Tags {
			query = query.Where("JSON_CONTAINS(tags, ?)", fmt.Sprintf(`"%s"`, tag))
		}
	}

	//
	if len(filter.Difficulty) > 0 {
		difficulties := make([]string, len(filter.Difficulty))
		for i, d := range filter.Difficulty {
			difficulties[i] = strconv.Itoa(d)
		}
		query = query.Where("difficulty IN ?", difficulties)
	}

	//
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", filter.DateFrom)
	}
	if filter.DateTo != nil {
		query = query.Where("created_at <= ?", filter.DateTo)
	}

	//
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf(": %w", err)
	}

	//
	orderClause := fmt.Sprintf("%s %s", filter.SortBy, filter.SortOrder)
	query = query.Order(orderClause)

	//
	offset := (filter.Page - 1) * filter.Size
	query = query.Offset(offset).Limit(filter.Size)

	//
	var wisdoms []models.CulturalWisdom
	if err := query.Find(&wisdoms).Error; err != nil {
		return nil, 0, fmt.Errorf(": %w", err)
	}

	// ?
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

// GetWisdomStats 获取文化智慧统计信息
// 从缓存中获取统计信息，若不存在则从数据库查询并更新缓存
func (s *WisdomService) GetWisdomStats(ctx context.Context) (*models.WisdomStats, error) {
	// ?
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

	//
	if err := s.db.Model(&models.CulturalWisdom{}).Count(&stats.TotalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// ?
	if err := s.db.Model(&models.CulturalWisdom{}).Where("status = ?", "published").Count(&stats.PublishedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get published count: %w", err)
	}

	//
	if err := s.db.Model(&models.CulturalWisdom{}).Where("status = ?", "draft").Count(&stats.DraftCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get draft count: %w", err)
	}

	//
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

	//
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

	//
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
		// difficultyint
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

	// ?2?
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

	//
	if s.cache != nil {
		s.cache.SetStats(ctx, stats)
	}

	return stats, nil
}
