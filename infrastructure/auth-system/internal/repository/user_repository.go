package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/models"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// UserRepository 用户仓储接口
type UserRepository interface {
	// 基础CRUD
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	
	// 查询方法
	List(ctx context.Context, query *models.UserQuery) ([]*models.User, int64, error)
	Search(ctx context.Context, keyword string, limit int) ([]*models.User, error)
	Exists(ctx context.Context, username, email string) (bool, error)
	
	// 认证相关
	Authenticate(ctx context.Context, username, password string) (*models.User, error)
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	ChangePassword(ctx context.Context, userID uuid.UUID, newPassword string) error
	
	// 状态管理器�?
	UpdateStatus(ctx context.Context, userID uuid.UUID, status models.UserStatus) error
	GetActiveUsers(ctx context.Context, limit int) ([]*models.User, error)
	
	// 批量操作
	BatchCreate(ctx context.Context, users []*models.User) error
	BatchUpdateStatus(ctx context.Context, userIDs []uuid.UUID, status models.UserStatus) error
	
	// 统计
	Count(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status models.UserStatus) (int64, error)
	CountByRole(ctx context.Context, role models.UserRole) (int64, error)
}

// userRepository 用户仓储实现
type userRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *gorm.DB, logger *zap.Logger) UserRepository {
	return &userRepository{
		db:     db,
		logger: logger,
	}
}

// Create 创建用户
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	// 检查用户名和邮箱件箱是否已存在
	exists, err := r.Exists(ctx, user.Username, user.Email)
	if err != nil {
		return err
	}
	if exists {
		return ErrUserAlreadyExists
	}
	
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		r.logger.Error("Failed to create user", 
			zap.String("username", user.Username),
			zap.String("email", user.Email),
			zap.Error(err),
		)
		return err
	}
	
	r.logger.Info("User created successfully", 
		zap.String("user_id", user.ID.String()),
		zap.String("username", user.Username),
	)
	
	return nil
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		r.logger.Error("Failed to get user by ID", 
			zap.String("user_id", id.String()),
			zap.Error(err),
		)
		return nil, err
	}
	
	return &user, nil
}

// GetByUsername 根据用户名获取用户�?
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, "username = ?", username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		r.logger.Error("Failed to get user by username", 
			zap.String("username", username),
			zap.Error(err),
		)
		return nil, err
	}
	
	return &user, nil
}

// GetByEmail 根据邮箱件箱获取用户
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		r.logger.Error("Failed to get user by email", 
			zap.String("email", email),
			zap.Error(err),
		)
		return nil, err
	}
	
	return &user, nil
}

// Update 更新用户
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		r.logger.Error("Failed to update user", 
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
		return err
	}
	
	r.logger.Info("User updated successfully", 
		zap.String("user_id", user.ID.String()),
	)
	
	return nil
}

// Delete 删除用户
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Unscoped().Delete(&models.User{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("Failed to delete user", 
			zap.String("user_id", id.String()),
			zap.Error(result.Error),
		)
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	
	r.logger.Info("User deleted successfully", 
		zap.String("user_id", id.String()),
	)
	
	return nil
}

// SoftDelete 软删除用户�?
func (r *userRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("Failed to soft delete user", 
			zap.String("user_id", id.String()),
			zap.Error(result.Error),
		)
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	
	r.logger.Info("User soft deleted successfully", 
		zap.String("user_id", id.String()),
	)
	
	return nil
}

// List 获取用户列表
func (r *userRepository) List(ctx context.Context, query *models.UserQuery) ([]*models.User, int64, error) {
	db := r.db.WithContext(ctx).Model(&models.User{})
	
	// 应用户过期滤条件
	if query.Username != "" {
		db = db.Where("username ILIKE ?", "%"+query.Username+"%")
	}
	if query.Email != "" {
		db = db.Where("email ILIKE ?", "%"+query.Email+"%")
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Role != "" {
		db = db.Where("role = ?", query.Role)
	}
	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		db = db.Where("username ILIKE ? OR email ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?", 
			searchPattern, searchPattern, searchPattern, searchPattern)
	}
	
	// 获取总数量
	var total int64
	if err := db.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count users", zap.Error(err))
		return nil, 0, err
	}
	
	// 应用户排序
	orderBy := "created_at"
	if query.OrderBy != "" {
		orderBy = query.OrderBy
	}
	order := "desc"
	if query.Order != "" {
		order = query.Order
	}
	db = db.Order(fmt.Sprintf("%s %s", orderBy, order))
	
	// 应用户分页
	if query.Page > 0 && query.PageSize > 0 {
		offset := (query.Page - 1) * query.PageSize
		db = db.Offset(offset).Limit(query.PageSize)
	}
	
	var users []*models.User
	if err := db.Find(&users).Error; err != nil {
		r.logger.Error("Failed to list users", zap.Error(err))
		return nil, 0, err
	}
	
	return users, total, nil
}

// Search 搜索用户
func (r *userRepository) Search(ctx context.Context, keyword string, limit int) ([]*models.User, error) {
	var users []*models.User
	searchPattern := "%" + keyword + "%"
	
	query := r.db.WithContext(ctx).
		Where("username ILIKE ? OR email ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?", 
			searchPattern, searchPattern, searchPattern, searchPattern).
		Where("status = ?", models.UserStatusActive).
		Order("username ASC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&users).Error; err != nil {
		r.logger.Error("Failed to search users", 
			zap.String("keyword", keyword),
			zap.Error(err),
		)
		return nil, err
	}
	
	return users, nil
}

// Exists 检查用户是否存�?
func (r *userRepository) Exists(ctx context.Context, username, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).
		Where("username = ? OR email = ?", username, email).
		Count(&count).Error; err != nil {
		r.logger.Error("Failed to check user existence", 
			zap.String("username", username),
			zap.String("email", email),
			zap.Error(err),
		)
		return false, err
	}
	
	return count > 0, nil
}

// Authenticate 用户认证
func (r *userRepository) Authenticate(ctx context.Context, username, password string) (*models.User, error) {
	var user models.User
	
	// 支持用户名或邮箱件箱登录
	if err := r.db.WithContext(ctx).
		Where("(username = ? OR email = ?) AND status = ?", username, username, models.UserStatusActive).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		r.logger.Error("Failed to authenticate user", 
			zap.String("username", username),
			zap.Error(err),
		)
		return nil, err
	}
	
	// 验证密码
	if !user.CheckPassword(password) {
		r.logger.Warn("Invalid password attempt", 
			zap.String("username", username),
			zap.String("user_id", user.ID.String()),
		)
		return nil, ErrInvalidCredentials
	}
	
	return &user, nil
}

// UpdateLastLogin 更新最后登录时�?
func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Update("last_login_at", gorm.Expr("NOW()")).Error; err != nil {
		r.logger.Error("Failed to update last login", 
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return err
	}
	
	return nil
}

// ChangePassword 修改密码
func (r *userRepository) ChangePassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	// 创建临时用户对象来利用户BeforeUpdate钩子
	user := &models.User{
		ID:       userID,
		Password: newPassword,
	}
	
	if err := user.HashPassword(); err != nil {
		return err
	}
	
	if err := r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Update("password", user.Password).Error; err != nil {
		r.logger.Error("Failed to change password", 
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return err
	}
	
	r.logger.Info("Password changed successfully", 
		zap.String("user_id", userID.String()),
	)
	
	return nil
}

// UpdateStatus 更新用户状态枚举�?
func (r *userRepository) UpdateStatus(ctx context.Context, userID uuid.UUID, status models.UserStatus) error {
	result := r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Update("status", status)
	
	if result.Error != nil {
		r.logger.Error("Failed to update user status", 
			zap.String("user_id", userID.String()),
			zap.String("status", string(status)),
			zap.Error(result.Error),
		)
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	
	r.logger.Info("User status updated successfully", 
		zap.String("user_id", userID.String()),
		zap.String("status", string(status)),
	)
	
	return nil
}

// GetActiveUsers 获取活跃用户
func (r *userRepository) GetActiveUsers(ctx context.Context, limit int) ([]*models.User, error) {
	var users []*models.User
	query := r.db.WithContext(ctx).
		Where("status = ?", models.UserStatusActive).
		Order("last_login_at DESC NULLS LAST")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&users).Error; err != nil {
		r.logger.Error("Failed to get active users", zap.Error(err))
		return nil, err
	}
	
	return users, nil
}

// BatchCreate 批量创建用户
func (r *userRepository) BatchCreate(ctx context.Context, users []*models.User) error {
	if len(users) == 0 {
		return nil
	}
	
	if err := r.db.WithContext(ctx).CreateInBatches(users, 100).Error; err != nil {
		r.logger.Error("Failed to batch create users", 
			zap.Int("count", len(users)),
			zap.Error(err),
		)
		return err
	}
	
	r.logger.Info("Users batch created successfully", 
		zap.Int("count", len(users)),
	)
	
	return nil
}

// BatchUpdateStatus 批量更新用户状态枚举�?
func (r *userRepository) BatchUpdateStatus(ctx context.Context, userIDs []uuid.UUID, status models.UserStatus) error {
	if len(userIDs) == 0 {
		return nil
	}
	
	result := r.db.WithContext(ctx).Model(&models.User{}).
		Where("id IN ?", userIDs).
		Update("status", status)
	
	if result.Error != nil {
		r.logger.Error("Failed to batch update user status", 
			zap.Int("count", len(userIDs)),
			zap.String("status", string(status)),
			zap.Error(result.Error),
		)
		return result.Error
	}
	
	r.logger.Info("User status batch updated successfully", 
		zap.Int("count", len(userIDs)),
		zap.Int64("affected", result.RowsAffected),
		zap.String("status", string(status)),
	)
	
	return nil
}

// Count 获取用户总数量
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error; err != nil {
		r.logger.Error("Failed to count users", zap.Error(err))
		return 0, err
	}
	
	return count, nil
}

// CountByStatus 根据状态统计用户数量
func (r *userRepository) CountByStatus(ctx context.Context, status models.UserStatus) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).
		Where("status = ?", status).
		Count(&count).Error; err != nil {
		r.logger.Error("Failed to count users by status", 
			zap.String("status", string(status)),
			zap.Error(err),
		)
		return 0, err
	}
	
	return count, nil
}

// CountByRole 根据角色统计用户�?
func (r *userRepository) CountByRole(ctx context.Context, role models.UserRole) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).
		Where("role = ?", role).
		Count(&count).Error; err != nil {
		r.logger.Error("Failed to count users by role", 
			zap.String("role", string(role)),
			zap.Error(err),
		)
		return 0, err
	}
	
	return count, nil
}
