package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// User 用户模型 - 与auth-system保持一致
type User struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	Username  string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Email     string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"`
	Level     int       `gorm:"type:int;default:1" json:"level"` // 用户等级，默认为1
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate 在创建前生成UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// AuthService 认证服务
type AuthService struct {
	db         *gorm.DB
	jwtMiddleware *JWTMiddleware
	logger     *zap.Logger
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// NewAuthService 创建认证服务
func NewAuthService(db *gorm.DB, jwtMiddleware *JWTMiddleware, logger *zap.Logger) *AuthService {
	return &AuthService{
		db:         db,
		jwtMiddleware: jwtMiddleware,
		logger:     logger,
	}
}

// AutoMigrate 自动迁移用户表
func (s *AuthService) AutoMigrate() error {
	return s.db.AutoMigrate(&User{})
}

// Register 用户注册
func (s *AuthService) Register(req *RegisterRequest) (*AuthResponse, error) {
	// 检查用户名是否已存在
	var existingUser User
	if err := s.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("用户名或邮箱已存在")
	}

	// 生成盐值
	// salt, err := generateSalt()
	// if err != nil {
	//	return nil, fmt.Errorf("生成盐值失败: %w", err)
	// }

	// 哈希密码
	hashedPassword := hashPassword(req.Password)

	// 创建用户
	user := &User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := s.db.Create(user).Error; err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 生成JWT令牌
	token, err := s.jwtMiddleware.GenerateToken(user.ID.String(), user.Username, user.Level)
	if err != nil {
		s.logger.Error("Failed to generate token", zap.Error(err))
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}

	return &AuthResponse{
		Token:    token,
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
	}, nil
}

// Login 用户登录
func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
	// 查找用户
	var user User
	if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户名或密码错误")
		}
		s.logger.Error("Failed to find user", zap.Error(err))
		return nil, fmt.Errorf("查找用户失败: %w", err)
	}

	// 验证密码
	if !verifyPassword(req.Password, user.Password) {
		return nil, fmt.Errorf("用户名或密码错误")
	}

	// 生成JWT令牌
	token, err := s.jwtMiddleware.GenerateToken(user.ID.String(), user.Username, user.Level)
	if err != nil {
		s.logger.Error("Failed to generate token", zap.Error(err))
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}

	return &AuthResponse{
		Token:    token,
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
	}, nil
}

// GetUserByID 根据ID获取用户信息
func (s *AuthService) GetUserByID(userID string) (*User, error) {
	var user User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查找用户失败: %w", err)
	}
	return &user, nil
}

// CreateTestUser 创建测试用户（仅用于开发和测试）
func (s *AuthService) CreateTestUser() (*AuthResponse, error) {
	// 检查是否已存在测试用户
	var existingUser User
	if err := s.db.Where("username = ?", "testuser").First(&existingUser).Error; err == nil {
		// 用户已存在，更新等级并生成令牌
		existingUser.Level = 5
		if err := s.db.Save(&existingUser).Error; err != nil {
			s.logger.Error("Failed to update test user level", zap.Error(err))
		}
		
		token, err := s.jwtMiddleware.GenerateToken(existingUser.ID.String(), existingUser.Username, existingUser.Level)
		if err != nil {
			return nil, fmt.Errorf("生成令牌失败: %w", err)
		}
		
		return &AuthResponse{
			Token:    token,
			UserID:   existingUser.ID.String(),
			Username: existingUser.Username,
			Email:    existingUser.Email,
		}, nil
	}

	// 创建测试用户
	hashedPassword := hashPassword("password123")
	
	user := User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
		Level:    5, // 设置为L5等级，满足权限要求
	}
	
	if err := s.db.Create(&user).Error; err != nil {
		s.logger.Error("Failed to create test user", zap.Error(err))
		return nil, fmt.Errorf("创建测试用户失败: %w", err)
	}
	
	// 生成JWT令牌
	token, err := s.jwtMiddleware.GenerateToken(user.ID.String(), user.Username, user.Level)
	if err != nil {
		s.logger.Error("Failed to generate token", zap.Error(err))
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}
	
	return &AuthResponse{
		Token:    token,
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
	}, nil
}

// hashPassword 哈希密码
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// verifyPassword 验证密码
func verifyPassword(password, hashedPassword string) bool {
	return hashPassword(password) == hashedPassword
}