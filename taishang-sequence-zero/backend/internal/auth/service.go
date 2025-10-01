package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Service 认证服务结构体
type Service struct {
	db        *sql.DB
	jwtSecret []byte
}

// User 用户结构体
type User struct {
	ID              int       `json:"id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	PasswordHash    string    `json:"-"`
	PermissionLevel int       `json:"permission_level"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// LoginRequest 登录请求结构体
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求结构体
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// TokenResponse 令牌响应结构体
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	User         User   `json:"user"`
}

// Claims JWT声明结构体
type Claims struct {
	UserID          int    `json:"user_id"`
	Username        string `json:"username"`
	PermissionLevel int    `json:"permission_level"`
	jwt.RegisteredClaims
}

// NewService 创建新的认证服务实例
func NewService(db *sql.DB) *Service {
	// 在生产环境中，应该从环境变量或配置文件中读取密钥
	jwtSecret := []byte("taishang-laojun-sequence-zero-secret-key")
	return &Service{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

// Register 用户注册
func (s *Service) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数", "details": err.Error()})
		return
	}

	// 检查用户名是否已存在
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 OR email = $2)", req.Username, req.Email).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询失败"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "用户名或邮箱已存在"})
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 创建用户
	var userID int
	err = s.db.QueryRow(`
		INSERT INTO users (username, email, password_hash, permission_level, is_active) 
		VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		req.Username, req.Email, string(hashedPassword), 1, true).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户创建失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "用户注册成功",
		"user_id": userID,
	})
}

// Login 用户登录
func (s *Service) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 查询用户
	var user User
	err := s.db.QueryRow(`
		SELECT id, username, email, password_hash, permission_level, is_active, created_at, updated_at 
		FROM users WHERE username = $1 AND is_active = true`,
		req.Username).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.PermissionLevel, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询失败"})
		}
		return
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 生成JWT令牌
	accessToken, refreshToken, expiresIn, err := s.generateTokens(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "令牌生成失败"})
		return
	}

	// 记录登录会话
	sessionID, err := s.createSession(user.ID, refreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "会话创建失败"})
		return
	}

	// 清除密码哈希
	user.PasswordHash = ""

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		User:         user,
	})

	// 记录审计日志
	s.logAuditEvent(user.ID, "LOGIN", fmt.Sprintf("用户 %s 登录成功，会话ID: %s", user.Username, sessionID))
}

// Logout 用户登出
func (s *Service) Logout(c *gin.Context) {
	refreshToken := c.GetHeader("Refresh-Token")
	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少刷新令牌"})
		return
	}

	// 删除会话
	_, err := s.db.Exec("DELETE FROM user_sessions WHERE refresh_token = $1", refreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "登出失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
}

// RefreshToken 刷新访问令牌
func (s *Service) RefreshToken(c *gin.Context) {
	refreshToken := c.GetHeader("Refresh-Token")
	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少刷新令牌"})
		return
	}

	// 验证刷新令牌
	var userID int
	var expiresAt time.Time
	err := s.db.QueryRow(`
		SELECT user_id, expires_at FROM user_sessions 
		WHERE refresh_token = $1 AND expires_at > NOW()`,
		refreshToken).Scan(&userID, &expiresAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效或过期的刷新令牌"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "令牌验证失败"})
		}
		return
	}

	// 获取用户信息
	var user User
	err = s.db.QueryRow(`
		SELECT id, username, email, permission_level, is_active, created_at, updated_at 
		FROM users WHERE id = $1 AND is_active = true`,
		userID).Scan(&user.ID, &user.Username, &user.Email,
		&user.PermissionLevel, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在或已禁用"})
		return
	}

	// 生成新的访问令牌
	accessToken, _, expiresIn, err := s.generateTokens(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "令牌生成失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"expires_in":   expiresIn,
	})
}

// VerifyToken 验证访问令牌
func (s *Service) VerifyToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少授权头"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := s.validateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":            true,
		"user_id":          claims.UserID,
		"username":         claims.Username,
		"permission_level": claims.PermissionLevel,
		"expires_at":       claims.ExpiresAt.Time,
	})
}

// GetProfile 获取用户资料
func (s *Service) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var user User
	err := s.db.QueryRow(`
		SELECT id, username, email, permission_level, is_active, created_at, updated_at 
		FROM users WHERE id = $1`,
		userID).Scan(&user.ID, &user.Username, &user.Email,
		&user.PermissionLevel, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile 更新用户资料
func (s *Service) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req struct {
		Email string `json:"email" binding:"omitempty,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	if req.Email != "" {
		_, err := s.db.Exec("UPDATE users SET email = $1, updated_at = NOW() WHERE id = $2", req.Email, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "资料更新成功"})
}

// AuthMiddleware 认证中间件
func (s *Service) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少授权头"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := s.validateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌"})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("permission_level", claims.PermissionLevel)
		c.Next()
	}
}

// generateTokens 生成访问令牌和刷新令牌
func (s *Service) generateTokens(user User) (string, string, int64, error) {
	// 访问令牌有效期15分钟
	accessExpiresAt := time.Now().Add(15 * time.Minute)
	accessClaims := &Claims{
		UserID:          user.ID,
		Username:        user.Username,
		PermissionLevel: user.PermissionLevel,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   strconv.Itoa(user.ID),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return "", "", 0, err
	}

	// 生成刷新令牌（随机字符串）
	refreshTokenBytes := make([]byte, 32)
	_, err = rand.Read(refreshTokenBytes)
	if err != nil {
		return "", "", 0, err
	}
	refreshToken := hex.EncodeToString(refreshTokenBytes)

	return accessTokenString, refreshToken, int64(accessExpiresAt.Sub(time.Now()).Seconds()), nil
}

// validateToken 验证JWT令牌
func (s *Service) validateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// createSession 创建用户会话
func (s *Service) createSession(userID int, refreshToken string) (string, error) {
	sessionIDBytes := make([]byte, 16)
	_, err := rand.Read(sessionIDBytes)
	if err != nil {
		return "", err
	}
	sessionID := hex.EncodeToString(sessionIDBytes)

	// 刷新令牌有效期7天
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	_, err = s.db.Exec(`
		INSERT INTO user_sessions (session_id, user_id, refresh_token, expires_at) 
		VALUES ($1, $2, $3, $4)`,
		sessionID, userID, refreshToken, expiresAt)

	return sessionID, err
}

// logAuditEvent 记录审计事件
func (s *Service) logAuditEvent(userID int, action, details string) {
	_, err := s.db.Exec(`
		INSERT INTO audit_logs (user_id, action, details, ip_address, user_agent) 
		VALUES ($1, $2, $3, $4, $5)`,
		userID, action, details, "", "")
	if err != nil {
		// 记录日志但不影响主要流程
		fmt.Printf("审计日志记录失败: %v\n", err)
	}
}