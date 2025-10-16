package services

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/community/models"
	"github.com/codetaoist/taishanglaojun/core-services/community/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PostService 
type PostService struct {
	db               *gorm.DB
	logger           *zap.Logger
	contentValidator *utils.ContentValidator
}

// NewPostService 
func NewPostService(db *gorm.DB, logger *zap.Logger) *PostService {
	return &PostService{
		db:               db,
		logger:           logger,
		contentValidator: utils.NewContentValidator(),
	}
}

// CreatePost 
func (s *PostService) CreatePost(userID string, req *models.PostCreateRequest) (*models.Post, error) {
	// 
	var userProfile models.UserProfile
	if err := s.db.Where("user_id = ? AND status = ?", userID, models.UserStatusActive).First(&userProfile).Error; err != nil {
		s.logger.Error("User not found or inactive", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	validationResult := s.contentValidator.ValidatePostContent(req.Title, req.Content)
	if !validationResult.IsValid {
		s.logger.Warn("Post content validation failed",
			zap.String("user_id", userID),
			zap.Strings("errors", validationResult.Errors),
			zap.Int("risk_level", validationResult.RiskLevel))
		return nil, fmt.Errorf(": %s", validationResult.Errors[0])
	}

	// 
	postStatus := models.PostStatusPublished
	if validationResult.RiskLevel > 0 {
		postStatus = models.PostStatusPending // 
		s.logger.Info("Post requires review due to risk level",
			zap.String("user_id", userID),
			zap.Int("risk_level", validationResult.RiskLevel),
			zap.Strings("warnings", validationResult.Warnings))
	}

	// JSON
	tagsJSON := ""
	if len(req.Tags) > 0 {
		tagsBytes, _ := json.Marshal(req.Tags)
		tagsJSON = string(tagsBytes)
	}

	// 
	post := &models.Post{
		ID:       uuid.New().String(),
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: userID,
		Category: req.Category,
		Tags:     tagsJSON,
		Status:   postStatus,
	}

	// 
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 
	if err := tx.Create(post).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to create post", zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	if err := tx.Model(&userProfile).UpdateColumn("post_count", gorm.Expr("post_count + ?", 1)).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to update user post count", zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	userProfile.AddExperience(10) // 10
	if err := tx.Save(&userProfile).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to update user experience", zap.Error(err))
		return nil, fmt.Errorf("")
	}

	tx.Commit()

	// 
	post.Author = &userProfile

	s.logger.Info("Post created successfully", zap.String("post_id", post.ID), zap.String("user_id", userID))
	return post, nil
}

// GetPost 
func (s *PostService) GetPost(postID string, userID *string) (*models.Post, error) {
	var post models.Post
	query := s.db.Preload("Author").Preload("Comments.Author").Where("id = ? AND status != ?", postID, models.PostStatusDeleted)

	if err := query.First(&post).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("")
		}
		s.logger.Error("Failed to get post", zap.String("post_id", postID), zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	go func() {
		s.db.Model(&post).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))
	}()

	return &post, nil
}

// GetPosts 
func (s *PostService) GetPosts(req *models.PostListRequest, userID *string) (*models.PostListResponse, error) {
	var posts []models.Post
	var total int64

	// 
	query := s.db.Model(&models.Post{}).Preload("Author").Where("status = ?", models.PostStatusPublished)

	// 
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}

	if req.AuthorID != "" {
		query = query.Where("author_id = ?", req.AuthorID)
	}

	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		query = query.Where("title LIKE ? OR content LIKE ?", keyword, keyword)
	}

	if req.Tag != "" {
		query = query.Where("tags LIKE ?", "%"+req.Tag+"%")
	}

	// 
	if err := query.Count(&total).Error; err != nil {
		s.logger.Error("Failed to count posts", zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	switch req.SortBy {
	case "hot":
		query = query.Order("is_hot DESC, like_count DESC, comment_count DESC, view_count DESC")
	case "likes":
		query = query.Order("like_count DESC")
	case "views":
		query = query.Order("view_count DESC")
	default: // latest
		query = query.Order("is_sticky DESC, created_at DESC")
	}

	// 
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Find(&posts).Error; err != nil {
		s.logger.Error("Failed to get posts", zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	postResponses := make([]models.PostResponse, len(posts))
	for i, post := range posts {
		postResponses[i] = post.ToResponse()
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	return &models.PostListResponse{
		Posts:      postResponses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdatePost 
func (s *PostService) UpdatePost(postID, userID string, req *models.PostUpdateRequest) (*models.Post, error) {
	var post models.Post
	if err := s.db.Where("id = ? AND author_id = ?", postID, userID).First(&post).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("")
		}
		s.logger.Error("Failed to find post for update", zap.String("post_id", postID), zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Tags != nil {
		tagsBytes, _ := json.Marshal(req.Tags)
		updates["tags"] = string(tagsBytes)
	}

	if len(updates) > 0 {
		if err := s.db.Model(&post).Updates(updates).Error; err != nil {
			s.logger.Error("Failed to update post", zap.String("post_id", postID), zap.Error(err))
			return nil, fmt.Errorf("")
		}
	}

	// 
	if err := s.db.Preload("Author").First(&post, "id = ?", postID).Error; err != nil {
		s.logger.Error("Failed to reload post", zap.String("post_id", postID), zap.Error(err))
		return nil, fmt.Errorf("")
	}

	s.logger.Info("Post updated successfully", zap.String("post_id", postID), zap.String("user_id", userID))
	return &post, nil
}

// DeletePost 
func (s *PostService) DeletePost(postID, userID string) error {
	var post models.Post
	if err := s.db.Where("id = ? AND author_id = ?", postID, userID).First(&post).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("")
		}
		s.logger.Error("Failed to find post for deletion", zap.String("post_id", postID), zap.Error(err))
		return fmt.Errorf("")
	}

	// 
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 
	if err := tx.Delete(&post).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to delete post", zap.String("post_id", postID), zap.Error(err))
		return fmt.Errorf("")
	}

	// 
	if err := tx.Model(&models.UserProfile{}).Where("user_id = ?", userID).UpdateColumn("post_count", gorm.Expr("post_count - ?", 1)).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to update user post count", zap.Error(err))
		return fmt.Errorf("")
	}

	tx.Commit()

	s.logger.Info("Post deleted successfully", zap.String("post_id", postID), zap.String("user_id", userID))
	return nil
}

// GetPostStats 
func (s *PostService) GetPostStats() (*models.PostStatsResponse, error) {
	var stats models.PostStatsResponse

	// 
	s.db.Model(&models.Post{}).Where("status = ?", models.PostStatusPublished).Count(&stats.TotalPosts)

	// 
	today := time.Now().Truncate(24 * time.Hour)
	s.db.Model(&models.Post{}).Where("status = ? AND created_at >= ?", models.PostStatusPublished, today).Count(&stats.TodayPosts)

	// 
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	s.db.Model(&models.Post{}).Where("status = ? AND created_at >= ?", models.PostStatusPublished, weekStart).Count(&stats.WeeklyPosts)

	// 
	monthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	s.db.Model(&models.Post{}).Where("status = ? AND created_at >= ?", models.PostStatusPublished, monthStart).Count(&stats.MonthlyPosts)

	// 
	s.db.Model(&models.Post{}).Where("status = ?", models.PostStatusPublished).Select("COALESCE(SUM(view_count), 0)").Scan(&stats.TotalViews)
	s.db.Model(&models.Post{}).Where("status = ?", models.PostStatusPublished).Select("COALESCE(SUM(like_count), 0)").Scan(&stats.TotalLikes)
	s.db.Model(&models.Post{}).Where("status = ?", models.PostStatusPublished).Select("COALESCE(SUM(comment_count), 0)").Scan(&stats.TotalComments)

	// 
	s.db.Model(&models.Post{}).Where("status = ? AND created_at >= ?", models.PostStatusPublished, weekStart).Distinct("author_id").Count(&stats.ActiveUsers)

	// 
	stats.PopularTags = []models.TagStats{
		{Tag: "", Count: 50},
		{Tag: "", Count: 30},
		{Tag: "", Count: 25},
	}

	// 
	var categoryStats []models.CategoryStats
	s.db.Model(&models.Post{}).
		Select("category, COUNT(*) as count").
		Where("status = ?", models.PostStatusPublished).
		Group("category").
		Order("count DESC").
		Limit(10).
		Scan(&categoryStats)
	stats.TopCategories = categoryStats

	return &stats, nil
}

// SetPostSticky 
func (s *PostService) SetPostSticky(postID string, sticky bool) error {
	if err := s.db.Model(&models.Post{}).Where("id = ?", postID).UpdateColumn("is_sticky", sticky).Error; err != nil {
		s.logger.Error("Failed to set post sticky", zap.String("post_id", postID), zap.Bool("sticky", sticky), zap.Error(err))
		return fmt.Errorf("")
	}

	s.logger.Info("Post sticky updated", zap.String("post_id", postID), zap.Bool("sticky", sticky))
	return nil
}

// SetPostHot 
func (s *PostService) SetPostHot(postID string, hot bool) error {
	if err := s.db.Model(&models.Post{}).Where("id = ?", postID).UpdateColumn("is_hot", hot).Error; err != nil {
		s.logger.Error("Failed to set post hot", zap.String("post_id", postID), zap.Bool("hot", hot), zap.Error(err))
		return fmt.Errorf("")
	}

	s.logger.Info("Post hot updated", zap.String("post_id", postID), zap.Bool("hot", hot))
	return nil
}

// SearchPosts 
func (s *PostService) SearchPosts(keyword string, page, pageSize int) (*models.PostListResponse, error) {
	var posts []models.Post
	var total int64

	searchTerm := "%" + strings.ToLower(keyword) + "%"
	query := s.db.Model(&models.Post{}).
		Preload("Author").
		Where("status = ? AND (LOWER(title) LIKE ? OR LOWER(content) LIKE ?)",
			models.PostStatusPublished, searchTerm, searchTerm)

	// 
	if err := query.Count(&total).Error; err != nil {
		s.logger.Error("Failed to count search results", zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&posts).Error; err != nil {
		s.logger.Error("Failed to search posts", zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	postResponses := make([]models.PostResponse, len(posts))
	for i, post := range posts {
		postResponses[i] = post.ToResponse()
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &models.PostListResponse{
		Posts:      postResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

