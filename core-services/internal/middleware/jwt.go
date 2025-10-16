package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret     string
	Issuer     string
	Expiration time.Duration
}

// JWTClaims JWT 声明
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`  // 角色
	Level    int    `json:"level"` // 等级
	jwt.RegisteredClaims
}

// JWTMiddleware JWT 中间件
type JWTMiddleware struct {
	config JWTConfig
	logger *zap.Logger
}

// NewJWTMiddleware JWT 中间件
func NewJWTMiddleware(config JWTConfig, logger *zap.Logger) *JWTMiddleware {
	return &JWTMiddleware{
		config: config,
		logger: logger,
	}
}

// AuthRequired JWT 认证中间件
func (m *JWTMiddleware) AuthRequired() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.logger.Warn("Missing authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "缺少授权头",
			})
			c.Abort()
			return
		}

		// Bearer 格式校验
		if !strings.HasPrefix(authHeader, "Bearer ") {
			m.logger.Warn("Invalid authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "无效的授权头格式",
			})
			c.Abort()
			return
		}

		// token 解析
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			m.logger.Warn("Empty token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "空的token",
			})
			c.Abort()
			return
		}

		// token 校验
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			//
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.config.Secret), nil
		})

		if err != nil {
			m.logger.Warn("Token validation failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "无效的token",
			})
			c.Abort()
			return
		}

		// token 校验
		if !token.Valid {
			m.logger.Warn("Invalid token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "无效的token",
			})
			c.Abort()
			return
		}

		// 校验token声明
		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			m.logger.Warn("Invalid token claims")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "无效的token声明",
			})
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("user_level", claims.Level) // 用户等级

		m.logger.Debug("User authenticated",
			zap.String("user_id", claims.UserID),
			zap.String("username", claims.Username),
			zap.Int("level", claims.Level))

		c.Next()
	})
}

// GenerateToken JWT 生成token
func (m *JWTMiddleware) GenerateToken(userID, username, role string, level int) (string, error) {
	now := time.Now()
	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
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

// OptionalAuth JWT 可选认证中间件
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
