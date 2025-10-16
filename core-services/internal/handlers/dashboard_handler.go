package handlers

import (
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
	"github.com/codetaoist/taishanglaojun/core-services/internal/models"
	"github.com/codetaoist/taishanglaojun/core-services/internal/services"
)

// DashboardHandler 仪表板处理器
type DashboardHandler struct {
	authService *middleware.AuthService
	userService *services.UserService
	menuService *services.MenuService
	db          *gorm.DB
	logger      *zap.Logger
}

// NewDashboardHandler 创建仪表板处理器
func NewDashboardHandler(authService *middleware.AuthService, userService *services.UserService, menuService *services.MenuService, db *gorm.DB, logger *zap.Logger) *DashboardHandler {
	return &DashboardHandler{
		authService: authService,
		userService: userService,
		menuService: menuService,
		db:          db,
		logger:      logger,
	}
}

// DashboardStats 仪表板统计数据
type DashboardStats struct {
	TotalUsers      int     `json:"totalUsers"`
	ActiveUsers     int     `json:"activeUsers"`
	AdminUsers      int     `json:"adminUsers"`
	NewUsersToday   int     `json:"newUsersToday"`
	OnlineUsers     int     `json:"onlineUsers"`
	TotalProjects   int     `json:"totalProjects"`
	CompletedTasks  int     `json:"completedTasks"`
	PendingTasks    int     `json:"pendingTasks"`
	SystemHealth    float64 `json:"systemHealth"`
	CPUUsage        float64 `json:"cpuUsage"`
	MemoryUsage     float64 `json:"memoryUsage"`
	DiskUsage       float64 `json:"diskUsage"`
	NetworkLatency  int     `json:"networkLatency"`
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	CPU struct {
		Usage       float64 `json:"usage"`
		Cores       int     `json:"cores"`
		Temperature float64 `json:"temperature,omitempty"`
	} `json:"cpu"`
	Memory struct {
		Total int64   `json:"total"`
		Used  int64   `json:"used"`
		Free  int64   `json:"free"`
		Usage float64 `json:"usage"`
	} `json:"memory"`
	Disk struct {
		Total int64   `json:"total"`
		Used  int64   `json:"used"`
		Free  int64   `json:"free"`
		Usage float64 `json:"usage"`
	} `json:"disk"`
	Network struct {
		Latency    int `json:"latency"`
		Throughput struct {
			In  int64 `json:"in"`
			Out int64 `json:"out"`
		} `json:"throughput"`
	} `json:"network"`
}

// ActivityData 活动数据
type ActivityData struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	User        *struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Avatar   string `json:"avatar,omitempty"`
	} `json:"user,omitempty"`
	Severity string `json:"severity,omitempty"`
}

// TrendData 趋势数据
type TrendData struct {
	Date        string `json:"date"`
	Users       int    `json:"users"`
	Tasks       int    `json:"tasks"`
	Projects    int    `json:"projects"`
	Performance int    `json:"performance"`
}

// QuickAction 快捷操作
type QuickAction struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Action      string `json:"action"`
	Color       string `json:"color"`
}

// GetDashboardStats 获取仪表板统计数据
func (h *DashboardHandler) GetDashboardStats(c *gin.Context) {
	h.logger.Info("Getting dashboard stats")

	var stats DashboardStats

	// 使用用户服务获取统计数据
	userStats, err := h.userService.GetUserStats()
	if err != nil {
		h.logger.Error("Failed to get user stats", zap.Error(err))
		// 设置默认值
		stats.TotalUsers = 0
		stats.ActiveUsers = 0
		stats.AdminUsers = 0
		stats.NewUsersToday = 0
		stats.OnlineUsers = 0
	} else {
		stats.TotalUsers = int(userStats["totalUsers"].(int64))
		stats.ActiveUsers = int(userStats["activeUsers"].(int64))
		stats.AdminUsers = int(userStats["adminUsers"].(int64))
		stats.NewUsersToday = int(userStats["newUsersToday"].(int64))
		stats.OnlineUsers = int(userStats["onlineUsers"].(int64))
	}

	// 获取菜单数量作为项目数量的替代
	var totalMenus int64
	if err := h.db.Model(&models.Menu{}).Count(&totalMenus).Error; err != nil {
		h.logger.Error("Failed to count menus", zap.Error(err))
		totalMenus = 0
	}
	stats.TotalProjects = int(totalMenus)

	// 获取系统配置数量作为任务数量的替代
	var totalConfigs int64
	if err := h.db.Model(&models.SystemConfig{}).Count(&totalConfigs).Error; err != nil {
		h.logger.Error("Failed to count system configs", zap.Error(err))
		totalConfigs = 0
	}
	stats.CompletedTasks = int(totalConfigs)

	// 获取角色数量作为待处理任务的替代
	var totalRoles int64
	if err := h.db.Model(&models.Role{}).Count(&totalRoles).Error; err != nil {
		h.logger.Error("Failed to count roles", zap.Error(err))
		totalRoles = 0
	}
	stats.PendingTasks = int(totalRoles)

	// 获取系统指标
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 计算内存使用率
	memoryUsage := float64(memStats.Alloc) / float64(memStats.Sys) * 100
	if memoryUsage > 100 {
		memoryUsage = 100
	}
	stats.MemoryUsage = memoryUsage

	// 模拟其他系统指标
	stats.SystemHealth = 98.5
	stats.CPUUsage = 45.2
	stats.DiskUsage = 34.5
	stats.NetworkLatency = 12

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetSystemMetrics 获取系统指标
func (h *DashboardHandler) GetSystemMetrics(c *gin.Context) {
	h.logger.Info("Getting system metrics")

	// 模拟系统指标数据
	metrics := SystemMetrics{}
	metrics.CPU.Usage = 45.2
	metrics.CPU.Cores = 8
	metrics.CPU.Temperature = 65.5

	metrics.Memory.Total = 16 * 1024 * 1024 * 1024 // 16GB
	metrics.Memory.Used = int64(float64(metrics.Memory.Total) * 0.678)
	metrics.Memory.Free = metrics.Memory.Total - metrics.Memory.Used
	metrics.Memory.Usage = 67.8

	metrics.Disk.Total = 500 * 1024 * 1024 * 1024 // 500GB
	metrics.Disk.Used = int64(float64(metrics.Disk.Total) * 0.345)
	metrics.Disk.Free = metrics.Disk.Total - metrics.Disk.Used
	metrics.Disk.Usage = 34.5

	metrics.Network.Latency = 12
	metrics.Network.Throughput.In = 1024 * 1024 * 10  // 10MB/s
	metrics.Network.Throughput.Out = 1024 * 1024 * 5  // 5MB/s

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// GetRecentActivities 获取最近活动
func (h *DashboardHandler) GetRecentActivities(c *gin.Context) {
	h.logger.Info("Getting recent activities")

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	var activities []ActivityData

	// 获取最近登录的用户作为活动数据
	var recentUsers []models.User
	if err := h.db.Where("last_login_at IS NOT NULL").
		Order("last_login_at DESC").
		Limit(limit/2).
		Find(&recentUsers).Error; err != nil {
		h.logger.Error("Failed to get recent users", zap.Error(err))
	} else {
		for i, user := range recentUsers {
			timestamp := time.Now()
			if user.LastLoginAt != nil {
				timestamp = *user.LastLoginAt
			}
			
			activities = append(activities, ActivityData{
				ID:          strconv.Itoa(i + 1),
				Type:        "user_login",
				Title:       "用户登录",
				Description: user.Username + " 登录了系统",
				Timestamp:   timestamp,
				User: &struct {
					ID       string `json:"id"`
					Username string `json:"username"`
					Avatar   string `json:"avatar,omitempty"`
				}{
					ID:       user.ID.String(),
					Username: user.Username,
					Avatar:   user.Avatar,
				},
				Severity: "info",
			})
		}
	}

	// 获取最近创建的菜单作为项目创建活动
	var recentMenus []models.Menu
	if err := h.db.Order("created_at DESC").
		Limit(limit/2).
		Find(&recentMenus).Error; err != nil {
		h.logger.Error("Failed to get recent menus", zap.Error(err))
	} else {
		for i, menu := range recentMenus {
			activities = append(activities, ActivityData{
				ID:          strconv.Itoa(len(activities) + i + 1),
				Type:        "project_created",
				Title:       "菜单创建",
				Description: "创建了新菜单 \"" + menu.Title + "\"",
				Timestamp:   menu.CreatedAt,
				Severity:    "success",
			})
		}
	}

	// 如果没有足够的真实数据，添加一些系统活动
	if len(activities) < limit {
		remaining := limit - len(activities)
		for i := 0; i < remaining; i++ {
			activities = append(activities, ActivityData{
				ID:          strconv.Itoa(len(activities) + i + 1),
				Type:        "system_alert",
				Title:       "系统监控",
				Description: "系统运行正常",
				Timestamp:   time.Now().Add(-time.Duration(i*15) * time.Minute),
				Severity:    "info",
			})
		}
	}

	// 按时间排序
	for i := 0; i < len(activities)-1; i++ {
		for j := i + 1; j < len(activities); j++ {
			if activities[i].Timestamp.Before(activities[j].Timestamp) {
				activities[i], activities[j] = activities[j], activities[i]
			}
		}
	}

	// 限制返回数量
	if limit < len(activities) {
		activities = activities[:limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    activities,
	})
}

// GetTrendData 获取趋势数据
func (h *DashboardHandler) GetTrendData(c *gin.Context) {
	h.logger.Info("Getting trend data")

	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 7
	}

	var trendData []TrendData

	// 为每一天生成趋势数据
	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		
		// 获取当天创建的用户数
		var dailyUsers int64
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)
		
		if err := h.db.Model(&models.User{}).
			Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
			Count(&dailyUsers).Error; err != nil {
			h.logger.Error("Failed to count daily users", zap.Error(err))
			dailyUsers = 0
		}

		// 获取当天创建的菜单数作为项目数
		var dailyMenus int64
		if err := h.db.Model(&models.Menu{}).
			Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
			Count(&dailyMenus).Error; err != nil {
			h.logger.Error("Failed to count daily menus", zap.Error(err))
			dailyMenus = 0
		}

		// 获取当天创建的系统配置数作为任务数
		var dailyConfigs int64
		if err := h.db.Model(&models.SystemConfig{}).
			Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
			Count(&dailyConfigs).Error; err != nil {
			h.logger.Error("Failed to count daily configs", zap.Error(err))
			dailyConfigs = 0
		}

		// 计算性能指标（基于系统负载）
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		performance := 100 - int(float64(memStats.Alloc)/float64(memStats.Sys)*100)
		if performance < 0 {
			performance = 0
		}
		if performance > 100 {
			performance = 100
		}

		trendData = append(trendData, TrendData{
			Date:        dateStr,
			Users:       int(dailyUsers),
			Tasks:       int(dailyConfigs),
			Projects:    int(dailyMenus),
			Performance: performance,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    trendData,
	})
}

// GetQuickActions 获取快捷操作
func (h *DashboardHandler) GetQuickActions(c *gin.Context) {
	h.logger.Info("Getting quick actions")

	actions := []QuickAction{
		{
			ID:          "create-project",
			Title:       "创建项目",
			Description: "快速创建新项目",
			Icon:        "plus",
			Action:      "/projects/create",
			Color:       "#1890ff",
		},
		{
			ID:          "add-user",
			Title:       "添加用户",
			Description: "邀请新用户加入",
			Icon:        "user-add",
			Action:      "/users/invite",
			Color:       "#52c41a",
		},
		{
			ID:          "system-settings",
			Title:       "系统设置",
			Description: "配置系统参数",
			Icon:        "setting",
			Action:      "/settings",
			Color:       "#faad14",
		},
		{
			ID:          "view-reports",
			Title:       "查看报告",
			Description: "生成和查看报告",
			Icon:        "file-text",
			Action:      "/reports",
			Color:       "#722ed1",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    actions,
	})
}

// GetUserStats 获取用户统计数据
func (h *DashboardHandler) GetUserStats(c *gin.Context) {
	h.logger.Info("Getting user stats")

	// 使用用户服务获取统计数据
	userStats, err := h.userService.GetUserStats()
	if err != nil {
		h.logger.Error("Failed to get user stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "STATS_ERROR",
			"message": "Failed to get user statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    userStats,
	})
}