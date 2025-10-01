package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrTokenNotFound    = errors.New("token not found")
	ErrInvalidSignature = errors.New("invalid token signature")
)

// Claims JWT声明
type Claims struct {
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	SessionID   uuid.UUID `json:"session_id"`
	TokenType   string    `json:"token_type"` // access, refresh
	Permissions []string  `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

// Config JWT配置
type Config struct {
	SecretKey        string        `yaml:"secret_key" env:"JWT_SECRET_KEY"`
	AccessTokenTTL   time.Duration `yaml:"access_token_ttl" env:"JWT_ACCESS_TOKEN_TTL"`
	RefreshTokenTTL  time.Duration `yaml:"refresh_token_ttl" env:"JWT_REFRESH_TOKEN_TTL"`
	Issuer           string        `yaml:"issuer" env:"JWT_ISSUER"`
	SigningMethod    string        `yaml:"signing_method" env:"JWT_SIGNING_METHOD"`
	RefreshThreshold time.Duration `yaml:"refresh_threshold" env:"JWT_REFRESH_THRESHOLD"`
}

// DefaultConfig 默认JWT配置
func DefaultConfig() *Config {
	return &Config{
		SecretKey:        "your-secret-key-change-in-production",
		AccessTokenTTL:   15 * time.Minute,
		RefreshTokenTTL:  7 * 24 * time.Hour, // 7天
		Issuer:           "taishang-auth-system",
		SigningMethod:    "HS256",
		RefreshThreshold: 5 * time.Minute, // 令牌剩余5分钟时可刷新
	}
}

// Manager JWT管理器
type Manager struct {
	config *Config
	logger *zap.Logger
}

// NewManager 创建JWT管理器
func NewManager(config *Config, logger *zap.Logger) *Manager {
	if config == nil {
		config = DefaultConfig()
	}
	
	return &Manager{
		config: config,
		logger: logger,
	}
}

// GenerateAccessToken 生成访问令牌
func (m *Manager) GenerateAccessToken(userID uuid.UUID, username, email, role string, sessionID uuid.UUID, permissions []string) (string, *Claims, error) {
	now := time.Now()
	expiresAt := now.Add(m.config.AccessTokenTTL)
	
	claims := &Claims{
		UserID:      userID,
		Username:    username,
		Email:       email,
		Role:        role,
		SessionID:   sessionID,
		TokenType:   "access",
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    m.config.Issuer,
			Subject:   userID.String(),
			Audience:  []string{"taishang-system"},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.config.SecretKey))
	if err != nil {
		m.logger.Error("Failed to generate access token", zap.Error(err))
		return "", nil, err
	}
	
	m.logger.Debug("Generated access token", 
		zap.String("user_id", userID.String()),
		zap.String("username", username),
		zap.Time("expires_at", expiresAt),
	)
	
	return tokenString, claims, nil
}

// GenerateRefreshToken 生成刷新令牌
func (m *Manager) GenerateRefreshToken(userID uuid.UUID, sessionID uuid.UUID) (string, *Claims, error) {
	now := time.Now()
	expiresAt := now.Add(m.config.RefreshTokenTTL)
	
	claims := &Claims{
		UserID:    userID,
		SessionID: sessionID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    m.config.Issuer,
			Subject:   userID.String(),
			Audience:  []string{"taishang-system"},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.config.SecretKey))
	if err != nil {
		m.logger.Error("Failed to generate refresh token", zap.Error(err))
		return "", nil, err
	}
	
	m.logger.Debug("Generated refresh token", 
		zap.String("user_id", userID.String()),
		zap.Time("expires_at", expiresAt),
	)
	
	return tokenString, claims, nil
}

// ValidateToken 验证令牌
func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSignature
		}
		return []byte(m.config.SecretKey), nil
	})
	
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		m.logger.Error("Failed to parse token", zap.Error(err))
		return nil, ErrInvalidToken
	}
	
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	
	return claims, nil
}

// RefreshToken 刷新令牌
func (m *Manager) RefreshToken(refreshTokenString string) (string, string, *Claims, error) {
	// 验证刷新令牌
	refreshClaims, err := m.ValidateToken(refreshTokenString)
	if err != nil {
		return "", "", nil, err
	}
	
	if refreshClaims.TokenType != "refresh" {
		return "", "", nil, ErrInvalidToken
	}
	
	// 生成新的访问令牌和刷新令牌
	// 注意：这里需要从数据库获取最新的用户信息
	// 为了简化，这里使用现有的claims信息
	accessToken, accessClaims, err := m.GenerateAccessToken(
		refreshClaims.UserID,
		refreshClaims.Username,
		refreshClaims.Email,
		refreshClaims.Role,
		refreshClaims.SessionID,
		refreshClaims.Permissions,
	)
	if err != nil {
		return "", "", nil, err
	}
	
	newRefreshToken, _, err := m.GenerateRefreshToken(refreshClaims.UserID, refreshClaims.SessionID)
	if err != nil {
		return "", "", nil, err
	}
	
	return accessToken, newRefreshToken, accessClaims, nil
}

// ExtractTokenFromHeader 从Authorization头提取令牌
func (m *Manager) ExtractTokenFromHeader(authHeader string) string {
	const bearerPrefix = "Bearer "
	if len(authHeader) > len(bearerPrefix) && authHeader[:len(bearerPrefix)] == bearerPrefix {
		return authHeader[len(bearerPrefix):]
	}
	return ""
}

// IsTokenExpiringSoon 检查令牌是否即将过期
func (m *Manager) IsTokenExpiringSoon(claims *Claims) bool {
	if claims.ExpiresAt == nil {
		return false
	}
	
	timeUntilExpiry := time.Until(claims.ExpiresAt.Time)
	return timeUntilExpiry <= m.config.RefreshThreshold
}

// GetTokenTTL 获取令牌TTL
func (m *Manager) GetTokenTTL(tokenType string) time.Duration {
	switch tokenType {
	case "access":
		return m.config.AccessTokenTTL
	case "refresh":
		return m.config.RefreshTokenTTL
	default:
		return m.config.AccessTokenTTL
	}
}

// GenerateTokenPair 生成令牌对（访问令牌+刷新令牌）
func (m *Manager) GenerateTokenPair(userID uuid.UUID, username, email, role string, sessionID uuid.UUID, permissions []string) (string, string, *Claims, error) {
	accessToken, accessClaims, err := m.GenerateAccessToken(userID, username, email, role, sessionID, permissions)
	if err != nil {
		return "", "", nil, err
	}
	
	refreshToken, _, err := m.GenerateRefreshToken(userID, sessionID)
	if err != nil {
		return "", "", nil, err
	}
	
	return accessToken, refreshToken, accessClaims, nil
}

// RevokeToken 撤销令牌（需要配合黑名单实现）
func (m *Manager) RevokeToken(tokenString string) error {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return err
	}
	
	// 这里应该将令牌添加到黑名单
	// 可以使用Redis存储黑名单，键为JTI，值为过期时间
	m.logger.Info("Token revoked", 
		zap.String("jti", claims.ID),
		zap.String("user_id", claims.UserID.String()),
	)
	
	return nil
}

// GetConfig 获取配置
func (m *Manager) GetConfig() *Config {
	return m.config
}