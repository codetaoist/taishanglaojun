package middleware

import (
    "fmt"
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "gorm.io/gorm"

    "github.com/codetaoist/taishanglaojun/core-services/internal/models"
)

// AuthHandler 
type AuthHandler struct {
    authService *AuthService
    logger      *zap.Logger
    db          *gorm.DB
}

// NewAuthHandler 
func NewAuthHandler(authService *AuthService, logger *zap.Logger, db *gorm.DB) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
		db:          db,
	}
}

// Register 
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid register request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": ": " + err.Error(),
		})
		return
	}

	resp, err := h.authService.Register(&req)
	if err != nil {
		h.logger.Error("Registration failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "REGISTRATION_FAILED",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
		"message": "",
	})
}

// Login 
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid login request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": ": " + err.Error(),
		})
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		h.logger.Error("Login failed", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "LOGIN_FAILED",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
		"message": "",
	})
}

// GetCurrentUser 
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "UNAUTHORIZED",
			"message": "",
		})
		return
	}

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		h.logger.Error("Failed to get user", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "USER_NOT_FOUND",
			"message": err.Error(),
		})
		return
	}

	// 从数据库获取用户的实际权限
	permissions, roles, err := h.getUserPermissionsFromDB(userID)
	if err != nil {
		h.logger.Error("Failed to get user permissions", zap.Error(err))
		// 如果获取权限失败，使用基础权限
		permissions = []string{"user:read"}
		roles = []string{"user"}
	} else {
		h.logger.Info("Successfully got user permissions", 
			zap.String("userID", userID),
			zap.Strings("permissions", permissions),
			zap.Strings("roles", roles))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":           userID,
			"user_id":      userID,
			"username":     user.Username,
			"email":        user.Email,
			"display_name": user.Username,
			"role":         string(user.Role),
			"roles":        roles,
			"permissions":  permissions,
			"level":        user.Level,
			"isAdmin":      contains(roles, "admin") || contains(roles, "super_admin") || contains(roles, "系统管理员"),
			"created_at":   user.CreatedAt,
			"updated_at":   user.UpdatedAt,
		},
		"message": "",
	})
}

// getUserPermissionsFromDB 从数据库获取用户权限
func (h *AuthHandler) getUserPermissionsFromDB(userID string) ([]string, []string, error) {
	// 获取用户角色 - 使用简化的结构体映射
	var userRoles []struct {
		RoleID   string `gorm:"column:role_id"`
		RoleIDDB string `gorm:"column:id"`
		RoleName string `gorm:"column:name"`
		RoleCode string `gorm:"column:code"`
	}

	err := h.db.Table("user_roles").
		Select("user_roles.role_id, roles.id, roles.name, roles.code").
		Joins("LEFT JOIN roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Scan(&userRoles).Error

	if err != nil {
		return nil, nil, fmt.Errorf("获取用户角色失败: %v", err)
	}

	var roleNames []string
	var roleIDs []string
	
	for _, ur := range userRoles {
		roleNames = append(roleNames, ur.RoleName)
		roleIDs = append(roleIDs, ur.RoleID)
	}

	// 如果用户没有角色，返回基础权限
	if len(roleIDs) == 0 {
		return []string{"user:read"}, []string{"user"}, nil
	}

	// 获取角色权限
	var permissions []struct {
		Code string
	}

	err = h.db.Table("role_permissions").
		Select("permissions.code").
		Joins("LEFT JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id IN ?", roleIDs).
		Scan(&permissions).Error

	if err != nil {
		return nil, nil, fmt.Errorf("获取角色权限失败: %v", err)
	}

	var permissionCodes []string
	for _, p := range permissions {
		permissionCodes = append(permissionCodes, p.Code)
	}

	// 去重
	permissionCodes = removeDuplicates(permissionCodes)
	roleNames = removeDuplicates(roleNames)

	return permissionCodes, roleNames, nil
}

// contains 检查字符串数组是否包含指定字符串
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

// removeDuplicates 去除字符串数组中的重复项
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string
	
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

// Logout
func (h *AuthHandler) Logout(c *gin.Context) {
    // 获取用户ID用于审计日志（如果存在）
    userID := c.GetString("user_id")
    if userID != "" {
        h.logger.Info("User logout", zap.String("user_id", userID))
    } else {
        h.logger.Info("User logout without user_id in context")
    }

    // 当前实现基于前端清除令牌；后续可加入令牌撤销/黑名单
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": gin.H{
            "message": "logout successful",
        },
        "message": "",
    })
}

// CreateTestUser 
func (h *AuthHandler) CreateTestUser(c *gin.Context) {
    // 支持通过查询参数指定用户名与角色
    username := c.DefaultQuery("username", "testuser")
    roleStr := c.DefaultQuery("role", "user")
    password := c.DefaultQuery("password", "password123")

    var role models.UserRole
    switch strings.ToLower(roleStr) {
    case "admin":
        role = models.RoleAdmin
    case "super_admin", "superadmin":
        role = models.RoleSuperAdmin
    default:
        role = models.RoleUser
    }

    resp, err := h.authService.CreateTestUser(username, role, password)
    if err != nil {
        h.logger.Error("Failed to create test user", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "CREATE_TEST_USER_FAILED",
            "message": err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    resp,
        "message": "",
    })
}

// SetupAuthRoutes 
func SetupAuthRoutes(router *gin.RouterGroup, authHandler *AuthHandler, jwtMiddleware *JWTMiddleware) {
    // 
    auth := router.Group("/auth")
    {
        auth.POST("/register", authHandler.Register)
        auth.POST("/login", authHandler.Login)
        auth.POST("/test-user", authHandler.CreateTestUser) // 
    }

    // 
    protected := router.Group("/auth")
    protected.Use(jwtMiddleware.AuthRequired())
    {
        protected.GET("/me", authHandler.GetCurrentUser)
        protected.POST("/logout", authHandler.Logout)
    }
}

