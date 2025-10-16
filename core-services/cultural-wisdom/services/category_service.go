package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
)

// CategoryService 
type CategoryService struct {
	db    *gorm.DB
	cache *CacheService
}

// NewCategoryService 
func NewCategoryService(db *gorm.DB, cache *CacheService) *CategoryService {
	return &CategoryService{
		db:    db,
		cache: cache,
	}
}

// GetCategories 
func (s *CategoryService) GetCategories(ctx context.Context, parentID *int, includeChildren bool) ([]models.Category, error) {
	var categories []models.Category
	
	query := s.db.WithContext(ctx).Where("is_active = ?", true)
	
	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}
	
	query = query.Order("sort_order ASC, name ASC")
	
	if err := query.Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	
	// 
	if includeChildren {
		for i := range categories {
			_, err := s.GetCategories(ctx, &categories[i].ID, true)
			if err != nil {
				return nil, err
			}
			// ChildrenCategory?
			// categories[i].Children = children
		}
	}
	
	return categories, nil
}

// GetCategoryByID ID
func (s *CategoryService) GetCategoryByID(ctx context.Context, id int) (*models.Category, error) {
	var category models.Category
	
	if err := s.db.WithContext(ctx).Where("id = ? AND is_active = ?", id, true).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	
	return &category, nil
}

// CreateCategory 
func (s *CategoryService) CreateCategory(ctx context.Context, category *models.Category) (*models.Category, error) {
	// ?
	if category.ParentID != nil {
		if _, err := s.GetCategoryByID(ctx, *category.ParentID); err != nil {
			return nil, fmt.Errorf("parent category not found")
		}
	}
	
	// ?
	var count int64
	query := s.db.WithContext(ctx).Model(&models.Category{}).Where("name = ? AND is_active = ?", category.Name, true)
	if category.ParentID != nil {
		query = query.Where("parent_id = ?", *category.ParentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}
	
	if err := query.Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to check category name: %w", err)
	}
	
	if count > 0 {
		return nil, fmt.Errorf("category name already exists in the same level")
	}
	
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()
	
	if err := s.db.WithContext(ctx).Create(category).Error; err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}
	
	// 
	s.cache.DeletePattern(ctx, "categories:*")
	
	return category, nil
}

// UpdateCategory 
func (s *CategoryService) UpdateCategory(ctx context.Context, id int, req interface{}) (*models.Category, error) {
	category, err := s.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// 
	// map
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	
	// req
	// 
	
	if err := s.db.WithContext(ctx).Model(category).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}
	
	// 
	s.cache.DeletePattern(ctx, "categories:*")
	
	return category, nil
}

// DeleteCategory ?
func (s *CategoryService) DeleteCategory(ctx context.Context, id int) error {
	// ?
	var childCount int64
	if err := s.db.WithContext(ctx).Model(&models.Category{}).Where("parent_id = ? AND is_active = ?", id, true).Count(&childCount).Error; err != nil {
		return fmt.Errorf("failed to check child categories: %w", err)
	}
	
	if childCount > 0 {
		return fmt.Errorf("cannot delete category with child categories")
	}
	
	// ?
	var wisdomCount int64
	if err := s.db.WithContext(ctx).Model(&models.CulturalWisdom{}).Where("category = ?", fmt.Sprintf("%d", id)).Count(&wisdomCount).Error; err != nil {
		return fmt.Errorf("failed to check associated wisdom content: %w", err)
	}
	
	if wisdomCount > 0 {
		return fmt.Errorf("cannot delete category with associated wisdom content")
	}
	
	// ?
	if err := s.db.WithContext(ctx).Model(&models.Category{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	
	// 
	s.cache.DeletePattern(ctx, "categories:*")
	
	return nil
}

// GetCategoryStats 
func (s *CategoryService) GetCategoryStats(ctx context.Context, categoryID int) (*models.CategoryStats, error) {
	var stats models.CategoryStats
	
	// ?
	var totalCount int64
	if err := s.db.WithContext(ctx).Model(&models.CulturalWisdom{}).
		Where("category = ?", fmt.Sprintf("%d", categoryID)).
		Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	
	// 
	var publishedCount int64
	if err := s.db.WithContext(ctx).Model(&models.CulturalWisdom{}).
		Where("category = ? AND status = ?", fmt.Sprintf("%d", categoryID), "published").
		Count(&publishedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get published count: %w", err)
	}
	
	// 
	var draftCount int64
	if err := s.db.WithContext(ctx).Model(&models.CulturalWisdom{}).
		Where("category = ? AND status = ?", fmt.Sprintf("%d", categoryID), "draft").
		Count(&draftCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get draft count: %w", err)
	}
	
	// 
	var totalViews int64
	if err := s.db.WithContext(ctx).Model(&models.CulturalWisdom{}).
		Where("category = ?", fmt.Sprintf("%d", categoryID)).
		Select("COALESCE(SUM(view_count), 0)").
		Scan(&totalViews).Error; err != nil {
		return nil, fmt.Errorf("failed to get total views: %w", err)
	}
	
	// 
	var totalLikes int64
	if err := s.db.WithContext(ctx).Model(&models.CulturalWisdom{}).
		Where("category = ?", fmt.Sprintf("%d", categoryID)).
		Select("COALESCE(SUM(like_count), 0)").
		Scan(&totalLikes).Error; err != nil {
		return nil, fmt.Errorf("failed to get total likes: %w", err)
	}
	
	stats = models.CategoryStats{
		CategoryID:     categoryID,
		TotalCount:     totalCount,
		PublishedCount: publishedCount,
		DraftCount:     draftCount,
		TotalViews:     totalViews,
		TotalLikes:     totalLikes,
	}
	
	return &stats, nil
}

// GetCategoryTree ?
func (s *CategoryService) GetCategoryTree(ctx context.Context) ([]models.CategoryNode, error) {
	var categories []models.Category
	
	if err := s.db.WithContext(ctx).Where("is_active = ?", true).Order("sort_order ASC, name ASC").Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	
	// ?
	categoryMap := make(map[int]*models.CategoryNode)
	var roots []models.CategoryNode
	
	// ?
	for _, cat := range categories {
		node := &models.CategoryNode{
			Category: cat,
			Children: []models.CategoryNode{},
		}
		categoryMap[cat.ID] = node
	}
	
	// 
	for _, cat := range categories {
		node := categoryMap[cat.ID]
		if cat.ParentID != nil {
			if parent, exists := categoryMap[*cat.ParentID]; exists {
				parent.Children = append(parent.Children, *node)
			}
		} else {
			roots = append(roots, *node)
		}
	}
	
	return roots, nil
}

// SearchCategories 
func (s *CategoryService) SearchCategories(ctx context.Context, keyword string) ([]models.Category, error) {
	var categories []models.Category
	
	query := s.db.WithContext(ctx).Where("is_active = ?", true)
	
	if keyword != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	
	if err := query.Order("sort_order ASC, name ASC").Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to search categories: %w", err)
	}
	
	return categories, nil
}

