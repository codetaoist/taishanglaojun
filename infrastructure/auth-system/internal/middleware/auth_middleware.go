package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/jwt"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/models"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/service"
)

// AuthMiddleware 认证中间件件
type AuthMiddleware struct {
	jwtManager  *jwt.Manager
	authService service.AuthService
	logger      *zap.Logger
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(jwtManager *jwt.Manager, authService service.AuthService, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager:  jwtManager,
		authService: authService,
		logger:      logger,
	}
}

// RequireAuth 需要认证的中间件
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 首先检查是否有来自API网关的认证头部
		if authValidated := c.GetHeader("X-Auth-Validated"); authValidated == "true" {
			// 从API网关传递的头部中获取用户信息
			userIDStr := c.GetHeader("X-User-ID")
			username := c.GetHeader("X-Username")
			email := c.GetHeader("X-User-Email")
			role := c.GetHeader("X-User-Role")
			sessionIDStr := c.GetHeader("X-Session-ID")
			
			if userIDStr != "" && username != "" {
				// 解析UUID
				userID, err := uuid.Parse(userIDStr)
				if err != nil {
					m.logger.Warn("Invalid user ID from gateway", zap.String("user_id", userIDStr), zap.Error(err))
				} else {
					sessionID, _ := uuid.Parse(sessionIDStr)
					
					// 设置用户信息到上下文
					c.Set("user_id", userID)
					c.Set("username", username)
					c.Set("email", email)
					c.Set("role", models.UserRole(role))
					c.Set("session_id", sessionID)
					
					// 设置完整的用户对象到上下文
					user := &models.User{
						ID:       userID,
						Username: username,
						Email:    email,
						Role:     models.UserRole(role),
					}
					c.Set("user", user)
					
					m.logger.Info("User authenticated via gateway", 
						zap.String("user_id", userID.String()),
						zap.String("username", username),
						zap.String("role", role))
					
					c.Next()
					return
				}
			}
		}
		
		// 如果没有API网关认证头部，则使用传统的token验证
		token := m.extractToken(c)
		if token == "" {
			m.logger.Debug("No token found in request", 
				zap.String("authorization_header", c.GetHeader("Authorization")),
				zap.String("query_token", c.Query("token")),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Missing or invalid authorization header",
			})
			c.Abort()
			return
		}

		m.logger.Debug("Token extracted from request", 
			zap.String("token_prefix", token[:min(len(token), 50)]),
			zap.Int("token_length", len(token)),
		)

		m.logger.Debug("About to call JWT ValidateToken method")
		claims, err := m.jwtManager.ValidateToken(token)
		m.logger.Debug("JWT ValidateToken method returned", 
			zap.Error(err),
			zap.Bool("claims_nil", claims == nil),
		)
		
		if err != nil {
			m.logger.Warn("Invalid token", 
				zap.Error(err),
				zap.String("token_prefix", token[:min(len(token), 50)]),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid token",
			})
			c.Abort()
			return
		}

		// 验证令牌是否为访问令牌
		if claims.TokenType != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid token type",
			})
			c.Abort()
			return
		}

		// 验证令牌状态枚举
		validateReq := &models.ValidateTokenRequest{Token: token}
		validateResp, err := m.authService.ValidateToken(c.Request.Context(), validateReq)
		if err != nil || !validateResp.Valid {
			message := "Token validation failed"
			if validateResp != nil {
				message = validateResp.Message
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": message,
			})
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", models.UserRole(claims.Role))
		c.Set("permissions", claims.Permissions)
		c.Set("session_id", claims.SessionID)

		// 设置完整的用户对象到上下
		user := &models.User{
			ID:       claims.UserID,
			Username: claims.Username,
			Email:    claims.Email,
			Role:     models.UserRole(claims.Role),
		}
		c.Set("user", user)

		c.Next()
	}
}

// RequireRole 需要特定角色的中间件
func (m *AuthMiddleware) RequireRole(roles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			m.logger.Warn("Role not found in context")
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "User role not found",
			})
			c.Abort()
			return
		}

		role, ok := userRole.(models.UserRole)
		if !ok {
			m.logger.Warn("Invalid role type in context", zap.Any("role", userRole))
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Invalid user role",
			})
			c.Abort()
			return
		}

		m.logger.Debug("Role validation",
			zap.String("user_role", string(role)),
			zap.Any("required_roles", roles))

		// 检查用户角色是否在允许的角色列表中间件
		for _, allowedRole := range roles {
			if role == allowedRole {
				m.logger.Debug("Role validation passed", zap.String("role", string(role)))
				c.Next()
				return
			}
		}

		m.logger.Warn("Insufficient permissions",
			zap.String("user_role", string(role)),
			zap.Any("required_roles", roles))
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": "Insufficient permissions",
		})
		c.Abort()
	}
}

// RequirePermission 需要特定权限的中间件
func (m *AuthMiddleware) RequirePermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userPermissions, exists := c.Get("permissions")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "User permissions not found",
			})
			c.Abort()
			return
		}

		perms, ok := userPermissions.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Invalid user permissions",
			})
			c.Abort()
			return
		}

		// 检查用户是否拥有效所需的权限
		for _, requiredPerm := range permissions {
			hasPermission := false
			for _, userPerm := range perms {
				if userPerm == requiredPerm {
					hasPermission = true
					break
				}
			}
			if !hasPermission {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "forbidden",
					"message": "Missing required permission: " + requiredPerm,
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// OptionalAuth 可选认证中间件件（不强制要求认证）
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			c.Next()
			return
		}

		claims, err := m.jwtManager.ValidateToken(token)
		if err != nil {
			m.logger.Debug("Optional auth failed", zap.Error(err))
			c.Next()
			return
		}

		// 验证令牌状态枚举
		validateReq := &models.ValidateTokenRequest{Token: token}
		validateResp, err := m.authService.ValidateToken(c.Request.Context(), validateReq)
		if err != nil || !validateResp.Valid {
			c.Next()
			return
		}

		// 设置用户信息到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("permissions", claims.Permissions)
		c.Set("session_id", claims.SessionID)
		c.Set("authenticated", true)

		c.Next()
	}
}

// AdminOnly 仅管理器员访问的中间件件
func (m *AuthMiddleware) AdminOnly() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		m.RequireAuth()(c)
		if c.IsAborted() {
			return
		}
		m.RequireRole(models.RoleAdmin)(c)
	})
}

// SuperAdminOnly 仅超级管理器员访问的中间件件
func (m *AuthMiddleware) SuperAdminOnly() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		m.RequireAuth()(c)
		if c.IsAborted() {
			return
		}
		m.RequireRole(models.RoleSuperAdmin)(c)
	})
}

// RateLimitByUser 按用户限流的中间件
func (m *AuthMiddleware) RateLimitByUser() gin.HandlerFunc {
	// 这里可以实现基于用户的限流逻辑
	// 例如使用户 Redis 存储用户请求计数量
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		// 实现限流逻辑
		// 这里只是示例，实际实现需要使用 Redis 或其他存储来存储用户请求计数
		m.logger.Debug("Rate limiting for user", zap.Any("user_id", userID))

		c.Next()
	}
}

// CORS 跨域中间件（允许所有源）
// CORS 跨域资源共享中间件
func (m *AuthMiddleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 定义允许的源列表（生产环境应该配置具体的域名）
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:5173", // 添加Vite开发服务器端口
			"http://localhost:8080",
			"https://taishang.example.com",
		}

		// 检查请求源是否在允许列表中
		isAllowed := false
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				isAllowed = true
				break
			}
		}

		// 开发环境允许所有localhost源
		if gin.Mode() == gin.DebugMode && strings.Contains(origin, "localhost") {
			isAllowed = true
		}

		if isAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		// 设置其他CORS头
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-Total-Count, X-Page-Count")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24小时预检缓存

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestLogger 请求日志中间件（记录每个 HTTP 请求）
func (m *AuthMiddleware) RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		userID := "anonymous"
		if param.Keys != nil {
			if uid, exists := param.Keys["user_id"]; exists {
				if id, ok := uid.(uuid.UUID); ok {
					userID = id.String()
				}
			}
		}

		m.logger.Info("HTTP Request",
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("ip", param.ClientIP),
			zap.String("user_agent", param.Request.UserAgent()),
			zap.String("user_id", userID),
		)

		return ""
	})
}

// SecurityHeaders 安全头中间件件
// SecurityHeaders 安全响应头中间件
func (m *AuthMiddleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止MIME类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")

		// 防止页面被嵌入到iframe中
		c.Header("X-Frame-Options", "DENY")

		// 启用XSS保护
		c.Header("X-XSS-Protection", "1; mode=block")

		// 强制HTTPS连接
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// 内容安全策略
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self' data:; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'"
		c.Header("Content-Security-Policy", csp)

		// 引用策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 权限策略
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// 防止缓存敏感信息
		if strings.Contains(c.Request.URL.Path, "/api/") {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}

		// 移除服务器信息泄露
		c.Header("Server", "")

		c.Next()
	}
}

// extractToken 从请求中间件提取令牌
func (m *AuthMiddleware) extractToken(c *gin.Context) string {
	// 从 Authorization 头中间件提取令牌
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// 从查询参数量中间件提取（不推荐，但有效时需要）
	token := c.Query("token")
	if token != "" {
		return token
	}

	// 从 Cookie 中间件提取（不推荐，但有效时需要）
	cookie, err := c.Cookie("access_token")
	if err == nil && cookie != "" {
		return cookie
	}

	return ""
}

// GetCurrentUser 获取当前用户信息的辅助函数
func GetCurrentUser(c *gin.Context) (*CurrentUser, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return nil, false
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		return nil, false
	}

	username, _ := c.Get("username")
	email, _ := c.Get("email")
	role, _ := c.Get("role")
	permissions, _ := c.Get("permissions")
	sessionID, _ := c.Get("session_id")

	user := &CurrentUser{
		ID:       uid,
		Username: username.(string),
		Email:    email.(string),
		Role:     role.(models.UserRole),
	}

	if perms, ok := permissions.([]string); ok {
		user.Permissions = perms
	}

	if sid, ok := sessionID.(uuid.UUID); ok {
		user.SessionID = sid
	}

	return user, true
}

// CurrentUser 当前用户信息
type CurrentUser struct {
	ID          uuid.UUID       `json:"id"`
	Username    string          `json:"username"`
	Email       string          `json:"email"`
	Role        models.UserRole `json:"role"`
	Permissions []string        `json:"permissions"`
	SessionID   uuid.UUID       `json:"session_id"`
}

// HasRole 检查用户是否拥有效指定角色
func (u *CurrentUser) HasRole(role models.UserRole) bool {
	return u.Role == role
}

// HasPermission 检查用户是否拥有效指定权限
func (u *CurrentUser) HasPermission(permission string) bool {
	for _, perm := range u.Permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// IsAdmin 检查用户是否为管理员角色
func (u *CurrentUser) IsAdmin() bool {
	return u.Role == models.RoleAdmin
}

// IsSuperAdmin 检查用户是否为超级管理员角色
func (u *CurrentUser) IsSuperAdmin() bool {
	return u.Role == models.RoleSuperAdmin
}
