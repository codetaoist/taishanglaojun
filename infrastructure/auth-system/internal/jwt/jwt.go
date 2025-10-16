package jwt

import (
	"errors"
	"fmt"
	"strings"
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
	// 新增安全配置
	MinSecretKeyLength int           `yaml:"min_secret_key_length" env:"JWT_MIN_SECRET_KEY_LENGTH"`
	MaxTokenAge        time.Duration `yaml:"max_token_age" env:"JWT_MAX_TOKEN_AGE"`
	RequireAudience    bool          `yaml:"require_audience" env:"JWT_REQUIRE_AUDIENCE"`
	AllowedAudiences   []string      `yaml:"allowed_audiences" env:"JWT_ALLOWED_AUDIENCES"`
}

// DefaultConfig 默认JWT配置
func DefaultConfig() *Config {
	return &Config{
		SecretKey:          "your-secret-key-change-in-production",
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    7 * 24 * time.Hour, // 7天
		Issuer:             "taishang-auth-system",
		SigningMethod:      "HS256",
		RefreshThreshold:   5 * time.Minute, // 令牌剩余5分钟时可刷新
		MinSecretKeyLength: 32,              // 最小密钥长度32字符
		MaxTokenAge:        24 * time.Hour,  // 最大令牌年龄24小时
		RequireAudience:    true,            // 要求验证audience
		AllowedAudiences:   []string{"taishang-system", "taishang-web", "taishang-mobile"},
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

	// 添加配置调试日志
	logger.Info("JWT Manager initialized with config",
		zap.String("issuer", config.Issuer),
		zap.Bool("require_audience", config.RequireAudience),
		zap.Strings("allowed_audiences", config.AllowedAudiences),
		zap.Duration("access_token_ttl", config.AccessTokenTTL),
		zap.Duration("refresh_token_ttl", config.RefreshTokenTTL),
		zap.Duration("max_token_age", config.MaxTokenAge),
	)

	return &Manager{
		config: config,
		logger: logger,
	}
}

// GenerateAccessToken 生成访问令牌
func (m *Manager) GenerateAccessToken(userID uuid.UUID, username, email, role string, sessionID uuid.UUID, permissions []string) (string, *Claims, error) {
	now := time.Now().UTC()
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
			Audience:  m.config.AllowedAudiences,
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
	now := time.Now().UTC()
	expiresAt := now.Add(m.config.RefreshTokenTTL)

	claims := &Claims{
		UserID:    userID,
		SessionID: sessionID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    m.config.Issuer,
			Subject:   userID.String(),
			Audience:  m.config.AllowedAudiences,
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
	m.logger.Debug("Starting token validation", 
		zap.String("token_prefix", tokenString[:min(len(tokenString), 50)]),
		zap.Int("token_length", len(tokenString)),
		zap.Int64("current_time", time.Now().Unix()),
	)

	// 检查token是否为空
	if tokenString == "" {
		m.logger.Debug("Token is empty")
		return nil, fmt.Errorf("token is empty")
	}

	// 检查token格式
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		m.logger.Debug("Invalid token format", zap.Int("parts_count", len(parts)))
		return nil, fmt.Errorf("invalid token format")
	}

	m.logger.Debug("Token format validation passed", zap.Int("parts_count", len(parts)))

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		m.logger.Debug("JWT parsing callback called", 
			zap.String("token_method", token.Method.Alg()),
			zap.String("config_method", m.config.SigningMethod),
		)
		
		// 验证签名方法
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			m.logger.Error("Invalid signing method type", 
				zap.String("method", token.Method.Alg()),
				zap.String("method_type", fmt.Sprintf("%T", token.Method)),
			)
			return nil, ErrInvalidSignature
		}
		if method.Alg() != m.config.SigningMethod {
			m.logger.Error("Invalid signing method algorithm", 
				zap.String("expected", m.config.SigningMethod),
				zap.String("actual", method.Alg()),
			)
			return nil, ErrInvalidSignature
		}
		m.logger.Debug("Signing method validated", 
			zap.String("method", method.Alg()),
			zap.Int("secret_key_length", len(m.config.SecretKey)),
		)
		return []byte(m.config.SecretKey), nil
	})

	if err != nil {
		m.logger.Error("Token parsing failed", 
			zap.Error(err),
			zap.String("error_type", fmt.Sprintf("%T", err)),
		)
		if errors.Is(err, jwt.ErrTokenExpired) {
			m.logger.Warn("Token has expired", zap.Error(err))
			return nil, ErrExpiredToken
		}
		m.logger.Error("Failed to parse token", zap.Error(err))
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		m.logger.Error("Invalid token claims or token not valid", 
			zap.Bool("claims_ok", ok),
			zap.Bool("token_valid", token.Valid),
		)
		return nil, ErrInvalidToken
	}

	// 添加时间戳debug信息
	currentTime := time.Now().Unix()
	m.logger.Debug("Token time validation", 
		zap.Int64("current_time", currentTime),
		zap.Int64("token_exp", claims.ExpiresAt.Unix()),
		zap.Int64("token_nbf", claims.NotBefore.Unix()),
		zap.Int64("token_iat", claims.IssuedAt.Unix()),
		zap.Bool("is_expired", currentTime > claims.ExpiresAt.Unix()),
		zap.Bool("is_before_nbf", currentTime < claims.NotBefore.Unix()),
	)

	m.logger.Debug("Token parsed successfully", 
		zap.String("user_id", claims.UserID.String()),
		zap.String("token_type", claims.TokenType),
		zap.String("issuer", claims.Issuer),
		zap.Strings("audience", claims.Audience),
	)

	// 验证发行者
	if claims.Issuer != m.config.Issuer {
		m.logger.Warn("Invalid token issuer",
			zap.String("expected", m.config.Issuer),
			zap.String("actual", claims.Issuer),
		)
		return nil, ErrInvalidToken
	}

	// 验证audience（如果启用）
	if m.config.RequireAudience && len(claims.Audience) > 0 {
		m.logger.Debug("Validating audience", 
			zap.Bool("require_audience", m.config.RequireAudience),
			zap.Strings("token_audience", claims.Audience),
			zap.Strings("allowed_audiences", m.config.AllowedAudiences),
		)
		
		audienceValid := false
		for _, aud := range claims.Audience {
			if m.config.IsAudienceAllowed(aud) {
				audienceValid = true
				break
			}
		}
		if !audienceValid {
			m.logger.Warn("Invalid token audience", 
				zap.Strings("audience", claims.Audience),
				zap.Strings("allowed", m.config.AllowedAudiences),
			)
			return nil, ErrInvalidToken
		}
	}

	// 验证令牌年龄
	if claims.IssuedAt != nil {
		tokenAge := time.Now().UTC().Sub(claims.IssuedAt.Time)
		m.logger.Debug("Token age validation", 
			zap.Duration("age", tokenAge),
			zap.Duration("max_age", m.config.MaxTokenAge),
			zap.Time("issued_at", claims.IssuedAt.Time),
			zap.Time("current_time", time.Now().UTC()),
		)
		if tokenAge > m.config.MaxTokenAge {
			m.logger.Warn("Token is too old",
				zap.Duration("age", tokenAge),
				zap.Duration("max_age", m.config.MaxTokenAge),
			)
			return nil, ErrExpiredToken
		}
	}

	// 验证NotBefore
	if claims.NotBefore != nil && time.Now().UTC().Before(claims.NotBefore.Time) {
		m.logger.Warn("Token not yet valid", zap.Time("not_before", claims.NotBefore.Time))
		return nil, ErrInvalidToken
	}

	m.logger.Debug("Token validation successful", 
		zap.String("user_id", claims.UserID.String()),
		zap.String("token_type", claims.TokenType),
	)

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
	// 注册意：这里需要从数量据库获取最新的用户信息
	// 为了简化，这里使用户现有效的claims信息
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

	// 这里应该将令牌添加到黑名单（Redis）
	// 键为JTI，值为过期时间
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

// ValidateConfig 验证JWT配置
func (c *Config) ValidateConfig() error {
	// 验证密钥长度
	if len(c.SecretKey) < c.MinSecretKeyLength {
		return errors.New("JWT secret key is too short")
	}

	// 验证密钥不能是默认值
	if c.SecretKey == "your-secret-key-change-in-production" {
		return errors.New("JWT secret key must be changed from default value")
	}

	// 验证令牌有效期
	if c.AccessTokenTTL <= 0 {
		return errors.New("access token TTL must be positive")
	}

	if c.RefreshTokenTTL <= 0 {
		return errors.New("refresh token TTL must be positive")
	}

	// 验证签名方法
	supportedMethods := []string{"HS256", "HS384", "HS512", "RS256", "RS384", "RS512"}
	methodSupported := false
	for _, method := range supportedMethods {
		if c.SigningMethod == method {
			methodSupported = true
			break
		}
	}
	if !methodSupported {
		return errors.New("unsupported JWT signing method")
	}

	return nil
}

// IsAudienceAllowed 检查audience是否被允许
func (c *Config) IsAudienceAllowed(audience string) bool {
	if !c.RequireAudience {
		return true
	}

	for _, allowed := range c.AllowedAudiences {
		if allowed == audience {
			return true
		}
	}
	return false
}
