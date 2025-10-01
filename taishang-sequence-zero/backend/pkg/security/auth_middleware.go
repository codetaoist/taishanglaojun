package security

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	jwtService *JWTService
}

// NewAuthMiddleware 创建新的认证中间件
func NewAuthMiddleware(jwtService *JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

// RequireAuth 需要认证的中间件
func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取令牌
		authHeader := c.GetHeader("Authorization")
		token, err := ExtractTokenFromHeader(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "认证失败",
				"message": err.Error(),
				"code":    "AUTHENTICATION_REQUIRED",
			})
			c.Abort()
			return
		}

		// 验证令牌
		claims, err := am.jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "认证失败",
				"message": err.Error(),
				"code":    "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("level", claims.Level)
		c.Set("permissions", claims.Permissions)
		c.Set("claims", claims)

		c.Next()
	}
}

// RequirePermission 需要特定权限的中间件
func (am *AuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行认证
		am.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		// 检查权限
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "内部错误",
				"message": "无法获取用户信息",
				"code":    "INTERNAL_ERROR",
			})
			c.Abort()
			return
		}

		userClaims := claims.(*JWTClaims)
		if !userClaims.HasPermission(permission) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "权限不足",
				"message": "您没有执行此操作的权限",
				"code":    "INSUFFICIENT_PERMISSION",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission 需要任一权限的中间件
func (am *AuthMiddleware) RequireAnyPermission(permissions []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行认证
		am.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		// 检查权限
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "内部错误",
				"message": "无法获取用户信息",
				"code":    "INTERNAL_ERROR",
			})
			c.Abort()
			return
		}

		userClaims := claims.(*JWTClaims)
		if !userClaims.HasAnyPermission(permissions) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "权限不足",
				"message": "您没有执行此操作的权限",
				"code":    "INSUFFICIENT_PERMISSION",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireLevel 需要特定等级的中间件
func (am *AuthMiddleware) RequireLevel(requiredLevel int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行认证
		am.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		// 检查等级
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "内部错误",
				"message": "无法获取用户信息",
				"code":    "INTERNAL_ERROR",
			})
			c.Abort()
			return
		}

		userClaims := claims.(*JWTClaims)
		if !userClaims.IsLevelSufficient(requiredLevel) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "等级不足",
				"message": "您的等级不足以执行此操作",
				"code":    "INSUFFICIENT_LEVEL",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth 可选认证中间件（不强制要求认证）
func (am *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 没有提供认证信息，继续处理
			c.Next()
			return
		}

		token, err := ExtractTokenFromHeader(authHeader)
		if err != nil {
			// 认证信息格式错误，继续处理但不设置用户信息
			c.Next()
			return
		}

		// 验证令牌
		claims, err := am.jwtService.ValidateToken(token)
		if err != nil {
			// 令牌无效，继续处理但不设置用户信息
			c.Next()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("level", claims.Level)
		c.Set("permissions", claims.Permissions)
		c.Set("claims", claims)

		c.Next()
	}
}

// GetUserFromContext 从上下文中获取用户信息
func GetUserFromContext(c *gin.Context) (*JWTClaims, bool) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, false
	}

	userClaims, ok := claims.(*JWTClaims)
	return userClaims, ok
}

// GetUserIDFromContext 从上下文中获取用户ID
func GetUserIDFromContext(c *gin.Context) (int, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	id, ok := userID.(int)
	return id, ok
}

// GetUserLevelFromContext 从上下文中获取用户等级
func GetUserLevelFromContext(c *gin.Context) (int, bool) {
	level, exists := c.Get("level")
	if !exists {
		return 0, false
	}

	userLevel, ok := level.(int)
	return userLevel, ok
}

// GetUserPermissionsFromContext 从上下文中获取用户权限
func GetUserPermissionsFromContext(c *gin.Context) ([]string, bool) {
	permissions, exists := c.Get("permissions")
	if !exists {
		return nil, false
	}

	userPermissions, ok := permissions.([]string)
	return userPermissions, ok
}

// AdminOnly 仅管理员可访问的中间件
func (am *AuthMiddleware) AdminOnly() gin.HandlerFunc {
	return am.RequireLevel(9) // L9为最高管理员等级
}

// ModeratorOrAbove 版主及以上等级可访问的中间件
func (am *AuthMiddleware) ModeratorOrAbove() gin.HandlerFunc {
	return am.RequireLevel(6) // L6及以上为版主等级
}

// VerifiedUserOnly 已验证用户可访问的中间件
func (am *AuthMiddleware) VerifiedUserOnly() gin.HandlerFunc {
	return am.RequireLevel(2) // L2及以上为已验证用户
}

// RoleBasedAccess 基于角色的访问控制中间件
func (am *AuthMiddleware) RoleBasedAccess(allowedRoles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行认证
		am.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		// 检查角色权限
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "内部错误",
				"message": "无法获取用户信息",
				"code":    "INTERNAL_ERROR",
			})
			c.Abort()
			return
		}

		userClaims := claims.(*JWTClaims)
		
		// 将用户等级转换为角色
		userRole := levelToRole(userClaims.Level)
		
		// 检查是否有允许的角色
		allowed := false
		for _, role := range allowedRoles {
			if userRole == role {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "角色权限不足",
				"message": "您的角色无权访问此资源",
				"code":    "INSUFFICIENT_ROLE",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// levelToRole 将等级转换为角色名称
func levelToRole(level int) string {
	switch {
	case level >= 9:
		return "admin"
	case level >= 6:
		return "moderator"
	case level >= 3:
		return "verified_user"
	case level >= 1:
		return "user"
	default:
		return "guest"
	}
}

// IPWhitelistMiddleware IP白名单中间件
func IPWhitelistMiddleware(allowedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := getClientIP(c)
		
		allowed := false
		for _, ip := range allowedIPs {
			if clientIP == ip {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "IP访问被拒绝",
				"message": "您的IP地址不在允许列表中",
				"code":    "IP_NOT_ALLOWED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SessionValidationMiddleware 会话验证中间件
func (am *AuthMiddleware) SessionValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行认证
		am.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		// 获取会话ID（可以从cookie或header中获取）
		sessionID := c.GetHeader("X-Session-ID")
		if sessionID == "" {
			if cookie, err := c.Cookie("session_id"); err == nil {
				sessionID = cookie
			}
		}

		if sessionID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "会话无效",
				"message": "缺少会话标识",
				"code":    "INVALID_SESSION",
			})
			c.Abort()
			return
		}

		// 这里可以添加会话验证逻辑
		// 例如检查Redis中的会话信息
		
		c.Set("session_id", sessionID)
		c.Next()
	}
}