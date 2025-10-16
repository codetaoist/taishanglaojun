package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
)

// TagService 
type TagService struct {
	db    *gorm.DB
	cache *CacheService
}

// NewTagService 
func NewTagService(db *gorm.DB, cache *CacheService) *TagService {
	return &TagService{
		db:    db,
		cache: cache,
	}
}

// GetTags 
func (s *TagService) GetTags(ctx context.Context, page, size int, search, sortBy, sortOrder string) ([]models.WisdomTag, int64, error) {
	var tags []models.WisdomTag
	var total int64
	
	query := s.db.WithContext(ctx).Model(&models.WisdomTag{}).Where("is_active = ?", true)
	
	// 
	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	
	// 
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tags: %w", err)
	}
	
	// 
	orderClause := "usage_count DESC"
	switch sortBy {
	case "name":
		orderClause = "name " + sortOrder
	case "usage_count":
		orderClause = "usage_count " + sortOrder
	case "created_at":
		orderClause = "created_at " + sortOrder
	}
	
	// 
	offset := (page - 1) * size
	if err := query.Order(orderClause).Offset(offset).Limit(size).Find(&tags).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get tags: %w", err)
	}
	
	return tags, total, nil
}

// GetTagByID ID
func (s *TagService) GetTagByID(ctx context.Context, id int) (*models.WisdomTag, error) {
	var tag models.WisdomTag
	
	if err := s.db.WithContext(ctx).Where("id = ? AND is_active = ?", id, true).First(&tag).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("tag not found")
		}
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}
	
	return &tag, nil
}

// GetTagByName 
func (s *TagService) GetTagByName(ctx context.Context, name string) (*models.WisdomTag, error) {
	var tag models.WisdomTag
	
	if err := s.db.WithContext(ctx).Where("name = ? AND is_active = ?", name, true).First(&tag).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("tag not found")
		}
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}
	
	return &tag, nil
}

// CreateTag 
func (s *TagService) CreateTag(ctx context.Context, tag *models.WisdomTag) (*models.WisdomTag, error) {
	// ?
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.WisdomTag{}).Where("name = ? AND is_active = ?", tag.Name, true).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to check tag name: %w", err)
	}
	
	if count > 0 {
		return nil, fmt.Errorf("tag name already exists")
	}
	
	tag.CreatedAt = time.Now()
	tag.UsageCount = 0
	
	if err := s.db.WithContext(ctx).Create(tag).Error; err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}
	
	// 
	s.cache.DeletePattern(ctx, "tags:*")
	
	return tag, nil
}

// UpdateTag 
func (s *TagService) UpdateTag(ctx context.Context, id int, req interface{}) (*models.WisdomTag, error) {
	tag, err := s.GetTagByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// 
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	
	// req
	// 
	
	if err := s.db.WithContext(ctx).Model(tag).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update tag: %w", err)
	}
	
	// 
	s.cache.DeletePattern(ctx, "tags:*")
	
	return tag, nil
}

// DeleteTag ?
func (s *TagService) DeleteTag(ctx context.Context, id int) error {
	// ?
	var relationCount int64
	if err := s.db.WithContext(ctx).Model(&models.WisdomTagRelation{}).Where("tag_id = ?", id).Count(&relationCount).Error; err != nil {
		return fmt.Errorf("failed to check tag relations: %w", err)
	}
	
	if relationCount > 0 {
		return fmt.Errorf("cannot delete tag with associated wisdom content")
	}
	
	// ?
	if err := s.db.WithContext(ctx).Model(&models.WisdomTag{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}
	
	// 
	s.cache.DeletePattern(ctx, "tags:*")
	
	return nil
}

// GetPopularTags 
func (s *TagService) GetPopularTags(ctx context.Context, limit int) ([]models.WisdomTag, error) {
	var tags []models.WisdomTag
	
	if err := s.db.WithContext(ctx).Where("is_active = ? AND usage_count > 0", true).
		Order("usage_count DESC").
		Limit(limit).
		Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("failed to get popular tags: %w", err)
	}
	
	return tags, nil
}

// GetTagStats 
func (s *TagService) GetTagStats(ctx context.Context, tagID int) (*models.TagStats, error) {
	var stats models.TagStats
	
	// 
	tag, err := s.GetTagByID(ctx, tagID)
	if err != nil {
		return nil, err
	}
	
	// ?
	var wisdomCount int64
	if err := s.db.WithContext(ctx).Model(&models.WisdomTagRelation{}).Where("tag_id = ?", tagID).Count(&wisdomCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get wisdom count: %w", err)
	}
	
	// 
	var totalViews int64
	if err := s.db.WithContext(ctx).Table("cultural_wisdom cw").
		Joins("JOIN wisdom_tag_relations wtr ON cw.id = wtr.wisdom_id").
		Where("wtr.tag_id = ?", tagID).
		Select("COALESCE(SUM(cw.view_count), 0)").
		Scan(&totalViews).Error; err != nil {
		return nil, fmt.Errorf("failed to get total views: %w", err)
	}
	
	// 
	var totalLikes int64
	if err := s.db.WithContext(ctx).Table("cultural_wisdom cw").
		Joins("JOIN wisdom_tag_relations wtr ON cw.id = wtr.wisdom_id").
		Where("wtr.tag_id = ?", tagID).
		Select("COALESCE(SUM(cw.like_count), 0)").
		Scan(&totalLikes).Error; err != nil {
		return nil, fmt.Errorf("failed to get total likes: %w", err)
	}
	
	stats = models.TagStats{
		TagID:       tagID,
		TagName:     tag.Name,
		UsageCount:  tag.UsageCount,
		WisdomCount: wisdomCount,
		TotalViews:  totalViews,
		TotalLikes:  totalLikes,
	}
	
	return &stats, nil
}

// IncrementUsageCount 
func (s *TagService) IncrementUsageCount(ctx context.Context, tagID int) error {
	if err := s.db.WithContext(ctx).Model(&models.WisdomTag{}).Where("id = ?", tagID).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error; err != nil {
		return fmt.Errorf("failed to increment usage count: %w", err)
	}
	
	// 
	s.cache.DeletePattern(ctx, "tags:*")
	
	return nil
}

// DecrementUsageCount 
func (s *TagService) DecrementUsageCount(ctx context.Context, tagID int) error {
	if err := s.db.WithContext(ctx).Model(&models.WisdomTag{}).Where("id = ? AND usage_count > 0", tagID).
		UpdateColumn("usage_count", gorm.Expr("usage_count - 1")).Error; err != nil {
		return fmt.Errorf("failed to decrement usage count: %w", err)
	}
	
	// 
	s.cache.DeletePattern(ctx, "tags:*")
	
	return nil
}

// GetOrCreateTag ?
func (s *TagService) GetOrCreateTag(ctx context.Context, name string) (*models.WisdomTag, error) {
	// ?
	tag, err := s.GetTagByName(ctx, name)
	if err == nil {
		return tag, nil
	}
	
	// ?
	newTag := &models.WisdomTag{
		Name:        name,
		Description: "",
		IsActive:    true,
	}
	
	return s.CreateTag(ctx, newTag)
}

// SearchTags 
func (s *TagService) SearchTags(ctx context.Context, keyword string, limit int) ([]models.WisdomTag, error) {
	var tags []models.WisdomTag
	
	query := s.db.WithContext(ctx).Where("is_active = ?", true)
	
	if keyword != "" {
		query = query.Where("name ILIKE ?", "%"+keyword+"%")
	}
	
	if err := query.Order("usage_count DESC").Limit(limit).Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("failed to search tags: %w", err)
	}
	
	return tags, nil
}

