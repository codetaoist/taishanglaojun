package services

import (
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/internal/models"
)

// UserService 用户服务
type UserService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewUserService 创建新的用户服务
func NewUserService(db *gorm.DB, logger *zap.Logger) *UserService {
	return &UserService{
		db:     db,
		logger: logger,
	}
}

// GetTotalUsers 获取用户总数
func (s *UserService) GetTotalUsers() (int64, error) {
	var count int64
	err := s.db.Model(&models.User{}).Count(&count).Error
	if err != nil {
		s.logger.Error("Failed to get total users count", zap.Error(err))
		return 0, err
	}
	return count, nil
}

// GetActiveUsers 获取活跃用户数（最近30天登录的用户）
func (s *UserService) GetActiveUsers() (int64, error) {
	var count int64
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	err := s.db.Model(&models.User{}).
		Where("last_login_at > ?", thirtyDaysAgo).
		Count(&count).Error
	if err != nil {
		s.logger.Error("Failed to get active users count", zap.Error(err))
		return 0, err
	}
	return count, nil
}

// GetRecentUsers 获取最近登录的用户
func (s *UserService) GetRecentUsers(limit int) ([]models.User, error) {
	var users []models.User
	err := s.db.Where("last_login_at IS NOT NULL").
		Order("last_login_at DESC").
		Limit(limit).
		Find(&users).Error
	if err != nil {
		s.logger.Error("Failed to get recent users", zap.Error(err))
		return nil, err
	}
	return users, nil
}

// GetUserCreationTrend 获取用户创建趋势数据
func (s *UserService) GetUserCreationTrend(days int) (map[string]int, error) {
	type DailyCount struct {
		Date  string `json:"date"`
		Count int    `json:"count"`
	}

	var results []DailyCount
	startDate := time.Now().AddDate(0, 0, -days)
	
	err := s.db.Model(&models.User{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ?", startDate).
		Group("DATE(created_at)").
		Order("date").
		Scan(&results).Error
	
	if err != nil {
		s.logger.Error("Failed to get user creation trend", zap.Error(err))
		return nil, err
	}

	// 转换为map格式
	trendData := make(map[string]int)
	for _, result := range results {
		trendData[result.Date] = result.Count
	}

	return trendData, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := s.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		s.logger.Error("Failed to get user by ID", zap.String("id", id.String()), zap.Error(err))
		return nil, err
	}
	return &user, nil
}

// CreateUser 创建用户
func (s *UserService) CreateUser(user *models.User) error {
	err := s.db.Create(user).Error
	if err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return err
	}
	return nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(user *models.User) error {
	err := s.db.Save(user).Error
	if err != nil {
		s.logger.Error("Failed to update user", zap.Error(err))
		return err
	}
	return nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id uuid.UUID) error {
	err := s.db.Delete(&models.User{}, id).Error
	if err != nil {
		s.logger.Error("Failed to delete user", zap.String("id", id.String()), zap.Error(err))
		return err
	}
	return nil
}

// GetAdminUsers 获取管理员用户数量
func (s *UserService) GetAdminUsers() (int64, error) {
	var count int64
	err := s.db.Model(&models.User{}).
		Where("role IN (?)", []string{"admin", "super_admin"}).
		Count(&count).Error
	if err != nil {
		s.logger.Error("Failed to get admin users count", zap.Error(err))
		return 0, err
	}
	return count, nil
}

// GetNewUsersToday 获取今日新增用户数
func (s *UserService) GetNewUsersToday() (int64, error) {
	var count int64
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	
	err := s.db.Model(&models.User{}).
		Where("created_at >= ? AND created_at < ?", today, tomorrow).
		Count(&count).Error
	if err != nil {
		s.logger.Error("Failed to get new users today count", zap.Error(err))
		return 0, err
	}
	return count, nil
}

// GetOnlineUsers 获取在线用户数（最近15分钟活跃的用户）
func (s *UserService) GetOnlineUsers() (int64, error) {
	var count int64
	fifteenMinutesAgo := time.Now().Add(-15 * time.Minute)
	
	err := s.db.Model(&models.User{}).
		Where("last_login_at > ?", fifteenMinutesAgo).
		Count(&count).Error
	if err != nil {
		s.logger.Error("Failed to get online users count", zap.Error(err))
		return 0, err
	}
	return count, nil
}

// GetUserStats 获取用户统计信息
func (s *UserService) GetUserStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 获取总用户数
	totalUsers, err := s.GetTotalUsers()
	if err != nil {
		return nil, err
	}
	stats["totalUsers"] = totalUsers
	
	// 获取活跃用户数
	activeUsers, err := s.GetActiveUsers()
	if err != nil {
		return nil, err
	}
	stats["activeUsers"] = activeUsers
	
	// 获取管理员数量
	adminUsers, err := s.GetAdminUsers()
	if err != nil {
		return nil, err
	}
	stats["adminUsers"] = adminUsers
	
	// 获取今日新增用户数
	newUsersToday, err := s.GetNewUsersToday()
	if err != nil {
		return nil, err
	}
	stats["newUsersToday"] = newUsersToday
	
	// 获取在线用户数
	onlineUsers, err := s.GetOnlineUsers()
	if err != nil {
		return nil, err
	}
	stats["onlineUsers"] = onlineUsers
	
	return stats, nil
}