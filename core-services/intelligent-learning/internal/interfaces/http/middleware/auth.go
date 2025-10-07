package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// AuthConfig 认证配置
type AuthConfig struct {
	SecretKey    string
	SkipPaths    []string
	RequiredRole string
}

// Auth 认证中间件
func Auth() gin.HandlerFunc {
	return AuthWithConfig(&AuthConfig{
		SecretKey: "your-secret-key", // 在生产环境中应该从环境变量获取
		SkipPaths: []string{
			"/health",
			"/ready",
			"/live",
			"/swagger",
		},
	})
}

// AuthWithConfig 带配置的认证中间件
func AuthWithConfig(config *AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否跳过认证
		path := c.Request.URL.Path
		for _, skipPath := range config.SkipPaths {
			if strings.HasPrefix(path, skipPath) {
				c.Next()
				return
			}
		}

		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Missing authorization header",
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid authorization header format",
				"message": "Authorization header must start with 'Bearer '",
			})
			c.Abort()
			return
		}

		// 提取token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 验证token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.SecretKey), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// 检查token是否有效
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"message": "Token is not valid",
			})
			c.Abort()
			return
		}

		// 获取claims
		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token claims",
				"message": "Cannot parse token claims",
			})
			c.Abort()
			return
		}

		// 检查角色权限（如果配置了）
		if config.RequiredRole != "" {
			hasRole := false
			for _, role := range claims.Roles {
				if role == config.RequiredRole {
					hasRole = true
					break
				}
			}
			if !hasRole {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "Insufficient permissions",
					"message": "Required role: " + config.RequiredRole,
				})
				c.Abort()
				return
			}
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}

// RequireRole 要求特定角色的中间件
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "No role information",
				"message": "User role information not found",
			})
			c.Abort()
			return
		}

		userRoles, ok := roles.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Invalid role information",
				"message": "Cannot parse user roles",
			})
			c.Abort()
			return
		}

		hasRole := false
		for _, userRole := range userRoles {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Insufficient permissions",
				"message": "Required role: " + role,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) string {
	if username, exists := c.Get("username"); exists {
		if name, ok := username.(string); ok {
			return name
		}
	}
	return ""
}

// GetUserRoles 从上下文获取用户角色
func GetUserRoles(c *gin.Context) []string {
	if roles, exists := c.Get("roles"); exists {
		if userRoles, ok := roles.([]string); ok {
			return userRoles
		}
	}
	return []string{}
}