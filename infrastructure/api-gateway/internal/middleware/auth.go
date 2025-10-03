package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/logger"
)

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecret     string        `yaml:"jwt_secret"`
	TokenExpiry   time.Duration `yaml:"token_expiry"`
	RefreshExpiry time.Duration `yaml:"refresh_expiry"`
	
	// Redis配置（用于token黑名单）
	RedisAddr     string `yaml:"redis_addr"`
	RedisPassword string `yaml:"redis_password"`
	RedisDB       int    `yaml:"redis_db"`
	
	// 跳过认证的路径
	SkipPaths []string `yaml:"skip_paths"`
	
	// 可选的认证路径（不强制要求token）
	OptionalPaths []string `yaml:"optional_paths"`
}

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	config      *AuthConfig
	redisClient *redis.Client
	logger      logger.Logger
}

// UserClaims JWT用户声明
type UserClaims struct {
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	SessionID   uuid.UUID `json:"session_id"`
	TokenType   string    `json:"token_type"`
	Permissions []string  `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(config *AuthConfig, log logger.Logger) (*AuthMiddleware, error) {
	var redisClient *redis.Client
	
	// 如果配置了Redis，则创建Redis客户端
	if config.RedisAddr != "" {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     config.RedisAddr,
			Password: config.RedisPassword,
			DB:       config.RedisDB,
		})
		
		// 测试连接
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if err := redisClient.Ping(ctx).Err(); err != nil {
			return nil, fmt.Errorf("failed to connect to Redis: %w", err)
		}
	}
	
	return &AuthMiddleware{
		config:      config,
		redisClient: redisClient,
		logger:      log,
	}, nil
}

// Handler 中间件处理器
func (a *AuthMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method
		
		// 调试日志：确认中间件被调用
		a.logger.Infof("AUTH MIDDLEWARE CALLED: %s %s", method, path)
		
		// 检查是否跳过认证
		if a.shouldSkipAuth(path) {
			c.Next()
			return
		}
		
		// 检查是否为可选认证路径
		optional := a.isOptionalAuth(path)
		
		// 提取token
		token, err := a.extractToken(c)
		if err != nil {
			if optional {
				c.Next()
				return
			}
			a.respondWithError(c, http.StatusUnauthorized, "Missing or invalid token")
			return
		}
		
		// 验证token
		claims, err := a.validateToken(token)
		if err != nil {
			if optional {
				c.Next()
				return
			}
			a.respondWithError(c, http.StatusUnauthorized, "Invalid token: "+err.Error())
			return
		}
		
		// 检查token是否在黑名单中
		if a.redisClient != nil {
			if blacklisted, err := a.isTokenBlacklisted(c.Request.Context(), token); err != nil {
				a.logger.Errorf("Failed to check token blacklist: %v", err)
			} else if blacklisted {
				if optional {
					c.Next()
					return
				}
				a.respondWithError(c, http.StatusUnauthorized, "Token has been revoked")
				return
			}
		}
		
		// 将用户信息添加到上下文
		a.setUserContext(c, claims)
		
		c.Next()
	}
}

// shouldSkipAuth 检查是否应该跳过认证
func (a *AuthMiddleware) shouldSkipAuth(path string) bool {
	for _, skipPath := range a.config.SkipPaths {
		if a.matchPath(path, skipPath) {
			return true
		}
	}
	return false
}

// isOptionalAuth 检查是否为可选认证路径
func (a *AuthMiddleware) isOptionalAuth(path string) bool {
	for _, optionalPath := range a.config.OptionalPaths {
		if a.matchPath(path, optionalPath) {
			return true
		}
	}
	return false
}

// matchPath 路径匹配（支持通配符）
func (a *AuthMiddleware) matchPath(path, pattern string) bool {
	// 简单的通配符匹配
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}
	return path == pattern
}

// extractToken 从请求中提取token
func (a *AuthMiddleware) extractToken(c *gin.Context) (string, error) {
	// 从Authorization header提取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1], nil
		}
	}
	
	// 从查询参数提取
	if token := c.Query("token"); token != "" {
		return token, nil
	}
	
	// 从Cookie提取
	if cookie, err := c.Cookie("access_token"); err == nil && cookie != "" {
		return cookie, nil
	}
	
	return "", fmt.Errorf("no token found")
}

// validateToken 验证JWT token
func (a *AuthMiddleware) validateToken(tokenString string) (*UserClaims, error) {
	// 添加调试日志
	a.logger.Infof("Validating token with secret length: %d", len(a.config.JWTSecret))
	a.logger.Infof("JWT Secret (first 10 chars): %s...", a.config.JWTSecret[:10])
	
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.config.JWTSecret), nil
	})
	
	if err != nil {
		a.logger.Errorf("JWT validation failed: %v", err)
		return nil, err
	}
	
	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		// 检查token是否过期
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			return nil, fmt.Errorf("token has expired")
		}
		a.logger.Infof("JWT validation successful for user: %s", claims.Username)
		return claims, nil
	}
	
	return nil, fmt.Errorf("invalid token claims")
}

// isTokenBlacklisted 检查token是否在黑名单中
func (a *AuthMiddleware) isTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	if a.redisClient == nil {
		return false, nil
	}
	
	key := fmt.Sprintf("blacklist:token:%s", token)
	result := a.redisClient.Exists(ctx, key)
	if result.Err() != nil {
		return false, result.Err()
	}
	
	return result.Val() > 0, nil
}

// setUserContext 设置用户上下文信息
func (a *AuthMiddleware) setUserContext(c *gin.Context, claims *UserClaims) {
	c.Set("user_id", claims.UserID)
	c.Set("username", claims.Username)
	c.Set("email", claims.Email)
	c.Set("role", claims.Role)
	c.Set("session_id", claims.SessionID)
	c.Set("permissions", claims.Permissions)
	c.Set("user_claims", claims)
	
	// 将用户信息添加到请求头中，传递给后端服务
	c.Request.Header.Set("X-User-ID", claims.UserID.String())
	c.Request.Header.Set("X-Username", claims.Username)
	c.Request.Header.Set("X-User-Email", claims.Email)
	c.Request.Header.Set("X-User-Role", claims.Role)
	c.Request.Header.Set("X-Session-ID", claims.SessionID.String())
	c.Request.Header.Set("X-Auth-Validated", "true")
	
	// 添加调试日志
	a.logger.Infof("Setting user headers for backend: user_id=%s, username=%s, role=%s", 
		claims.UserID.String(), claims.Username, claims.Role)
}

// respondWithError 返回错误响应
func (a *AuthMiddleware) respondWithError(c *gin.Context, statusCode int, message string) {
	a.logger.WithFields(map[string]interface{}{
		"path":        c.Request.URL.Path,
		"method":      c.Request.Method,
		"remote_addr": c.ClientIP(),
		"user_agent":  c.GetHeader("User-Agent"),
	}).Warn("Authentication failed: " + message)
	
	c.JSON(statusCode, gin.H{
		"error":   "Authentication failed",
		"message": message,
		"code":    statusCode,
	})
	c.Abort()
}

// BlacklistToken 将token加入黑名单
func (a *AuthMiddleware) BlacklistToken(ctx context.Context, token string, expiry time.Duration) error {
	if a.redisClient == nil {
		return fmt.Errorf("Redis client not configured")
	}
	
	key := fmt.Sprintf("blacklist:token:%s", token)
	return a.redisClient.Set(ctx, key, "1", expiry).Err()
}

// GetUserFromContext 从上下文获取用户信息
func GetUserFromContext(c *gin.Context) (*UserClaims, bool) {
	if claims, exists := c.Get("user_claims"); exists {
		if userClaims, ok := claims.(*UserClaims); ok {
			return userClaims, true
		}
	}
	return nil, false
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uuid.UUID); ok {
			return id, true
		}
	}
	return uuid.Nil, false
}

// GetUserRoles 从上下文获取用户角色
func GetUserRoles(c *gin.Context) (string, bool) {
	if role, exists := c.Get("role"); exists {
		if userRole, ok := role.(string); ok {
			return userRole, true
		}
	}
	return "", false
}

// HasRole 检查用户是否具有指定角色
func HasRole(c *gin.Context, role string) bool {
	userRole, exists := GetUserRoles(c)
	if !exists {
		return false
	}
	
	return userRole == role
}

// RequireRole 要求用户具有指定角色的中间件
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !HasRole(c, role) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Insufficient permissions",
				"message": fmt.Sprintf("Role '%s' required", role),
				"code":    http.StatusForbidden,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAnyRole 要求用户具有任意指定角色的中间件
func RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := GetUserRoles(c)
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Insufficient permissions",
				"message": "No role found",
				"code":    http.StatusForbidden,
			})
			c.Abort()
			return
		}
		
		for _, requiredRole := range roles {
			if userRole == requiredRole {
				c.Next()
				return
			}
		}
		
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Insufficient permissions",
			"message": fmt.Sprintf("One of roles %v required", roles),
			"code":    http.StatusForbidden,
		})
		c.Abort()
	}
}

// Close 关闭中间件资源
func (a *AuthMiddleware) Close() error {
	if a.redisClient != nil {
		return a.redisClient.Close()
	}
	return nil
}