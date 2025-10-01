package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService JWT服务
type JWTService struct {
	secretKey []byte
	issuer    string
}

// JWTClaims JWT声明
type JWTClaims struct {
	UserID      int    `json:"user_id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Level       int    `json:"level"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// NewJWTService 创建新的JWT服务
func NewJWTService(secretKey, issuer string) *JWTService {
	return &JWTService{
		secretKey: []byte(secretKey),
		issuer:    issuer,
	}
}

// GenerateToken 生成JWT令牌
func (j *JWTService) GenerateToken(userID int, username, email string, level int, permissions []string, expiration time.Duration) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID:      userID,
		Username:    username,
		Email:       email,
		Level:       level,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", fmt.Errorf("生成JWT令牌失败: %w", err)
	}

	return tokenString, nil
}

// ValidateToken 验证JWT令牌
func (j *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("解析JWT令牌失败: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		// 验证发行者
		if claims.Issuer != j.issuer {
			return nil, errors.New("无效的令牌发行者")
		}

		// 验证过期时间
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			return nil, errors.New("令牌已过期")
		}

		// 验证生效时间
		if claims.NotBefore != nil && claims.NotBefore.Time.After(time.Now()) {
			return nil, errors.New("令牌尚未生效")
		}

		return claims, nil
	}

	return nil, errors.New("无效的JWT令牌")
}

// RefreshToken 刷新JWT令牌
func (j *JWTService) RefreshToken(tokenString string, expiration time.Duration) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("验证原令牌失败: %w", err)
	}

	// 检查令牌是否在刷新窗口内（例如，过期前30分钟内）
	refreshWindow := 30 * time.Minute
	if claims.ExpiresAt != nil && time.Until(claims.ExpiresAt.Time) > refreshWindow {
		return "", errors.New("令牌刷新时间过早")
	}

	// 生成新令牌
	return j.GenerateToken(
		claims.UserID,
		claims.Username,
		claims.Email,
		claims.Level,
		claims.Permissions,
		expiration,
	)
}

// ExtractTokenFromHeader 从HTTP头中提取令牌
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("缺少Authorization头")
	}

	// 检查Bearer前缀
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.New("无效的Authorization头格式")
	}

	return authHeader[len(bearerPrefix):], nil
}

// HasPermission 检查用户是否有指定权限
func (c *JWTClaims) HasPermission(permission string) bool {
	for _, p := range c.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyPermission 检查用户是否有任一指定权限
func (c *JWTClaims) HasAnyPermission(permissions []string) bool {
	for _, permission := range permissions {
		if c.HasPermission(permission) {
			return true
		}
	}
	return false
}

// HasAllPermissions 检查用户是否有所有指定权限
func (c *JWTClaims) HasAllPermissions(permissions []string) bool {
	for _, permission := range permissions {
		if !c.HasPermission(permission) {
			return false
		}
	}
	return true
}

// IsLevelSufficient 检查用户等级是否足够
func (c *JWTClaims) IsLevelSufficient(requiredLevel int) bool {
	return c.Level >= requiredLevel
}

// GetTokenInfo 获取令牌信息
func (c *JWTClaims) GetTokenInfo() map[string]interface{} {
	return map[string]interface{}{
		"user_id":     c.UserID,
		"username":    c.Username,
		"email":       c.Email,
		"level":       c.Level,
		"permissions": c.Permissions,
		"issuer":      c.Issuer,
		"issued_at":   c.IssuedAt,
		"expires_at":  c.ExpiresAt,
		"not_before":  c.NotBefore,
	}
}