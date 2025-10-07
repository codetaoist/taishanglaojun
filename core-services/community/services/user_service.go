package services

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/community/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewUserService 创建用户服务实例
func NewUserService(db *gorm.DB, logger *zap.Logger) *UserService {
	return &UserService{
		db:     db,
		logger: logger,
	}
}

// CreateOrUpdateUserProfile 创建或更新用户资料
func (s *UserService) CreateOrUpdateUserProfile(userID, username, nickname string) (*models.UserProfile, error) {
	var userProfile models.UserProfile
	
	// 尝试查找现有用户资料
	err := s.db.Where("user_id = ?", userID).First(&userProfile).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		s.logger.Error("Failed to query user profile", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("查询用户资料失败")
	}

	if err == gorm.ErrRecordNotFound {
		// 创建新用户资料
		userProfile = models.UserProfile{
			UserID:   userID,
			Username: username,
			Nickname: nickname,
			Status:   models.UserStatusActive,
		}

		if err := s.db.Create(&userProfile).Error; err != nil {
			s.logger.Error("Failed to create user profile", zap.String("user_id", userID), zap.Error(err))
			return nil, fmt.Errorf("创建用户资料失败")
		}

		s.logger.Info("User profile created", zap.String("user_id", userID))
	} else {
		// 更新现有用户资料
		updates := map[string]interface{}{
			"username": username,
			"nickname": nickname,
		}
		userProfile.UpdateLastActive()
		updates["last_active_at"] = userProfile.LastActiveAt

		if err := s.db.Model(&userProfile).Updates(updates).Error; err != nil {
			s.logger.Error("Failed to update user profile", zap.String("user_id", userID), zap.Error(err))
			return nil, fmt.Errorf("更新用户资料失败")
		}

		s.logger.Info("User profile updated", zap.String("user_id", userID))
	}

	return &userProfile, nil
}

// GetUserProfile 获取用户资料
func (s *UserService) GetUserProfile(userID string, viewerID *string) (*models.UserProfile, error) {
	var userProfile models.UserProfile
	if err := s.db.Where("user_id = ?", userID).First(&userProfile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		s.logger.Error("Failed to get user profile", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("获取用户资料失败")
	}

	return &userProfile, nil
}

// UpdateUserProfile 更新用户资料
func (s *UserService) UpdateUserProfile(userID string, req *models.UserProfileUpdateRequest) (*models.UserProfile, error) {
	var userProfile models.UserProfile
	if err := s.db.Where("user_id = ?", userID).First(&userProfile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		s.logger.Error("Failed to find user profile for update", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("查找用户资料失败")
	}

	// 更新字段
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

	// 更新最后活跃时间
	userProfile.UpdateLastActive()
	updates["last_active_at"] = userProfile.LastActiveAt

	if len(updates) > 0 {
		if err := s.db.Model(&userProfile).Updates(updates).Error; err != nil {
			s.logger.Error("Failed to update user profile", zap.String("user_id", userID), zap.Error(err))
			return nil, fmt.Errorf("更新用户资料失败")
		}
	}

	// 重新加载用户资料
	if err := s.db.First(&userProfile, "user_id = ?", userID).Error; err != nil {
		s.logger.Error("Failed to reload user profile", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("重新加载用户资料失败")
	}

	s.logger.Info("User profile updated successfully", zap.String("user_id", userID))
	return &userProfile, nil
}

// GetUsers 获取用户列表
func (s *UserService) GetUsers(req *models.UserListRequest, viewerID *string) (*models.UserListResponse, error) {
	var users []models.UserProfile
	var total int64

	// 构建查询
	query := s.db.Model(&models.UserProfile{}).Where("status != ?", models.UserStatusDeleted)

	// 添加筛选条件
	if req.Keyword != "" {
		keyword := "%" + strings.ToLower(req.Keyword) + "%"
		query = query.Where("LOWER(username) LIKE ? OR LOWER(nickname) LIKE ?", keyword, keyword)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		s.logger.Error("Failed to count users", zap.Error(err))
		return nil, fmt.Errorf("获取用户数量失败")
	}

	// 添加排序
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

	// 分页
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Find(&users).Error; err != nil {
		s.logger.Error("Failed to get users", zap.Error(err))
		return nil, fmt.Errorf("获取用户列表失败")
	}

	// 转换为响应格式
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

// GetUserStats 获取用户统计
func (s *UserService) GetUserStats() (*models.UserStatsResponse, error) {
	var stats models.UserStatsResponse

	// 总用户数
	s.db.Model(&models.UserProfile{}).Where("status != ?", models.UserStatusDeleted).Count(&stats.TotalUsers)

	// 活跃用户数
	s.db.Model(&models.UserProfile{}).Where("status = ?", models.UserStatusActive).Count(&stats.ActiveUsers)

	// 今日新用户
	today := time.Now().Truncate(24 * time.Hour)
	s.db.Model(&models.UserProfile{}).Where("created_at >= ?", today).Count(&stats.NewUsers)

	// 本周新用户
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	s.db.Model(&models.UserProfile{}).Where("created_at >= ?", weekStart).Count(&stats.WeeklyUsers)

	// 本月新用户
	monthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	s.db.Model(&models.UserProfile{}).Where("created_at >= ?", monthStart).Count(&stats.MonthlyUsers)

	// 在线用户（最近1小时活跃）
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	s.db.Model(&models.UserProfile{}).Where("last_active_at >= ?", oneHourAgo).Count(&stats.OnlineUsers)

	// 活跃用户排行
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

// SearchUsers 搜索用户
func (s *UserService) SearchUsers(keyword string, page, pageSize int) (*models.UserListResponse, error) {
	var users []models.UserProfile
	var total int64

	searchTerm := "%" + strings.ToLower(keyword) + "%"
	query := s.db.Model(&models.UserProfile{}).
		Where("status = ? AND (LOWER(username) LIKE ? OR LOWER(nickname) LIKE ?)", 
			models.UserStatusActive, searchTerm, searchTerm)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		s.logger.Error("Failed to count search results", zap.Error(err))
		return nil, fmt.Errorf("搜索失败")
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("follower_count DESC, post_count DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		s.logger.Error("Failed to search users", zap.Error(err))
		return nil, fmt.Errorf("搜索失败")
	}

	// 转换为响应格式
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

// UpdateUserActivity 更新用户活跃状态
func (s *UserService) UpdateUserActivity(userID string) error {
	now := time.Now()
	if err := s.db.Model(&models.UserProfile{}).
		Where("user_id = ?", userID).
		UpdateColumn("last_active_at", now).Error; err != nil {
		s.logger.Error("Failed to update user activity", zap.String("user_id", userID), zap.Error(err))
		return fmt.Errorf("更新用户活跃状态失败")
	}

	return nil
}

// BanUser 封禁用户
func (s *UserService) BanUser(userID string, adminID string) error {
	if err := s.db.Model(&models.UserProfile{}).
		Where("user_id = ?", userID).
		UpdateColumn("status", models.UserStatusBanned).Error; err != nil {
		s.logger.Error("Failed to ban user", zap.String("user_id", userID), zap.String("admin_id", adminID), zap.Error(err))
		return fmt.Errorf("封禁用户失败")
	}

	s.logger.Info("User banned", zap.String("user_id", userID), zap.String("admin_id", adminID))
	return nil
}

// UnbanUser 解封用户
func (s *UserService) UnbanUser(userID string, adminID string) error {
	if err := s.db.Model(&models.UserProfile{}).
		Where("user_id = ?", userID).
		UpdateColumn("status", models.UserStatusActive).Error; err != nil {
		s.logger.Error("Failed to unban user", zap.String("user_id", userID), zap.String("admin_id", adminID), zap.Error(err))
		return fmt.Errorf("解封用户失败")
	}

	s.logger.Info("User unbanned", zap.String("user_id", userID), zap.String("admin_id", adminID))
	return nil
}

// GetUserPosts 获取用户的帖子列表
func (s *UserService) GetUserPosts(userID string, page, pageSize int, viewerID *string) (*models.PostListResponse, error) {
	var posts []models.Post
	var total int64

	// 构建查询
	query := s.db.Model(&models.Post{}).
		Preload("Author").
		Where("author_id = ? AND status = ?", userID, models.PostStatusPublished)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		s.logger.Error("Failed to count user posts", zap.Error(err))
		return nil, fmt.Errorf("获取用户帖子数量失败")
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&posts).Error; err != nil {
		s.logger.Error("Failed to get user posts", zap.Error(err))
		return nil, fmt.Errorf("获取用户帖子列表失败")
	}

	// 转换为响应格式
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

// IsUserActive 检查用户是否活跃
func (s *UserService) IsUserActive(userID string) (bool, error) {
	var count int64
	if err := s.db.Model(&models.UserProfile{}).
		Where("user_id = ? AND status = ?", userID, models.UserStatusActive).
		Count(&count).Error; err != nil {
		s.logger.Error("Failed to check user status", zap.String("user_id", userID), zap.Error(err))
		return false, fmt.Errorf("检查用户状态失败")
	}

	return count > 0, nil
}