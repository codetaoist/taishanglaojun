package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/taishanglaojun/auth_system/internal/config"
	"github.com/taishanglaojun/auth_system/internal/database"
	"go.uber.org/zap"
)

// DatabaseHandler 动态数据库管理处理器
type DatabaseHandler struct {
	dynamicDB *database.DynamicDatabase
	logger    *zap.Logger
}

// NewDatabaseHandler 创建数据库处理器
func NewDatabaseHandler(dynamicDB *database.DynamicDatabase, logger *zap.Logger) *DatabaseHandler {
	return &DatabaseHandler{
		dynamicDB: dynamicDB,
		logger:    logger,
	}
}

// AddDatabaseRequest 添加数据库请求结构
type AddDatabaseRequest struct {
	Name     string                `json:"name" binding:"required"`
	Database config.DatabaseConfig `json:"database" binding:"required"`
	Redis    config.RedisConfig    `json:"redis"`
}

// ListDatabases 获取数据库配置列表
func (h *DatabaseHandler) ListDatabases(c *gin.Context) {
	databases := h.dynamicDB.ListDatabases()
	current := h.dynamicDB.GetCurrentDatabaseName()
	
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"data": gin.H{
			"databases": databases,
			"current":   current,
			"total":     len(databases),
		},
	})
}

// GetCurrentDatabase 获取当前数据库配置
func (h *DatabaseHandler) GetCurrentDatabase(c *gin.Context) {
	current := h.dynamicDB.GetCurrentDatabaseName()
	if current == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "No active database configuration",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"current": current,
		},
	})
}

// AddDatabase 添加数据库配置
func (h *DatabaseHandler) AddDatabase(c *gin.Context) {
	var req AddDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		return
	}

	// 验证数据库名称
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Database name is required",
		})
		return
	}

	// 创建配置
	cfg := &config.Config{
		Database: req.Database,
		Redis:    req.Redis,
	}

	// 添加数据库配置
	if err := h.dynamicDB.AddDatabase(req.Name, cfg); err != nil {
		h.logger.Error("Failed to add database configuration",
			zap.String("name", req.Name),
			zap.Error(err))
		
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	h.logger.Info("Database configuration added successfully",
		zap.String("name", req.Name),
		zap.String("type", req.Database.Type))

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Database configuration added successfully",
		"data": gin.H{
			"name": req.Name,
		},
	})
}

// SwitchDatabase 切换数据库
func (h *DatabaseHandler) SwitchDatabase(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Database name is required",
		})
		return
	}

	// 切换数据库
	if err := h.dynamicDB.SwitchDatabase(name); err != nil {
		h.logger.Error("Failed to switch database",
			zap.String("name", name),
			zap.Error(err))
		
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	h.logger.Info("Database switched successfully", zap.String("name", name))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Database switched successfully",
		"data": gin.H{
			"current": name,
		},
	})
}

// RemoveDatabase 删除数据库配置
func (h *DatabaseHandler) RemoveDatabase(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Database name is required",
		})
		return
	}

	// 删除数据库配置
	if err := h.dynamicDB.RemoveDatabase(name); err != nil {
		h.logger.Error("Failed to remove database configuration",
			zap.String("name", name),
			zap.Error(err))
		
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	h.logger.Info("Database configuration removed successfully", zap.String("name", name))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Database configuration removed successfully",
	})
}

// HealthCheck 数据库健康检查
func (h *DatabaseHandler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results := h.dynamicDB.HealthCheck(ctx)
	
	// 统计健康状态
	healthy := 0
	unhealthy := 0
	details := make(map[string]interface{})
	
	for name, err := range results {
		if err == nil {
			healthy++
			details[name] = gin.H{
				"status": "healthy",
				"error":  nil,
			}
		} else {
			unhealthy++
			details[name] = gin.H{
				"status": "unhealthy",
				"error":  err.Error(),
			}
		}
	}

	status := http.StatusOK
	if unhealthy > 0 {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"success": unhealthy == 0,
		"data": gin.H{
			"summary": gin.H{
				"total":     len(results),
				"healthy":   healthy,
				"unhealthy": unhealthy,
			},
			"details": details,
		},
	})
}

// GetDatabaseStats 获取数据库统计信息
func (h *DatabaseHandler) GetDatabaseStats(c *gin.Context) {
	stats := h.dynamicDB.GetDatabaseStats()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}