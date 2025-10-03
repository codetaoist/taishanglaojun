package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *AuthService
	logger      *zap.Logger
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid register request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "请求参数无效: " + err.Error(),
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
		"message": "注册成功",
	})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid login request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "请求参数无效: " + err.Error(),
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
		"message": "登录成功",
	})
}

// GetCurrentUser 获取当前用户信息
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "UNAUTHORIZED",
			"message": "用户未认证",
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

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user_id":  userID,
			"username": user.Username,
			"email":    user.Email,
		},
		"message": "获取用户信息成功",
	})
}

// CreateTestUser 创建测试用户（仅用于开发和测试）
func (h *AuthHandler) CreateTestUser(c *gin.Context) {
	resp, err := h.authService.CreateTestUser()
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
		"message": "测试用户创建成功",
	})
}

// SetupAuthRoutes 设置认证路由
func SetupAuthRoutes(router *gin.RouterGroup, authHandler *AuthHandler, jwtMiddleware *JWTMiddleware) {
	// 公开路由（不需要认证）
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/test-user", authHandler.CreateTestUser) // 仅用于开发测试
	}

	// 需要认证的路由
	protected := router.Group("/auth")
	protected.Use(jwtMiddleware.AuthRequired())
	{
		protected.GET("/me", authHandler.GetCurrentUser)
	}
}