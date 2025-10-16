package middleware

import (
    "fmt"
    "time"
    "strings"

    "golang.org/x/crypto/bcrypt"
    "go.uber.org/zap"
    "gorm.io/gorm"

    "github.com/codetaoist/taishanglaojun/core-services/internal/models"
)



// AuthService 
type AuthService struct {
    db            *gorm.DB
    jwtMiddleware *JWTMiddleware
    logger        *zap.Logger
}

// LoginRequest 
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// AuthResponse 
type AuthResponse struct {
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// NewAuthService 
func NewAuthService(db *gorm.DB, jwtMiddleware *JWTMiddleware, logger *zap.Logger) *AuthService {
	return &AuthService{
		db:            db,
		jwtMiddleware: jwtMiddleware,
		logger:        logger,
	}
}

// AutoMigrate 
func (s *AuthService) AutoMigrate() error {
	// 
	if err := s.db.AutoMigrate(&models.User{}); err != nil {
		return fmt.Errorf(": %v", err)
	}
	return nil
}

// Register 
func (s *AuthService) Register(req *RegisterRequest) (*AuthResponse, error) {
	// 
	var existingUser models.User
	if err := s.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("")
	}

	// 
	// salt, err := generateSalt()
	// if err != nil {
	//	return nil, fmt.Errorf(": %w", err)
	// }

	// 
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 
	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     models.RoleUser,
	}

	if err := s.db.Create(user).Error; err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, fmt.Errorf(": %w", err)
	}

	// 
	now := time.Now()
	user.LastLoginAt = &now
	s.db.Save(&user)

	// JWT
	token, err := s.jwtMiddleware.GenerateToken(user.ID.String(), user.Username, string(user.Role), user.Level)
	if err != nil {
		s.logger.Error("Failed to generate token", zap.Error(err))
		return nil, fmt.Errorf(": %w", err)
	}

	return &AuthResponse{
		Token:    token,
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
	}, nil
}

// Login 
func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
    // 支持用户名或邮箱登录
    var user models.User
    if err := s.db.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, fmt.Errorf("用户不存在")
        }
        s.logger.Error("Failed to find user", zap.Error(err))
        return nil, fmt.Errorf("查找用户失败: %w", err)
    }

    // 验证密码；支持明文到bcrypt的迁移（一次性）
    if !verifyPassword(req.Password, user.Password) {
        // 如果存储的密码不是bcrypt哈希且与输入明文一致，则自动迁移为bcrypt
        if !isBcryptHash(user.Password) && user.Password == req.Password {
            hashed, err := hashPassword(req.Password)
            if err != nil {
                s.logger.Error("Failed to hash plaintext password during migration", zap.Error(err))
                return nil, fmt.Errorf("密码处理失败: %w", err)
            }
            user.Password = hashed
            if err := s.db.Save(&user).Error; err != nil {
                s.logger.Error("Failed to persist migrated bcrypt password", zap.Error(err))
                return nil, fmt.Errorf("密码迁移失败: %w", err)
            }
        } else {
            return nil, fmt.Errorf("密码验证失败")
        }
    }

    // 更新最近登录时间
    now := time.Now()
    user.LastLoginAt = &now
    _ = s.db.Save(&user)

	// JWT
	token, err := s.jwtMiddleware.GenerateToken(user.ID.String(), user.Username, string(user.Role), user.Level)
	if err != nil {
		s.logger.Error("Failed to generate token", zap.Error(err))
		return nil, fmt.Errorf(": %w", err)
	}

	return &AuthResponse{
		Token:    token,
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
	}, nil
}

// GetUserByID ID
func (s *AuthService) GetUserByID(userID string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("")
		}
		s.logger.Error("Failed to get user by ID", zap.Error(err))
		return nil, fmt.Errorf(": %v", err)
	}
	return &user, nil
}

// CreateTestUser 允许指定用户名、角色与密码，便于前端测试管理员页面
func (s *AuthService) CreateTestUser(username string, role models.UserRole, password string) (*AuthResponse, error) {
    // 删除同名测试用户（如果存在）
    s.db.Where("username = ?", username).Delete(&models.User{})
    s.logger.Info("Deleted existing test user if any", zap.String("username", username))

    // 使用传入的测试密码
    hashedPassword, err := hashPassword(password)
    if err != nil {
        s.logger.Error("Failed to hash password", zap.Error(err))
        return nil, fmt.Errorf("密码加密失败: %w", err)
    }

    // 默认用简单邮箱占位
    email := fmt.Sprintf("%s@example.com", username)

    user := models.User{
        Username: username,
        Email:    email,
        Password: hashedPassword,
        Role:     role,
        Level:    5, // 测试用户默认L5
    }

    if err := s.db.Create(&user).Error; err != nil {
        s.logger.Error("Failed to create test user", zap.Error(err))
        return nil, fmt.Errorf(": %w", err)
    }

    // 签发JWT
    token, err := s.jwtMiddleware.GenerateToken(user.ID.String(), user.Username, string(user.Role), user.Level)
    if err != nil {
        s.logger.Error("Failed to generate token", zap.Error(err))
        return nil, fmt.Errorf(": %w", err)
    }

    return &AuthResponse{
        Token:    token,
        UserID:   user.ID.String(),
        Username: user.Username,
        Email:    user.Email,
    }, nil
}

// hashPassword 
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// verifyPassword 
func verifyPassword(password, hashedPassword string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    return err == nil
}

// isBcryptHash 判断是否为bcrypt哈希格式
func isBcryptHash(p string) bool {
    return strings.HasPrefix(p, "$2a$") || strings.HasPrefix(p, "$2b$") || strings.HasPrefix(p, "$2y$")
}

