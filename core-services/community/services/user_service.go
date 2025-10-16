﻿package services

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/community/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserService 
type UserService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewUserService 
func NewUserService(db *gorm.DB, logger *zap.Logger) *UserService {
	return &UserService{
		db:     db,
		logger: logger,
	}
}

// CreateOrUpdateUserProfile 
func (s *UserService) CreateOrUpdateUserProfile(userID, username, nickname string) (*models.UserProfile, error) {
	var userProfile models.UserProfile

	// 
	err := s.db.Where("user_id = ?", userID).First(&userProfile).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		s.logger.Error("Failed to query user profile", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("")
	}

	if err == gorm.ErrRecordNotFound {
		// 
		userProfile = models.UserProfile{
			UserID:   userID,
			Username: username,
			Nickname: nickname,
			Status:   models.UserStatusActive,
		}

		if err := s.db.Create(&userProfile).Error; err != nil {
			s.logger.Error("Failed to create user profile", zap.String("user_id", userID), zap.Error(err))
			return nil, fmt.Errorf("")
		}

		s.logger.Info("User profile created", zap.String("user_id", userID))
	} else {
		// 
		updates := map[string]interface{}{
			"username": username,
			"nickname": nickname,
		}
		userProfile.UpdateLastActive()
		updates["last_active_at"] = userProfile.LastActiveAt

		if err := s.db.Model(&userProfile).Updates(updates).Error; err != nil {
			s.logger.Error("Failed to update user profile", zap.String("user_id", userID), zap.Error(err))
			return nil, fmt.Errorf("")
		}

		s.logger.Info("User profile updated", zap.String("user_id", userID))
	}

	return &userProfile, nil
}

// GetUserProfile 
func (s *UserService) GetUserProfile(userID string, viewerID *string) (*models.UserProfile, error) {
	var userProfile models.UserProfile
	if err := s.db.Where("user_id = ?", userID).First(&userProfile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("")
		}
		s.logger.Error("Failed to get user profile", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("")
	}

	return &userProfile, nil
}

// UpdateUserProfile 
func (s *UserService) UpdateUserProfile(userID string, req *models.UserProfileUpdateRequest) (*models.UserProfile, error) {
	var userProfile models.UserProfile
	if err := s.db.Where("user_id = ?", userID).First(&userProfile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("")
		}
		s.logger.Error("Failed to find user profile for update", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	updates := make(map[string]interface{})
	if req.Nickname != nil {
		updates["nickname"] = *req.Nickname
	}
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}
	if req.Bio != nil {
		updates["bio"] = *req.Bio
	}
	if req.Location != nil {
		updates["location"] = *req.Location
	}
	if req.Website != nil {
		updates["website"] = *req.Website
	}

	// 
	userProfile.UpdateLastActive()
	updates["last_active_at"] = userProfile.LastActiveAt

	if len(updates) > 0 {
		if err := s.db.Model(&userProfile).Updates(updates).Error; err != nil {
			s.logger.Error("Failed to update user profile", zap.String("user_id", userID), zap.Error(err))
			return nil, fmt.Errorf("")
		}
	}

	// 
	if err := s.db.First(&userProfile, "user_id = ?", userID).Error; err != nil {
		s.logger.Error("Failed to reload user profile", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("")
	}

	s.logger.Info("User profile updated successfully", zap.String("user_id", userID))
	return &userProfile, nil
}

// GetUsers 
func (s *UserService) GetUsers(req *models.UserListRequest, viewerID *string) (*models.UserListResponse, error) {
	var users []models.UserProfile
	var total int64

	// 
	query := s.db.Model(&models.UserProfile{}).Where("status != ?", models.UserStatusDeleted)

	// 
	if req.Keyword != "" {
		keyword := "%" + strings.ToLower(req.Keyword) + "%"
		query = query.Where("LOWER(username) LIKE ? OR LOWER(nickname) LIKE ?", keyword, keyword)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 
	if err := query.Count(&total).Error; err != nil {
		s.logger.Error("Failed to count users", zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	switch req.SortBy {
	case "posts":
		query = query.Order("post_count DESC")
	case "followers":
		query = query.Order("follower_count DESC")
	case "level":
		query = query.Order("level DESC, experience DESC")
	default: // latest
		query = query.Order("created_at DESC")
	}

	// 
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Find(&users).Error; err != nil {
		s.logger.Error("Failed to get users", zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	userResponses := make([]models.UserProfileResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	return &models.UserListResponse{
		Users:      userResponses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetUserStats 
func (s *UserService) GetUserStats() (*models.UserStatsResponse, error) {
	var stats models.UserStatsResponse

	// 
	s.db.Model(&models.UserProfile{}).Where("status != ?", models.UserStatusDeleted).Count(&stats.TotalUsers)

	// 
	s.db.Model(&models.UserProfile{}).Where("status = ?", models.UserStatusActive).Count(&stats.ActiveUsers)

	// 
	today := time.Now().Truncate(24 * time.Hour)
	s.db.Model(&models.UserProfile{}).Where("created_at >= ?", today).Count(&stats.NewUsers)

	// 
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	s.db.Model(&models.UserProfile{}).Where("created_at >= ?", weekStart).Count(&stats.WeeklyUsers)

	// 
	monthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	s.db.Model(&models.UserProfile{}).Where("created_at >= ?", monthStart).Count(&stats.MonthlyUsers)

	//  1 
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	s.db.Model(&models.UserProfile{}).Where("last_active_at >= ?", oneHourAgo).Count(&stats.OnlineUsers)

	// 
	var topUsers []models.UserProfile
	s.db.Where("status = ?", models.UserStatusActive).
		Order("post_count DESC, follower_count DESC").
		Limit(10).
		Find(&topUsers)

	stats.TopUsers = make([]models.UserProfileBrief, len(topUsers))
	for i, user := range topUsers {
		stats.TopUsers[i] = *user.ToBrief()
	}

	return &stats, nil
}

// SearchUsers 
func (s *UserService) SearchUsers(keyword string, page, pageSize int) (*models.UserListResponse, error) {
	var users []models.UserProfile
	var total int64

	searchTerm := "%" + strings.ToLower(keyword) + "%"
	query := s.db.Model(&models.UserProfile{}).
		Where("status = ? AND (LOWER(username) LIKE ? OR LOWER(nickname) LIKE ?)",
			models.UserStatusActive, searchTerm, searchTerm)

	// 
	if err := query.Count(&total).Error; err != nil {
		s.logger.Error("Failed to count search results", zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	offset := (page - 1) * pageSize
	if err := query.Order("follower_count DESC, post_count DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		s.logger.Error("Failed to search users", zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	userResponses := make([]models.UserProfileResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &models.UserListResponse{
		Users:      userResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateUserActivity 
func (s *UserService) UpdateUserActivity(userID string) error {
	now := time.Now()
	if err := s.db.Model(&models.UserProfile{}).
		Where("user_id = ?", userID).
		UpdateColumn("last_active_at", now).Error; err != nil {
		s.logger.Error("Failed to update user activity", zap.String("user_id", userID), zap.Error(err))
		return fmt.Errorf("")
	}

	return nil
}

// BanUser 
func (s *UserService) BanUser(userID string, adminID string) error {
	if err := s.db.Model(&models.UserProfile{}).
		Where("user_id = ?", userID).
		UpdateColumn("status", models.UserStatusBanned).Error; err != nil {
		s.logger.Error("Failed to ban user", zap.String("user_id", userID), zap.String("admin_id", adminID), zap.Error(err))
		return fmt.Errorf("")
	}

	s.logger.Info("User banned", zap.String("user_id", userID), zap.String("admin_id", adminID))
	return nil
}

// UnbanUser 
func (s *UserService) UnbanUser(userID string, adminID string) error {
	if err := s.db.Model(&models.UserProfile{}).
		Where("user_id = ?", userID).
		UpdateColumn("status", models.UserStatusActive).Error; err != nil {
		s.logger.Error("Failed to unban user", zap.String("user_id", userID), zap.String("admin_id", adminID), zap.Error(err))
		return fmt.Errorf("")
	}

	s.logger.Info("User unbanned", zap.String("user_id", userID), zap.String("admin_id", adminID))
	return nil
}

// GetUserPosts 
func (s *UserService) GetUserPosts(userID string, page, pageSize int, viewerID *string) (*models.PostListResponse, error) {
	var posts []models.Post
	var total int64

	// 
	query := s.db.Model(&models.Post{}).
		Preload("Author").
		Where("author_id = ? AND status = ?", userID, models.PostStatusPublished)

	// 
	if err := query.Count(&total).Error; err != nil {
		s.logger.Error("Failed to count user posts", zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&posts).Error; err != nil {
		s.logger.Error("Failed to get user posts", zap.Error(err))
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

// IsUserActive 
func (s *UserService) IsUserActive(userID string) (bool, error) {
	var count int64
	if err := s.db.Model(&models.UserProfile{}).
		Where("user_id = ? AND status = ?", userID, models.UserStatusActive).
		Count(&count).Error; err != nil {
		s.logger.Error("Failed to check user status", zap.String("user_id", userID), zap.Error(err))
		return false, fmt.Errorf("")
	}

	return count > 0, nil
}

