package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"golang.org/x/crypto/bcrypt"

	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
	"github.com/codetaoist/taishanglaojun/core-services/internal/models"
)

// AdminHandler 管理员处理器
type AdminHandler struct {
	authService *middleware.AuthService
	logger      *zap.Logger
	db          *gorm.DB
}

// NewAdminHandler 创建管理员处理器
func NewAdminHandler(authService *middleware.AuthService, logger *zap.Logger, db *gorm.DB) *AdminHandler {
	return &AdminHandler{
		authService: authService,
		logger:      logger,
		db:          db,
	}
}

// hashPassword 哈希密码
func (h *AdminHandler) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// UserResponse 用户响应结构
type UserResponse struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	Level       int       `json:"level"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastLogin   *time.Time `json:"last_login"`
}

// UserStats 用户统计
type UserStats struct {
    TotalUsers    int64 `json:"total_users"`
    ActiveUsers   int64 `json:"active_users"`
    AdminUsers    int64 `json:"admin_users"`
    NewUsers      int64 `json:"new_users"`
    NewUsersToday int64 `json:"new_users_today"`
    OnlineUsers   int64 `json:"online_users"`
}

// GetUsers 获取用户列表
func (h *AdminHandler) GetUsers(c *gin.Context) {
	// 检查管理员权限
	if !h.checkAdminPermission(c) {
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	role := c.Query("role")
	status := c.Query("status")

	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// 从数据库获取用户数据
	var users []models.User
	var totalCount int64
	
	// 构建查询
	query := h.db.Model(&models.User{})
	
	// 应用搜索过滤
	if search != "" {
		query = query.Where("username LIKE ? OR email LIKE ? OR display_name LIKE ?", 
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	
	// 应用角色过滤
	if role != "" {
		query = query.Where("role = ?", role)
	}
	
	// 应用状态过滤
	if status != "" {
		if status == "active" {
			query = query.Where("is_active = ?", true)
		} else if status == "inactive" {
			query = query.Where("is_active = ?", false)
		}
	}
	
	// 获取总数
	if err := query.Count(&totalCount).Error; err != nil {
		h.logger.Error("获取用户总数失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取用户数据失败",
		})
		return
	}
	
	// 分页查询
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error; err != nil {
		h.logger.Error("获取用户列表失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取用户数据失败",
		})
		return
	}
	
	// 转换为响应格式
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		status := "active"
		if !user.IsActive {
			status = "inactive"
		}
		
		userResponses[i] = UserResponse{
			ID:          user.ID.String(),
			Username:    user.Username,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			Role:        string(user.Role),
			Status:      status,
			Level:       user.Level,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			LastLogin:   user.LastLoginAt,
		}
	}

	// 计算总页数
	totalPages := int((totalCount + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"users": userResponses,
			"pagination": gin.H{
				"page":        page,
				"limit":       limit,
				"total":       totalCount,
				"total_pages": totalPages,
			},
		},
	})
}

// GetUserStats 获取用户统计
func (h *AdminHandler) GetUserStats(c *gin.Context) {
	// 检查管理员权限
	if !h.checkAdminPermission(c) {
		return
	}

	var stats UserStats
	
	// 获取总用户数
	if err := h.db.Model(&models.User{}).Count(&stats.TotalUsers).Error; err != nil {
		h.logger.Error("获取总用户数失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取统计数据失败",
		})
		return
	}
	
	// 获取活跃用户数（is_active = true）
	if err := h.db.Model(&models.User{}).Where("is_active = ?", true).Count(&stats.ActiveUsers).Error; err != nil {
		h.logger.Error("获取活跃用户数失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取统计数据失败",
		})
		return
	}
	
    // 获取管理员数量（含 super_admin）
    if err := h.db.Model(&models.User{}).Where("role IN (?)", []string{"admin", "super_admin"}).Count(&stats.AdminUsers).Error; err != nil {
        h.logger.Error("获取管理员数量失败", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "获取统计数据失败",
        })
        return
    }

    // 获取新用户数（最近7天注册的用户）
    sevenDaysAgo := time.Now().AddDate(0, 0, -7)
    if err := h.db.Model(&models.User{}).Where("created_at >= ?", sevenDaysAgo).Count(&stats.NewUsers).Error; err != nil {
        h.logger.Error("获取新用户数失败", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "获取统计数据失败",
        })
        return
    }

    // 获取今日新增用户数（当天0点至当前）
    now := time.Now()
    todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
    if err := h.db.Model(&models.User{}).Where("created_at >= ?", todayStart).Count(&stats.NewUsersToday).Error; err != nil {
        h.logger.Error("获取今日新增用户数失败", zap.Error(err))
        // 今日新增失败时不影响整体统计，设为0
        stats.NewUsersToday = 0
    }
	
	// 获取在线用户数（最近30分钟有登录活动的用户）
	thirtyMinutesAgo := time.Now().Add(-30 * time.Minute)
	if err := h.db.Model(&models.User{}).Where("last_login_at >= ?", thirtyMinutesAgo).Count(&stats.OnlineUsers).Error; err != nil {
		h.logger.Error("获取在线用户数失败", zap.Error(err))
		// 在线用户数获取失败时，设为0而不是返回错误
		stats.OnlineUsers = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// CreateUser 创建用户
func (h *AdminHandler) CreateUser(c *gin.Context) {
	// 检查管理员权限
	if !h.checkAdminPermission(c) {
		return
	}

	var req struct {
		Username    string `json:"username" binding:"required"`
		Email       string `json:"email" binding:"required,email"`
		Password    string `json:"password" binding:"required,min=6"`
		DisplayName string `json:"display_name"`
		Role        string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	// 设置默认值
	if req.Role == "" {
		req.Role = "USER"
	}
	if req.DisplayName == "" {
		req.DisplayName = req.Username
	}

	// 检查用户名和邮箱是否已存在
	var existingUser models.User
	if err := h.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "用户名或邮箱已存在",
		})
		return
	}

	// 哈希密码
	hashedPassword, err := h.hashPassword(req.Password)
	if err != nil {
		h.logger.Error("密码哈希失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "密码处理失败",
		})
		return
	}

	// 创建用户
	user := models.User{
		Username:    req.Username,
		Email:       req.Email,
		Password:    hashedPassword,
		DisplayName: req.DisplayName,
		Role:        models.UserRole(req.Role),
		Level:       1,
		IsActive:    true,
	}

	if err := h.db.Create(&user).Error; err != nil {
		h.logger.Error("创建用户失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "创建用户失败",
		})
		return
	}

	// 转换为响应格式
	userResponse := UserResponse{
		ID:          user.ID.String(),
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        string(user.Role),
		Status:      "active",
		Level:       user.Level,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		LastLogin:   user.LastLoginAt,
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    userResponse,
		"message": "用户创建成功",
	})
}

// UpdateUser 更新用户
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	// 检查管理员权限
	if !h.checkAdminPermission(c) {
		return
	}

	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "用户ID不能为空",
		})
		return
	}

	var req struct {
		Username    string `json:"username"`
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
		Role        string `json:"role"`
		Status      string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	// 查找用户
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "用户不存在",
			})
		} else {
			h.logger.Error("查找用户失败", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "查找用户失败",
			})
		}
		return
	}

	// 检查用户名和邮箱是否被其他用户使用
	if req.Username != "" && req.Username != user.Username {
		var existingUser models.User
		if err := h.db.Where("username = ? AND id != ?", req.Username, userID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "用户名已被使用",
			})
			return
		}
	}

	if req.Email != "" && req.Email != user.Email {
		var existingUser models.User
		if err := h.db.Where("email = ? AND id != ?", req.Email, userID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "邮箱已被使用",
			})
			return
		}
	}

	// 更新用户信息
	updates := make(map[string]interface{})
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.DisplayName != "" {
		updates["display_name"] = req.DisplayName
	}
	if req.Role != "" {
		updates["role"] = models.UserRole(req.Role)
	}
	if req.Status != "" {
		updates["is_active"] = req.Status == "active"
	}

	if err := h.db.Model(&user).Updates(updates).Error; err != nil {
		h.logger.Error("更新用户失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "更新用户失败",
		})
		return
	}

	// 重新查询更新后的用户信息
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		h.logger.Error("查询更新后用户失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询更新后用户失败",
		})
		return
	}

	// 转换为响应格式
	userResponse := UserResponse{
		ID:          user.ID.String(),
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        string(user.Role),
		Status:      "active",
		Level:       user.Level,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		LastLogin:   user.LastLoginAt,
	}

	if !user.IsActive {
		userResponse.Status = "inactive"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    userResponse,
		"message": "用户更新成功",
	})
}

// DeleteUser 删除用户
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	// 检查管理员权限
	if !h.checkAdminPermission(c) {
		return
	}

	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "用户ID不能为空",
		})
		return
	}

	// 检查用户是否存在
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "用户不存在",
			})
		} else {
			h.logger.Error("查找用户失败", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "查找用户失败",
			})
		}
		return
	}

	// 删除用户
	if err := h.db.Delete(&user).Error; err != nil {
		h.logger.Error("删除用户失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "删除用户失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户删除成功",
	})
}

// BatchDeleteUsers 批量删除用户
func (h *AdminHandler) BatchDeleteUsers(c *gin.Context) {
	// 检查管理员权限
	if !h.checkAdminPermission(c) {
		return
	}

	var req struct {
		UserIDs []string `json:"userIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	// 批量删除用户
	var success, failed int
	for _, userID := range req.UserIDs {
		var user models.User
		if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				h.logger.Warn("用户不存在", zap.String("userID", userID))
			} else {
				h.logger.Error("查找用户失败", zap.String("userID", userID), zap.Error(err))
			}
			failed++
			continue
		}

		if err := h.db.Delete(&user).Error; err != nil {
			h.logger.Error("删除用户失败", zap.String("userID", userID), zap.Error(err))
			failed++
			continue
		}

		success++
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"success": success,
			"failed":  failed,
		},
		"message": "批量删除完成",
	})
}

// UpdateUserStatus 更新用户状态
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	// 检查管理员权限
	if !h.checkAdminPermission(c) {
		return
	}

	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "用户ID不能为空",
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	// 验证状态值
	validStatuses := []string{"active", "inactive", "suspended", "banned"}
	isValid := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValid = true
			break
		}
	}

	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的状态值",
		})
		return
	}

	// 查找用户
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "用户不存在",
			})
		} else {
			h.logger.Error("查找用户失败", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "查找用户失败",
			})
		}
		return
	}

	// 更新用户状态
	isActive := req.Status == "active"
	if err := h.db.Model(&user).Update("is_active", isActive).Error; err != nil {
		h.logger.Error("更新用户状态失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "更新用户状态失败",
		})
		return
	}

	// 重新查询更新后的用户信息
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		h.logger.Error("查询更新后用户失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询更新后用户失败",
		})
		return
	}

	// 转换为响应格式
	userResponse := UserResponse{
		ID:          user.ID.String(),
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        string(user.Role),
		Status:      req.Status,
		Level:       user.Level,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		LastLogin:   user.LastLoginAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    userResponse,
		"message": "用户状态更新成功",
	})
}

// checkAdminPermission 检查管理员权限
func (h *AdminHandler) checkAdminPermission(c *gin.Context) bool {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "未授权访问",
		})
		return false
	}

	// 获取用户信息
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		h.logger.Error("Failed to get user", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "用户验证失败",
		})
		return false
	}

	// 检查是否为管理员
	role := string(user.Role)
	if role != "admin" && role != "ADMIN" && role != "super_admin" && role != "SUPER_ADMIN" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "权限不足，需要管理员权限",
		})
		return false
	}

	return true
}

// contains 检查字符串是否包含子字符串（忽略大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(substr) > 0 && 
		 (s[:len(substr)] == substr || 
		  s[len(s)-len(substr):] == substr ||
		  (len(s) > len(substr) && 
		   func() bool {
			   for i := 1; i <= len(s)-len(substr); i++ {
				   if s[i:i+len(substr)] == substr {
					   return true
				   }
			   }
			   return false
		   }()))))
}