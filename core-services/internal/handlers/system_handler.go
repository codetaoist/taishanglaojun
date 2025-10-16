package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "runtime"
    "strconv"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "go.uber.org/zap"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"

    "github.com/codetaoist/taishanglaojun/core-services/internal/models"
)

// SystemHandler 系统设置处理器
type SystemHandler struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewSystemHandler 创建系统设置处理器
func NewSystemHandler(db *gorm.DB, logger *zap.Logger) *SystemHandler {
	return &SystemHandler{
		db:     db,
		logger: logger,
	}
}

// 统一使用 models.SystemConfig，避免与数据库迁移定义不一致导致写入失败

// SystemLog 系统日志模型
type SystemLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Module    string    `json:"module"`
	UserID    string    `json:"user_id"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Extra     string    `json:"extra" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
}

// SystemBackup 系统备份模型
type SystemBackup struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	FilePath    string    `json:"file_path"`
	FileSize    int64     `json:"file_size"`
	Status      string    `json:"status" gorm:"default:pending"` // pending, completed, failed
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// UpdateSystemConfigRequest 更新系统配置请求
type UpdateSystemConfigRequest struct {
	Configs []ConfigItem `json:"configs" binding:"required"`
}

// ConfigItem 配置项
type ConfigItem struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value"`
}

// CreateBackupRequest 创建备份请求
type CreateBackupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// GetSystemConfig 获取系统配置
func (h *SystemHandler) GetSystemConfig(c *gin.Context) {
    category := c.Query("category")
    isPublic := c.Query("public")

    var configs []models.SystemConfig
    query := h.db.Model(&models.SystemConfig{})

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if isPublic == "true" {
		query = query.Where("is_public = ?", true)
	}

	if err := query.Find(&configs).Error; err != nil {
		h.logger.Error("Failed to get system config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get system config"})
		return
	}

	// 按分类组织配置
    configMap := make(map[string][]models.SystemConfig)
    for _, config := range configs {
        if configMap[config.Category] == nil {
            configMap[config.Category] = []models.SystemConfig{}
        }
        configMap[config.Category] = append(configMap[config.Category], config)
    }

	c.JSON(http.StatusOK, gin.H{
		"configs":    configs,
		"categories": configMap,
	})
}

// UpdateSystemConfig 更新系统配置
func (h *SystemHandler) UpdateSystemConfig(c *gin.Context) {
    var req UpdateSystemConfigRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 基本校验
    if len(req.Configs) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No configs provided"})
        return
    }

    // 开始事务
    tx := h.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    for _, configItem := range req.Configs {
        key := strings.TrimSpace(configItem.Key)
        if key == "" {
            // 跳过空 key
            continue
        }

        // 根据 key 前缀设置分类，默认 general
        category := "general"
        if idx := strings.Index(key, "."); idx > 0 {
            category = key[:idx]
        }

        // 使用 UPSERT 语义按 key 更新/创建，避免主键类型不匹配或唯一约束冲突
        cfg := models.SystemConfig{
            ID:       uuid.New(),
            Key:      key,
            Value:    configItem.Value,
            Type:     "string",
            Category: category,
        }

        if err := tx.Clauses(clause.OnConflict{
            Columns:   []clause.Column{{Name: "key"}},
            DoUpdates: clause.Assignments(map[string]interface{}{
                "value":      cfg.Value,
                "type":       cfg.Type,
                "category":   cfg.Category,
                "updated_at": time.Now(),
            }),
        }).Create(&cfg).Error; err != nil {
            tx.Rollback()
            h.logger.Error("Failed to upsert config",
                zap.String("key", key),
                zap.Error(err))
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update config"})
            return
        }
    }

    if err := tx.Commit().Error; err != nil {
        h.logger.Error("Failed to commit config update", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update config"})
        return
    }

	c.JSON(http.StatusOK, gin.H{"message": "System config updated successfully"})
}

// GetSystemInfo 获取系统信息
func (h *SystemHandler) GetSystemInfo(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 获取数据库统计
    var userCount, configCount int64
    h.db.Model(&models.SystemConfig{}).Count(&configCount)

	systemInfo := gin.H{
		"version":     "1.0.0",
		"go_version":  runtime.Version(),
		"goroutines":  runtime.NumGoroutine(),
		"memory": gin.H{
			"alloc":       m.Alloc,
			"total_alloc": m.TotalAlloc,
			"sys":         m.Sys,
			"heap_alloc":  m.HeapAlloc,
			"heap_sys":    m.HeapSys,
		},
		"database": gin.H{
			"users":   userCount,
			"configs": configCount,
		},
		"uptime": time.Since(time.Now().Add(-time.Hour * 24)), // 模拟运行时间
	}

	c.JSON(http.StatusOK, gin.H{"system_info": systemInfo})
}

// GetSystemStatus 获取系统状态
func (h *SystemHandler) GetSystemStatus(c *gin.Context) {
	// 检查数据库连接
	sqlDB, err := h.db.DB()
	dbStatus := "healthy"
	if err != nil {
		dbStatus = "error"
	} else if err := sqlDB.Ping(); err != nil {
		dbStatus = "error"
	}

	// 检查内存使用
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memoryUsage := float64(m.Alloc) / float64(m.Sys) * 100

	status := gin.H{
		"overall": "healthy",
		"components": gin.H{
			"database": gin.H{
				"status": dbStatus,
				"message": "Database connection is working",
			},
			"memory": gin.H{
				"status": func() string {
					if memoryUsage > 90 {
						return "warning"
					}
					return "healthy"
				}(),
				"usage": memoryUsage,
			},
			"disk": gin.H{
				"status": "healthy",
				"usage":  45.2, // 模拟磁盘使用率
			},
		},
		"timestamp": time.Now(),
	}

	c.JSON(http.StatusOK, status)
}

// GetSystemLogs 获取系统日志
func (h *SystemHandler) GetSystemLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	level := c.Query("level")
	module := c.Query("module")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	var logs []SystemLog
	var total int64

	query := h.db.Model(&SystemLog{})
	if level != "" {
		query = query.Where("level = ?", level)
	}
	if module != "" {
		query = query.Where("module = ?", module)
	}

    // 获取总数
    if err := query.Count(&total).Error; err != nil {
        // 当表不存在时（MySQL: Error 1146），返回空结果而非500
        if strings.Contains(strings.ToLower(err.Error()), "1146") || strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
            h.logger.Warn("system_logs table not found, returning empty logs")
            c.JSON(http.StatusOK, gin.H{
                "logs":  []SystemLog{},
                "total": 0,
                "page":  page,
                "limit": limit,
                "pages": 0,
                "message": "system_logs table missing; returning empty list",
            })
            return
        }
        h.logger.Error("Failed to count logs", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logs"})
        return
    }

	// 获取日志列表
    if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
        if strings.Contains(strings.ToLower(err.Error()), "1146") || strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
            h.logger.Warn("system_logs table not found on query, returning empty logs")
            c.JSON(http.StatusOK, gin.H{
                "logs":  []SystemLog{},
                "total": 0,
                "page":  page,
                "limit": limit,
                "pages": 0,
                "message": "system_logs table missing; returning empty list",
            })
            return
        }
        h.logger.Error("Failed to query logs", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logs"})
        return
    }

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
		"pages": (total + int64(limit) - 1) / int64(limit),
	})
}

// GetSystemLogStats 获取系统日志统计
func (h *SystemHandler) GetSystemLogStats(c *gin.Context) {
    type StatItem struct {
        Level  string `json:"level"`
        Module string `json:"module"`
        Count  int64  `json:"count"`
    }

    var stats []StatItem
    // 统计不同级别和模块的数量
    err := h.db.Model(&SystemLog{}).
        Select("COALESCE(level, '') AS level, COALESCE(module, '') AS module, COUNT(*) AS count").
        Group("level, module").
        Order("count DESC").
        Scan(&stats).Error
    if err != nil {
        if strings.Contains(strings.ToLower(err.Error()), "1146") || strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
            h.logger.Warn("system_logs table not found, returning zero stats")
            c.JSON(http.StatusOK, gin.H{
                "stats":       []StatItem{},
                "error_count": 0,
                "today_count": 0,
                "message":     "system_logs table missing; returning zero stats",
            })
            return
        }
        h.logger.Error("Failed to aggregate log stats", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get log stats"})
        return
    }

    // 额外统计：错误日志数量、今日日志数量
    var errorCount int64
    var todayCount int64
    if err := h.db.Model(&SystemLog{}).Where("level = ?", "error").Count(&errorCount).Error; err != nil {
        if strings.Contains(strings.ToLower(err.Error()), "1146") || strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
            errorCount = 0
        }
    }
    if err := h.db.Model(&SystemLog{}).Where("created_at >= ?", time.Now().Add(-24*time.Hour)).Count(&todayCount).Error; err != nil {
        if strings.Contains(strings.ToLower(err.Error()), "1146") || strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
            todayCount = 0
        }
    }

    c.JSON(http.StatusOK, gin.H{
        "stats":       stats,
        "error_count": errorCount,
        "today_count": todayCount,
    })
}

// CreateBackup 创建系统备份
func (h *SystemHandler) CreateBackup(c *gin.Context) {
	var req CreateBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户ID（从JWT中间件）
	userID, exists := c.Get("user_id")
	if !exists {
		userID = "system"
	}

	backup := SystemBackup{
		Name:        req.Name,
		Description: req.Description,
		FilePath:    "/backups/" + req.Name + "_" + time.Now().Format("20060102_150405") + ".sql",
		Status:      "pending",
		CreatedBy:   userID.(string),
	}

	if err := h.db.Create(&backup).Error; err != nil {
		h.logger.Error("Failed to create backup record", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create backup"})
		return
	}

	// 这里应该启动实际的备份过程（异步）
	go func() {
		// 模拟备份过程
		time.Sleep(5 * time.Second)
		
		// 更新备份状态
		now := time.Now()
		backup.Status = "completed"
		backup.FileSize = 1024 * 1024 * 10 // 模拟10MB文件
		backup.CompletedAt = &now
		h.db.Save(&backup)
	}()

	c.JSON(http.StatusCreated, gin.H{
		"message": "Backup started successfully",
		"backup":  backup,
	})
}

// GetBackups 获取备份列表
func (h *SystemHandler) GetBackups(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var backups []SystemBackup
	var total int64

	// 获取总数
	if err := h.db.Model(&SystemBackup{}).Count(&total).Error; err != nil {
		h.logger.Error("Failed to count backups", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get backups"})
		return
	}

	// 获取备份列表
	if err := h.db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&backups).Error; err != nil {
		h.logger.Error("Failed to query backups", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get backups"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"backups": backups,
		"total":   total,
		"page":    page,
		"limit":   limit,
		"pages":   (total + int64(limit) - 1) / int64(limit),
	})
}

// RestoreBackup 恢复备份
func (h *SystemHandler) RestoreBackup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid backup ID"})
		return
	}

	var backup SystemBackup
	if err := h.db.First(&backup, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Backup not found"})
			return
		}
		h.logger.Error("Failed to get backup", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore backup"})
		return
	}

	if backup.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Backup is not completed"})
		return
	}

	// 这里应该启动实际的恢复过程（异步）
	go func() {
		// 模拟恢复过程
		h.logger.Info("Starting backup restore", zap.String("backup_id", strconv.Itoa(int(backup.ID))))
		time.Sleep(10 * time.Second)
		h.logger.Info("Backup restore completed", zap.String("backup_id", strconv.Itoa(int(backup.ID))))
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Backup restore started successfully",
		"backup":  backup,
	})
}

// StartMaintenance 开始维护模式
func (h *SystemHandler) StartMaintenance(c *gin.Context) {
    // 设置维护模式配置（使用 models.SystemConfig 并生成 UUID 主键）
    config := models.SystemConfig{
        ID:       uuid.New(),
        Key:      "maintenance_mode",
        Value:    "true",
        Type:     "boolean",
        Category: "system",
    }

	if err := h.db.Where("key = ?", config.Key).Assign(config).FirstOrCreate(&config).Error; err != nil {
		h.logger.Error("Failed to enable maintenance mode", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enable maintenance mode"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Maintenance mode enabled"})
}

// StopMaintenance 停止维护模式
func (h *SystemHandler) StopMaintenance(c *gin.Context) {
    // 设置维护模式配置（使用 models.SystemConfig 并生成 UUID 主键）
    config := models.SystemConfig{
        ID:       uuid.New(),
        Key:      "maintenance_mode",
        Value:    "false",
        Type:     "boolean",
        Category: "system",
    }

	if err := h.db.Where("key = ?", config.Key).Assign(config).FirstOrCreate(&config).Error; err != nil {
		h.logger.Error("Failed to disable maintenance mode", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable maintenance mode"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Maintenance mode disabled"})
}

// GetMaintenanceStatus 获取维护状态
func (h *SystemHandler) GetMaintenanceStatus(c *gin.Context) {
    var config models.SystemConfig
    if err := h.db.Where("key = ?", "maintenance_mode").First(&config).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusOK, gin.H{
                "maintenance_mode": false,
                "message":          "System is operational",
            })
            return
        }
        h.logger.Error("Failed to get maintenance status", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get maintenance status"})
        return
    }

	isMaintenanceMode := config.Value == "true"
	c.JSON(http.StatusOK, gin.H{
		"maintenance_mode": isMaintenanceMode,
		"message": func() string {
			if isMaintenanceMode {
				return "System is in maintenance mode"
			}
			return "System is operational"
		}(),
	})
}

// ClearCache 清除缓存
func (h *SystemHandler) ClearCache(c *gin.Context) {
	// 这里应该实现实际的缓存清除逻辑
	// 例如清除Redis缓存、内存缓存等
	
	c.JSON(http.StatusOK, gin.H{"message": "Cache cleared successfully"})
}

// GetCacheStats 获取缓存统计
func (h *SystemHandler) GetCacheStats(c *gin.Context) {
	// 模拟缓存统计数据
	stats := gin.H{
		"total_keys":   1250,
		"memory_usage": "45.2MB",
		"hit_rate":     "94.5%",
		"miss_rate":    "5.5%",
		"evictions":    23,
		"connections":  15,
	}

	c.JSON(http.StatusOK, gin.H{"cache_stats": stats})
}

// GetDatabaseStats 获取数据库统计
func (h *SystemHandler) GetDatabaseStats(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil {
		h.logger.Error("Failed to get database connection", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database stats"})
		return
	}

	stats := sqlDB.Stats()
	
	dbStats := gin.H{
		"open_connections":     stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration.String(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}

	c.JSON(http.StatusOK, gin.H{"database_stats": dbStats})
}

// OptimizeDatabase 优化数据库
func (h *SystemHandler) OptimizeDatabase(c *gin.Context) {
	// 这里应该实现实际的数据库优化逻辑
	// 例如重建索引、清理过期数据等
	
	go func() {
		h.logger.Info("Starting database optimization")
		time.Sleep(30 * time.Second) // 模拟优化过程
		h.logger.Info("Database optimization completed")
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Database optimization started"})
}

// ListDBTables 列出数据库中的表（支持 MySQL 与 PostgreSQL）
func (h *SystemHandler) ListDBTables(c *gin.Context) {
    // 分页参数
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
    if page < 1 { page = 1 }
    if limit < 1 || limit > 200 { limit = 50 }
    offset := (page - 1) * limit

    sqlDB, err := h.db.DB()
    if err != nil {
        h.logger.Error("Failed to get raw DB", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access database"})
        return
    }

    // 优先使用客户端提供的 schema；如果未提供且为 MySQL，则回退到当前数据库名
    schemaParam, hasSchema := c.GetQuery("schema")
    schema := strings.TrimSpace(schemaParam)
    dialect := h.db.Dialector.Name()
    if !hasSchema || schema == "" {
        if dialect == "mysql" {
            var currentDB string
            if err := sqlDB.QueryRow("SELECT DATABASE()").Scan(&currentDB); err == nil && strings.TrimSpace(currentDB) != "" {
                schema = currentDB
            } else {
                // 如果无法获取当前数据库名，保持为空以返回合理的错误或空列表
                schema = ""
            }
        } else {
            schema = "public"
        }
    }

    // 统计总数与查询列表，支持 MySQL 与 PostgreSQL
    var total int64
    var countQuery string
    var listQuery string
    if dialect == "mysql" {
        countQuery = "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_type='BASE TABLE'"
        listQuery = "SELECT table_schema, table_name FROM information_schema.tables WHERE table_schema = ? AND table_type='BASE TABLE' ORDER BY table_schema, table_name LIMIT ? OFFSET ?"
    } else {
        countQuery = "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = $1 AND table_type='BASE TABLE'"
        listQuery = `
        SELECT table_schema, table_name
        FROM information_schema.tables
        WHERE table_schema = $1 AND table_type='BASE TABLE'
        ORDER BY table_schema, table_name
        LIMIT $2 OFFSET $3`
    }

    if err := sqlDB.QueryRow(countQuery, schema).Scan(&total); err != nil {
        h.logger.Error("Failed to count tables", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tables"})
        return
    }

    rows, err := sqlDB.Query(listQuery, schema, limit, offset)
    if err != nil {
        h.logger.Error("Failed to query tables", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tables"})
        return
    }
    defer rows.Close()

    type TableInfo struct {
        Schema string `json:"schema"`
        Name   string `json:"name"`
    }
    tables := make([]TableInfo, 0, limit)
    for rows.Next() {
        var ti TableInfo
        if err := rows.Scan(&ti.Schema, &ti.Name); err != nil {
            h.logger.Error("Failed to scan table row", zap.Error(err))
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tables"})
            return
        }
        tables = append(tables, ti)
    }

    c.JSON(http.StatusOK, gin.H{
        "tables": tables,
        "total":  total,
        "page":   page,
        "limit":  limit,
        "pages":  (total + int64(limit) - 1) / int64(limit),
    })
}

// ListSchemas 列出数据库中的 schema（支持 MySQL 与 PostgreSQL）
func (h *SystemHandler) ListSchemas(c *gin.Context) {
    // 分页参数
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
    if page < 1 { page = 1 }
    if limit < 1 || limit > 200 { limit = 50 }
    offset := (page - 1) * limit

    sqlDB, err := h.db.DB()
    if err != nil {
        h.logger.Error("Failed to get raw DB", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access database"})
        return
    }

    dialect := h.db.Dialector.Name()

    var total int64
    var countQuery string
    var listQuery string
    if dialect == "mysql" {
        countQuery = "SELECT COUNT(*) FROM information_schema.schemata"
        listQuery = "SELECT schema_name FROM information_schema.schemata ORDER BY schema_name LIMIT ? OFFSET ?"
    } else {
        countQuery = "SELECT COUNT(*) FROM pg_namespace WHERE nspname NOT LIKE 'pg_%' AND nspname != 'information_schema'"
        listQuery = `
        SELECT nspname AS schema_name
        FROM pg_namespace
        WHERE nspname NOT LIKE 'pg_%' AND nspname != 'information_schema'
        ORDER BY nspname
        LIMIT $1 OFFSET $2`
    }

    if err := sqlDB.QueryRow(countQuery).Scan(&total); err != nil {
        h.logger.Error("Failed to count schemas", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list schemas"})
        return
    }

    rows, err := sqlDB.Query(listQuery, limit, offset)
    if err != nil {
        h.logger.Error("Failed to query schemas", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list schemas"})
        return
    }
    defer rows.Close()

    schemas := make([]string, 0, limit)
    for rows.Next() {
        var name string
        if err := rows.Scan(&name); err != nil {
            h.logger.Error("Failed to scan schema row", zap.Error(err))
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list schemas"})
            return
        }
        schemas = append(schemas, name)
    }

    c.JSON(http.StatusOK, gin.H{
        "schemas": schemas,
        "total":   total,
        "page":    page,
        "limit":   limit,
        "pages":   (total + int64(limit) - 1) / int64(limit),
    })
}

// GetTableColumns 获取表的列信息（支持 MySQL 与 PostgreSQL）
func (h *SystemHandler) GetTableColumns(c *gin.Context) {
    // 优先使用客户端提供的 schema；如果未提供且为 MySQL，则回退到当前数据库名
    schemaParam, hasSchema := c.GetQuery("schema")
    schema := strings.TrimSpace(schemaParam)
    table := c.Param("name")
    if strings.TrimSpace(table) == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table name"})
        return
    }

    sqlDB, err := h.db.DB()
    if err != nil {
        h.logger.Error("Failed to get raw DB", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access database"})
        return
    }

    if (schema == "" || !hasSchema) && h.db.Dialector.Name() == "mysql" {
        var currentDB string
        if err := sqlDB.QueryRow("SELECT DATABASE()").Scan(&currentDB); err == nil && strings.TrimSpace(currentDB) != "" {
            schema = currentDB
        } else {
            schema = ""
        }
    }

    var query string
    if h.db.Dialector.Name() == "mysql" {
        query = `
        SELECT column_name, data_type, is_nullable, character_maximum_length
        FROM information_schema.columns
        WHERE table_schema = ? AND table_name = ?
        ORDER BY ordinal_position`
    } else {
        query = `
        SELECT column_name, data_type, is_nullable, character_maximum_length
        FROM information_schema.columns
        WHERE table_schema = $1 AND table_name = $2
        ORDER BY ordinal_position`
    }

    rows, err := sqlDB.Query(query, schema, table)
    if err != nil {
        h.logger.Error("Failed to query columns", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get columns"})
        return
    }
    defer rows.Close()

    type ColumnInfo struct {
        Name     string      `json:"name"`
        Type     string      `json:"type"`
        Nullable bool        `json:"nullable"`
        Length   interface{} `json:"length"`
    }
    cols := []ColumnInfo{}
    for rows.Next() {
        var name, dtype, nullableStr string
        var length sql.NullInt64
        if err := rows.Scan(&name, &dtype, &nullableStr, &length); err != nil {
            h.logger.Error("Failed to scan column row", zap.Error(err))
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get columns"})
            return
        }
        cols = append(cols, ColumnInfo{
            Name:     name,
            Type:     dtype,
            Nullable: strings.ToLower(nullableStr) == "yes",
            Length:   func() interface{} { if length.Valid { return length.Int64 }; return nil }(),
        })
    }

    c.JSON(http.StatusOK, gin.H{"columns": cols})
}

// RunReadOnlyQuery 执行只读查询（仅允许 SELECT）
func (h *SystemHandler) RunReadOnlyQuery(c *gin.Context) {
    var req struct {
        Query   string `json:"query" binding:"required"`
        MaxRows int    `json:"max_rows"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    q := strings.TrimSpace(req.Query)
    lower := strings.ToLower(q)
    if !strings.HasPrefix(lower, "select") {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Only SELECT queries are allowed"})
        return
    }
    // 简单防止多语句
    if strings.Count(lower, ";") > 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Semicolons are not allowed"})
        return
    }

    if req.MaxRows <= 0 || req.MaxRows > 1000 {
        req.MaxRows = 200
    }

    sqlDB, err := h.db.DB()
    if err != nil {
        h.logger.Error("Failed to get raw DB", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access database"})
        return
    }

    // 限制返回行数
    qWithLimit := q + " LIMIT " + strconv.Itoa(req.MaxRows)
    rows, err := sqlDB.Query(qWithLimit)
    if err != nil {
        h.logger.Error("Failed to execute query", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Query execution failed"})
        return
    }
    defer rows.Close()

    cols, err := rows.Columns()
    if err != nil {
        h.logger.Error("Failed to get columns from result", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Query execution failed"})
        return
    }

    // 扫描数据到通用结构
    results := make([]map[string]interface{}, 0, req.MaxRows)
    for rows.Next() {
        vals := make([]interface{}, len(cols))
        valPtrs := make([]interface{}, len(cols))
        for i := range vals { valPtrs[i] = &vals[i] }
        if err := rows.Scan(valPtrs...); err != nil {
            h.logger.Error("Failed to scan query row", zap.Error(err))
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Query execution failed"})
            return
        }
        rowMap := map[string]interface{}{}
        for i, col := range cols {
            v := vals[i]
            // 处理 []byte -> string 便于显示
            if b, ok := v.([]byte); ok {
                rowMap[col] = string(b)
            } else {
                rowMap[col] = v
            }
        }
        results = append(results, rowMap)
    }

    c.JSON(http.StatusOK, gin.H{
        "columns": cols,
        "rows":    results,
        "count":   len(results),
    })
}

// DetectIssues 基于系统错误日志进行简单问题检测与分类
func (h *SystemHandler) DetectIssues(c *gin.Context) {
    // 最近24小时的错误日志
    since := time.Now().Add(-24 * time.Hour)
    var logs []SystemLog
    if err := h.db.Where("level = ? AND created_at >= ?", "error", since).Order("created_at DESC").Limit(1000).Find(&logs).Error; err != nil {
        if strings.Contains(strings.ToLower(err.Error()), "1146") || strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
            h.logger.Warn("system_logs table not found, returning empty issues")
            c.JSON(http.StatusOK, gin.H{"issues": []interface{}{}, "count": 0, "message": "system_logs table missing; returning empty issues"})
            return
        }
        h.logger.Error("Failed to query error logs", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to detect issues"})
        return
    }

    type Issue struct {
        ID          string      `json:"id"`
        Title       string      `json:"title"`
        Category    string      `json:"category"` // frontend/backend/database/security
        Severity    string      `json:"severity"` // low/medium/high/critical
        Occurrences int         `json:"occurrences"`
        FirstSeen   time.Time   `json:"first_seen"`
        LastSeen    time.Time   `json:"last_seen"`
        Example     string      `json:"example"`
        Suggestions []string    `json:"suggestions"`
    }

    groups := map[string]*Issue{}
    for _, l := range logs {
        key := normalizeMessageKey(l.Message)
        if _, ok := groups[key]; !ok {
            groups[key] = &Issue{
                ID:       uuid.NewString(),
                Title:    truncate(l.Message, 80),
                Category: classifyMessage(l.Message, l.Module),
                Severity: "low",
                Occurrences: 0,
                FirstSeen:   l.CreatedAt,
                LastSeen:    l.CreatedAt,
                Example:     l.Message,
                Suggestions: suggestFixes(l.Message),
            }
        }
        g := groups[key]
        g.Occurrences++
        if l.CreatedAt.Before(g.FirstSeen) { g.FirstSeen = l.CreatedAt }
        if l.CreatedAt.After(g.LastSeen) { g.LastSeen = l.CreatedAt }
    }

    // 评估严重程度
    issues := make([]*Issue, 0, len(groups))
    for _, g := range groups {
        switch {
        case g.Occurrences >= 100:
            g.Severity = "critical"
        case g.Occurrences >= 50:
            g.Severity = "high"
        case g.Occurrences >= 10:
            g.Severity = "medium"
        default:
            g.Severity = "low"
        }
        issues = append(issues, g)
    }

    c.JSON(http.StatusOK, gin.H{"issues": issues, "count": len(issues)})
}

// CreateSystemLog 写入系统日志（供前端/客户端上报）
func (h *SystemHandler) CreateSystemLog(c *gin.Context) {
    // 请求体定义，extra 支持任意对象
    var req struct {
        Level     string      `json:"level"`
        Message   string      `json:"message" binding:"required"`
        Module    string      `json:"module"`
        Extra     interface{} `json:"extra"`
        Timestamp string      `json:"timestamp"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    level := strings.ToLower(strings.TrimSpace(req.Level))
    if level == "" {
        level = "info"
    }
    switch level {
    case "debug", "info", "warn", "warning", "error":
        if level == "warning" {
            level = "warn"
        }
    default:
        level = "info"
    }

    msg := strings.TrimSpace(req.Message)
    if msg == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "message required"})
        return
    }

    var extraStr string
    if req.Extra != nil {
        if b, err := json.Marshal(req.Extra); err == nil {
            extraStr = string(b)
        } else {
            extraStr = "{}"
        }
    }

    // 从 JWT 中间件注入的上下文获取用户信息
    userID := ""
    if uid, ok := c.Get("user_id"); ok {
        if idStr, ok2 := uid.(string); ok2 {
            userID = idStr
        }
    }

    ip := c.ClientIP()
    ua := c.Request.UserAgent()

    logRecord := SystemLog{
        Level:     strings.ToUpper(level),
        Message:   msg,
        Module:    strings.TrimSpace(req.Module),
        UserID:    userID,
        IP:        ip,
        UserAgent: ua,
        Extra:     extraStr,
        CreatedAt: time.Now(),
    }

    if err := h.db.Create(&logRecord).Error; err != nil {
        lower := strings.ToLower(err.Error())
        // 如果表不存在，尝试自动迁移一次
        if strings.Contains(lower, "1146") || strings.Contains(lower, "does not exist") || strings.Contains(lower, "no such table") || strings.Contains(lower, "missing") {
            h.logger.Warn("system_logs table missing on insert, attempting auto-migrate")
            if migErr := h.db.AutoMigrate(&SystemLog{}); migErr != nil {
                h.logger.Error("Auto-migrate system_logs failed", zap.Error(migErr))
                c.JSON(http.StatusOK, gin.H{"created": false, "message": "system_logs table missing; auto-migrate failed"})
                return
            }
            if retryErr := h.db.Create(&logRecord).Error; retryErr != nil {
                h.logger.Error("Failed to insert system log after migrate", zap.Error(retryErr))
                c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write system log"})
                return
            }
        } else {
            h.logger.Error("Failed to write system log", zap.Error(err))
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write system log"})
            return
        }
    }

    c.JSON(http.StatusOK, gin.H{"created": true, "log": logRecord})
}

// TriggerIssueAlert 触发问题告警（示例实现）
func (h *SystemHandler) TriggerIssueAlert(c *gin.Context) {
    var req struct {
        IssueID  string   `json:"issue_id"`
        Channels []string `json:"channels"` // email, sms, system
        Message  string   `json:"message"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if req.IssueID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "issue_id required"})
        return
    }

    // 简单记录日志，模拟发送
    h.logger.Info("Trigger issue alert", zap.String("issue_id", req.IssueID), zap.Strings("channels", req.Channels), zap.String("message", req.Message))
    c.JSON(http.StatusOK, gin.H{"message": "Alert triggered", "channels": req.Channels})
}

// --- 辅助函数 ---
func normalizeMessageKey(s string) string {
    s = strings.TrimSpace(strings.ToLower(s))
    // 去掉数字，稳定分组（例如包含ID或时间戳的错误）
    b := strings.Builder{}
    for _, r := range s {
        if r < '0' || r > '9' {
            b.WriteRune(r)
        }
    }
    return b.String()
}

func truncate(s string, n int) string {
    if len(s) <= n { return s }
    return s[:n] + "..."
}

func classifyMessage(msg, module string) string {
    m := strings.ToLower(msg)
    if strings.Contains(m, "sql") || strings.Contains(m, "database") || strings.Contains(m, "pq:") {
        return "database"
    }
    if strings.Contains(m, "network") || strings.Contains(m, "timeout") || strings.Contains(m, "http") {
        return "backend"
    }
    if strings.Contains(m, "react") || strings.Contains(m, "javascript") || strings.Contains(m, "element is undefined") {
        return "frontend"
    }
    if strings.Contains(m, "permission") || strings.Contains(m, "forbidden") || strings.Contains(m, "unauthorized") {
        return "security"
    }
    if module != "" { return strings.ToLower(module) }
    return "backend"
}

func suggestFixes(msg string) []string {
    m := strings.ToLower(msg)
    fixes := []string{"查看详细日志和堆栈以定位问题"}
    if strings.Contains(m, "timeout") {
        fixes = append(fixes, "检查网络连接和服务可用性", "增加超时或优化查询")
    }
    if strings.Contains(m, "sql") || strings.Contains(m, "pq:") {
        fixes = append(fixes, "检查SQL语法和表结构", "验证索引和约束是否正确")
    }
    if strings.Contains(m, "permission") || strings.Contains(m, "forbidden") {
        fixes = append(fixes, "检查用户权限配置", "验证角色与策略是否正确应用")
    }
    return fixes
}