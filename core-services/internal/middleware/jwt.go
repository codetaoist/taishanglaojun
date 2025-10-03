package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string
	Issuer     string
	Expiration time.Duration
}

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Level    int    `json:"level"` // 用户等级
	jwt.RegisteredClaims
}

// JWTMiddleware JWT认证中间件
type JWTMiddleware struct {
	config JWTConfig
	logger *zap.Logger
}

// NewJWTMiddleware 创建JWT中间件
func NewJWTMiddleware(config JWTConfig, logger *zap.Logger) *JWTMiddleware {
	return &JWTMiddleware{
		config: config,
		logger: logger,
	}
}

// AuthRequired JWT认证中间件
func (m *JWTMiddleware) AuthRequired() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.logger.Warn("Missing authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "用户未认证",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			m.logger.Warn("Invalid authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "认证格式无效",
			})
			c.Abort()
			return
		}

		// 提取token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			m.logger.Warn("Empty token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "认证令牌为空",
			})
			c.Abort()
			return
		}

		// 解析和验证token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// 验证签名方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.config.Secret), nil
		})

		if err != nil {
			m.logger.Warn("Token validation failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "认证令牌无效",
			})
			c.Abort()
			return
		}

		// 检查token是否有效
		if !token.Valid {
			m.logger.Warn("Invalid token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "认证令牌无效",
			})
			c.Abort()
			return
		}

		// 提取用户信息
		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			m.logger.Warn("Invalid token claims")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "认证令牌格式错误",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_level", claims.Level) // 添加用户等级到上下文

		m.logger.Debug("User authenticated", 
			zap.String("user_id", claims.UserID),
			zap.String("username", claims.Username),
			zap.Int("level", claims.Level))

		c.Next()
	})
}

// GenerateToken 生成JWT令牌
func (m *JWTMiddleware) GenerateToken(userID, username string, level int) (string, error) {
	now := time.Now()
	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		Level:    level,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.config.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.Expiration)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.config.Secret))
}

// OptionalAuth 可选认证中间件（不强制要求认证）
func (m *JWTMiddleware) OptionalAuth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			c.Next()
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.config.Secret), nil
		})

		if err != nil || !token.Valid {
			c.Next()
			return
		}

		if claims, ok := token.Claims.(*JWTClaims); ok {
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
		}

		c.Next()
	})
}