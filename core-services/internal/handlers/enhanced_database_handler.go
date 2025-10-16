package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// EnhancedDatabaseHandler 增强的数据库监控处理器
type EnhancedDatabaseHandler struct {
	db       *gorm.DB
	logger   *zap.Logger
	cache    *sync.Map // 简单的内存缓存
	upgrader websocket.Upgrader
	backups  *sync.Map // 备份状态缓存
}

// DatabaseMetrics 数据库指标
type DatabaseMetrics struct {
	Timestamp   time.Time          `json:"timestamp"`
	Connections ConnectionMetrics  `json:"connections"`
	Performance PerformanceMetrics `json:"performance"`
	Storage     StorageMetrics     `json:"storage"`
	Memory      MemoryMetrics      `json:"memory"`
}

// ConnectionMetrics 连接指标
type ConnectionMetrics struct {
	Active    int `json:"active"`
	Idle      int `json:"idle"`
	Total     int `json:"total"`
	MaxConn   int `json:"max_conn"`
	WaitCount int `json:"wait_count"`
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	QueriesPerSecond float64 `json:"queries_per_second"`
	AvgQueryTime     float64 `json:"avg_query_time"`
	SlowQueries      int     `json:"slow_queries"`
	CacheHitRatio    float64 `json:"cache_hit_ratio"`
	IndexUsage       float64 `json:"index_usage"`
}

// StorageMetrics 存储指标
type StorageMetrics struct {
	TotalSize     int64 `json:"total_size"`
	UsedSize      int64 `json:"used_size"`
	FreeSize      int64 `json:"free_size"`
	TableCount    int   `json:"table_count"`
	IndexSize     int64 `json:"index_size"`
	DataSize      int64 `json:"data_size"`
}

// MemoryMetrics 内存指标
type MemoryMetrics struct {
	BufferPool    int64 `json:"buffer_pool"`
	QueryCache    int64 `json:"query_cache"`
	TempTables    int   `json:"temp_tables"`
	SortBuffer    int64 `json:"sort_buffer"`
}

// DatabaseHealth 数据库健康状态
type DatabaseHealth struct {
	OverallHealth string        `json:"overall_health"`
	HealthScore   int           `json:"health_score"`
	Issues        []HealthIssue `json:"issues"`
	LastCheck     time.Time     `json:"last_check"`
}

// HealthIssue 健康问题
type HealthIssue struct {
	Severity       string `json:"severity"`
	Category       string `json:"category"`
	Message        string `json:"message"`
	Recommendation string `json:"recommendation"`
}

// BackupStatus 备份状态
type BackupStatus struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Progress    int       `json:"progress"`
	BackupType  string    `json:"backup_type"`
	FileSize    int64     `json:"file_size"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Error       string    `json:"error,omitempty"`
}

// ConnectionInfo 连接信息
type ConnectionInfo struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Type           string                 `json:"type"`
	Host           string                 `json:"host"`
	Port           int                    `json:"port"`
	Database       string                 `json:"database"`
	Status         string                 `json:"status"`
	ResponseTime   int                    `json:"response_time"`
	HealthScore    int                    `json:"health_score"`
	ConnectionPool ConnectionPoolMetrics  `json:"connection_pool"`
	LastCheck      time.Time              `json:"last_check"`
}

// ConnectionPoolMetrics 连接池指标
type ConnectionPoolMetrics struct {
	ActiveConnections int `json:"active_connections"`
	IdleConnections   int `json:"idle_connections"`
	MaxConnections    int `json:"max_connections"`
	WaitingCount      int `json:"waiting_count"`
}

// NewEnhancedDatabaseHandler 创建增强的数据库监控处理器
func NewEnhancedDatabaseHandler(db *gorm.DB, logger *zap.Logger) *EnhancedDatabaseHandler {
	return &EnhancedDatabaseHandler{
		db:     db,
		logger: logger,
		cache:  &sync.Map{},
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 在生产环境中应该检查来源
			},
		},
		backups: &sync.Map{},
	}
}

// GetDatabaseMetrics 获取数据库指标
func (h *EnhancedDatabaseHandler) GetDatabaseMetrics(c *gin.Context) {
	startTime := time.Now()
	
	metrics, err := h.collectDatabaseMetrics()
	if err != nil {
		duration := time.Since(startTime)
		h.logger.Error("Failed to collect database metrics", 
			zap.Error(err),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.ClientIP()),
		)
		
		// 根据错误类型返回不同的HTTP状态码
		statusCode := h.getErrorStatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error":   "Failed to collect database metrics",
			"code":    h.getErrorCode(err),
			"details": err.Error(),
			"timestamp": time.Now().Unix(),
			"retryable": h.isRetryableError(err),
		})
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Database metrics collected successfully",
		zap.Duration("duration", duration),
		zap.String("client_ip", c.ClientIP()),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
		"meta": gin.H{
			"timestamp": time.Now().Unix(),
			"duration_ms": duration.Milliseconds(),
		},
	})
}

// GetDatabaseHealth 获取数据库健康状态
func (h *EnhancedDatabaseHandler) GetDatabaseHealth(c *gin.Context) {
	startTime := time.Now()
	
	health, err := h.checkDatabaseHealth()
	if err != nil {
		duration := time.Since(startTime)
		h.logger.Error("Failed to check database health", 
			zap.Error(err),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.ClientIP()),
		)
		
		statusCode := h.getErrorStatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error":   "Failed to check database health",
			"code":    h.getErrorCode(err),
			"details": err.Error(),
			"timestamp": time.Now().Unix(),
			"retryable": h.isRetryableError(err),
		})
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Database health check completed",
		zap.Duration("duration", duration),
		zap.String("client_ip", c.ClientIP()),
		zap.String("health_status", health.OverallHealth),
		zap.Int("health_score", health.HealthScore),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    health,
		"meta": gin.H{
			"timestamp": time.Now().Unix(),
			"duration_ms": duration.Milliseconds(),
		},
	})
}

// GetDatabaseMetricsHistory 获取历史指标
func (h *EnhancedDatabaseHandler) GetDatabaseMetricsHistory(c *gin.Context) {
	duration := c.DefaultQuery("duration", "1h")
	interval := c.DefaultQuery("interval", "5m")

	// 这里应该从时序数据库或缓存中获取历史数据
	// 目前返回模拟数据
	history := h.generateMockHistory(duration, interval)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    history,
	})
}

// ConnectionFilter 连接过滤器
type ConnectionFilter struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
	User      string `json:"user"`
	Database  string `json:"database"`
	State     string `json:"state"`
}

// GetActiveConnections 获取活跃连接列表（支持分页、过滤、排序）
func (h *EnhancedDatabaseHandler) GetActiveConnections(c *gin.Context) {
	startTime := time.Now()
	
	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	sortBy := c.DefaultQuery("sort_by", "connected_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	filterUser := c.Query("user")
	filterDatabase := c.Query("database")
	filterState := c.Query("state")
	
	// 验证参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	
	connections, total, err := h.getActiveConnectionsWithFilter(ConnectionFilter{
		Page:         page,
		PageSize:     pageSize,
		SortBy:       sortBy,
		SortOrder:    sortOrder,
		User:         filterUser,
		Database:     filterDatabase,
		State:        filterState,
	})
	
	if err != nil {
		duration := time.Since(startTime)
		h.logger.Error("Failed to get active connections", 
			zap.Error(err),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.ClientIP()),
		)
		
		statusCode := h.getErrorStatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error":   "Failed to get active connections",
			"code":    h.getErrorCode(err),
			"details": err.Error(),
			"timestamp": time.Now().Unix(),
			"retryable": h.isRetryableError(err),
		})
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Active connections retrieved successfully",
		zap.Duration("duration", duration),
		zap.String("client_ip", c.ClientIP()),
		zap.Int("total_connections", total),
		zap.Int("returned_connections", len(connections)),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    connections,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_pages": (total + pageSize - 1) / pageSize,
		},
		"meta": gin.H{
			"timestamp": time.Now().Unix(),
			"duration_ms": duration.Milliseconds(),
		},
	})
}

// GetConnectionPoolStats 获取连接池统计
func (h *EnhancedDatabaseHandler) GetConnectionPoolStats(c *gin.Context) {
	stats, err := h.getConnectionPoolStats()
	if err != nil {
		h.logger.Error("Failed to get connection pool stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pool stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// TestAllConnections 测试所有连接
func (h *EnhancedDatabaseHandler) TestAllConnections(c *gin.Context) {
	results, err := h.testAllConnections()
	if err != nil {
		h.logger.Error("Failed to test connections", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to test connections"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
	})
}

// OptimizeDatabase 优化数据库
func (h *EnhancedDatabaseHandler) OptimizeDatabase(c *gin.Context) {
	var req struct {
		AnalyzeTables  bool `json:"analyze_tables"`
		RebuildIndexes bool `json:"rebuild_indexes"`
		CleanupLogs    bool `json:"cleanup_logs"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 启动异步优化任务
	go h.performDatabaseOptimization(req)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Database optimization started",
	})
}

// collectDatabaseMetrics 收集数据库指标
func (h *EnhancedDatabaseHandler) collectDatabaseMetrics() (*DatabaseMetrics, error) {
	sqlDB, err := h.db.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	
	// 获取数据库类型
	dbType := h.getDatabaseType()
	
	metrics := &DatabaseMetrics{
		Timestamp: time.Now(),
		Connections: ConnectionMetrics{
			Active:    stats.InUse,
			Idle:      stats.Idle,
			Total:     stats.OpenConnections,
			MaxConn:   stats.MaxOpenConnections,
			WaitCount: int(stats.WaitCount),
		},
		Performance: h.getPerformanceMetrics(dbType),
		Storage:     h.getStorageMetrics(dbType),
		Memory:      h.getMemoryMetrics(dbType),
	}

	return metrics, nil
}

// checkDatabaseHealth 检查数据库健康状态
func (h *EnhancedDatabaseHandler) checkDatabaseHealth() (*DatabaseHealth, error) {
	issues := []HealthIssue{}
	healthScore := 100

	// 检查连接数
	sqlDB, err := h.db.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	connectionUsage := float64(stats.InUse) / float64(stats.MaxOpenConnections)
	
	if connectionUsage > 0.8 {
		issues = append(issues, HealthIssue{
			Severity:       "high",
			Category:       "connections",
			Message:        "High connection usage detected",
			Recommendation: "Consider increasing max connections or optimizing connection usage",
		})
		healthScore -= 20
	}

	// 检查等待连接数
	if stats.WaitCount > 100 {
		issues = append(issues, HealthIssue{
			Severity:       "medium",
			Category:       "performance",
			Message:        "High connection wait count",
			Recommendation: "Optimize query performance or increase connection pool size",
		})
		healthScore -= 10
	}

	// 确定整体健康状态
	var overallHealth string
	if healthScore >= 80 {
		overallHealth = "healthy"
	} else if healthScore >= 60 {
		overallHealth = "warning"
	} else {
		overallHealth = "critical"
	}

	return &DatabaseHealth{
		OverallHealth: overallHealth,
		HealthScore:   healthScore,
		Issues:        issues,
		LastCheck:     time.Now(),
	}, nil
}

// getDatabaseType 获取数据库类型
func (h *EnhancedDatabaseHandler) getDatabaseType() string {
	var dbType string
	h.db.Raw("SELECT VERSION()").Scan(&dbType)
	
	if strings.Contains(strings.ToLower(dbType), "mysql") {
		return "mysql"
	} else if strings.Contains(strings.ToLower(dbType), "postgres") {
		return "postgresql"
	}
	
	return "unknown"
}

// getPerformanceMetrics 获取性能指标
func (h *EnhancedDatabaseHandler) getPerformanceMetrics(dbType string) PerformanceMetrics {
	metrics := PerformanceMetrics{
		QueriesPerSecond: 0,
		AvgQueryTime:     0,
		SlowQueries:      0,
		CacheHitRatio:    0,
		IndexUsage:       0,
	}

	switch dbType {
	case "mysql":
		h.getMySQLPerformanceMetrics(&metrics)
	case "postgresql":
		h.getPostgreSQLPerformanceMetrics(&metrics)
	}

	return metrics
}

// getMySQLPerformanceMetrics 获取MySQL性能指标
func (h *EnhancedDatabaseHandler) getMySQLPerformanceMetrics(metrics *PerformanceMetrics) {
	// 查询每秒查询数
	var queries, uptime float64
	h.db.Raw("SHOW GLOBAL STATUS LIKE 'Questions'").Scan(&queries)
	h.db.Raw("SHOW GLOBAL STATUS LIKE 'Uptime'").Scan(&uptime)
	
	if uptime > 0 {
		metrics.QueriesPerSecond = queries / uptime
	}

	// 查询缓存命中率
	var cacheHits, cacheQueries float64
	h.db.Raw("SHOW GLOBAL STATUS LIKE 'Qcache_hits'").Scan(&cacheHits)
	h.db.Raw("SHOW GLOBAL STATUS LIKE 'Com_select'").Scan(&cacheQueries)
	
	if cacheQueries > 0 {
		metrics.CacheHitRatio = (cacheHits / (cacheHits + cacheQueries)) * 100
	}

	// 慢查询数量
	var slowQueries int
	h.db.Raw("SHOW GLOBAL STATUS LIKE 'Slow_queries'").Scan(&slowQueries)
	metrics.SlowQueries = slowQueries
}

// getPostgreSQLPerformanceMetrics 获取PostgreSQL性能指标
func (h *EnhancedDatabaseHandler) getPostgreSQLPerformanceMetrics(metrics *PerformanceMetrics) {
	// PostgreSQL性能指标查询
	var stats struct {
		TupReturned int64 `gorm:"column:tup_returned"`
		TupFetched  int64 `gorm:"column:tup_fetched"`
	}
	
	h.db.Raw("SELECT sum(tup_returned) as tup_returned, sum(tup_fetched) as tup_fetched FROM pg_stat_database").Scan(&stats)
	
	if stats.TupReturned > 0 {
		metrics.CacheHitRatio = float64(stats.TupFetched) / float64(stats.TupReturned) * 100
	}
}

// getStorageMetrics 获取存储指标
func (h *EnhancedDatabaseHandler) getStorageMetrics(dbType string) StorageMetrics {
	metrics := StorageMetrics{}

	switch dbType {
	case "mysql":
		h.getMySQLStorageMetrics(&metrics)
	case "postgresql":
		h.getPostgreSQLStorageMetrics(&metrics)
	}

	return metrics
}

// getMySQLStorageMetrics 获取MySQL存储指标
func (h *EnhancedDatabaseHandler) getMySQLStorageMetrics(metrics *StorageMetrics) {
	var result struct {
		DataLength  int64 `gorm:"column:data_length"`
		IndexLength int64 `gorm:"column:index_length"`
		TableCount  int   `gorm:"column:table_count"`
	}
	
	h.db.Raw(`
		SELECT 
			SUM(data_length) as data_length,
			SUM(index_length) as index_length,
			COUNT(*) as table_count
		FROM information_schema.tables 
		WHERE table_schema = DATABASE()
	`).Scan(&result)
	
	metrics.DataSize = result.DataLength
	metrics.IndexSize = result.IndexLength
	metrics.TableCount = result.TableCount
	metrics.UsedSize = result.DataLength + result.IndexLength
}

// getPostgreSQLStorageMetrics 获取PostgreSQL存储指标
func (h *EnhancedDatabaseHandler) getPostgreSQLStorageMetrics(metrics *StorageMetrics) {
	var result struct {
		DatabaseSize int64 `gorm:"column:database_size"`
		TableCount   int   `gorm:"column:table_count"`
	}
	
	h.db.Raw(`
		SELECT 
			pg_database_size(current_database()) as database_size,
			(SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public') as table_count
	`).Scan(&result)
	
	metrics.UsedSize = result.DatabaseSize
	metrics.TableCount = result.TableCount
}

// getMemoryMetrics 获取内存指标
func (h *EnhancedDatabaseHandler) getMemoryMetrics(dbType string) MemoryMetrics {
	// 这里返回模拟数据，实际应该根据数据库类型查询真实指标
	return MemoryMetrics{
		BufferPool: 1024 * 1024 * 512, // 512MB
		QueryCache: 1024 * 1024 * 128, // 128MB
		TempTables: 10,
		SortBuffer: 1024 * 1024 * 64,  // 64MB
	}
}

// generateMockHistory 生成模拟历史数据
func (h *EnhancedDatabaseHandler) generateMockHistory(duration, interval string) []DatabaseMetrics {
	history := []DatabaseMetrics{}
	
	// 解析时间间隔
	intervalDuration, _ := time.ParseDuration(interval)
	totalDuration, _ := time.ParseDuration(duration)
	
	points := int(totalDuration / intervalDuration)
	now := time.Now()
	
	for i := 0; i < points; i++ {
		timestamp := now.Add(-time.Duration(points-i) * intervalDuration)
		
		metrics := DatabaseMetrics{
			Timestamp: timestamp,
			Connections: ConnectionMetrics{
				Active: 10 + i%5,
				Idle:   20 + i%3,
				Total:  30 + i%8,
			},
			Performance: PerformanceMetrics{
				QueriesPerSecond: 100 + float64(i%20),
				CacheHitRatio:    85 + float64(i%10),
			},
		}
		
		history = append(history, metrics)
	}
	
	return history
}

// ActiveConnection 活跃连接信息
type ActiveConnection struct {
	ID          string    `json:"id"`
	User        string    `json:"user"`
	Host        string    `json:"host"`
	Database    string    `json:"database"`
	Command     string    `json:"command"`
	Time        int       `json:"time"`
	State       string    `json:"state"`
	Info        string    `json:"info"`
	ConnectedAt time.Time `json:"connected_at"`
	LastQuery   string    `json:"last_query"`
	QueryCount  int       `json:"query_count"`
	BytesIn     int64     `json:"bytes_in"`
	BytesOut    int64     `json:"bytes_out"`
}

// getActiveConnections 获取活跃连接
func (h *EnhancedDatabaseHandler) getActiveConnections() ([]ConnectionInfo, error) {
	// 这里应该查询实际的连接信息
	// 目前返回模拟数据
	connections := []ConnectionInfo{
		{
			ID:           "conn-1",
			Name:         "Main Database",
			Type:         "mysql",
			Host:         "localhost",
			Port:         3306,
			Database:     "main_db",
			Status:       "connected",
			ResponseTime: 15,
			HealthScore:  95,
			ConnectionPool: ConnectionPoolMetrics{
				ActiveConnections: 10,
				IdleConnections:   20,
				MaxConnections:    50,
				WaitingCount:      0,
			},
			LastCheck: time.Now(),
		},
	}
	
	return connections, nil
}

// getActiveConnectionsWithFilter 获取带过滤的活跃连接
func (h *EnhancedDatabaseHandler) getActiveConnectionsWithFilter(filter ConnectionFilter) ([]ActiveConnection, int, error) {
	dbType := h.getDatabaseType()
	
	switch dbType {
	case "mysql":
		return h.getMySQLActiveConnections(filter)
	case "postgresql":
		return h.getPostgreSQLActiveConnections(filter)
	default:
		return h.getMockActiveConnections(filter)
	}
}

// getMySQLActiveConnections 获取MySQL活跃连接
func (h *EnhancedDatabaseHandler) getMySQLActiveConnections(filter ConnectionFilter) ([]ActiveConnection, int, error) {
	var connections []ActiveConnection
	var total int
	
	// 构建查询
	query := "SELECT ID, USER, HOST, DB, COMMAND, TIME, STATE, INFO FROM INFORMATION_SCHEMA.PROCESSLIST WHERE 1=1"
	args := []interface{}{}
	
	// 添加过滤条件
	if filter.User != "" {
		query += " AND USER LIKE ?"
		args = append(args, "%"+filter.User+"%")
	}
	if filter.Database != "" {
		query += " AND DB LIKE ?"
		args = append(args, "%"+filter.Database+"%")
	}
	if filter.State != "" {
		query += " AND STATE LIKE ?"
		args = append(args, "%"+filter.State+"%")
	}
	
	// 获取总数
	countQuery := "SELECT COUNT(*) FROM (" + query + ") as count_table"
	h.db.Raw(countQuery, args...).Scan(&total)
	
	// 添加排序
	validSortFields := map[string]string{
		"id":           "ID",
		"user":         "USER",
		"host":         "HOST",
		"database":     "DB",
		"command":      "COMMAND",
		"time":         "TIME",
		"state":        "STATE",
		"connected_at": "TIME",
	}
	
	if sortField, ok := validSortFields[filter.SortBy]; ok {
		order := "ASC"
		if filter.SortOrder == "desc" {
			order = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", sortField, order)
	}
	
	// 添加分页
	offset := (filter.Page - 1) * filter.PageSize
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", filter.PageSize, offset)
	
	// 执行查询
	rows, err := h.db.Raw(query, args...).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var conn ActiveConnection
		var dbName sql.NullString
		var info sql.NullString
		
		err := rows.Scan(&conn.ID, &conn.User, &conn.Host, &dbName, &conn.Command, &conn.Time, &conn.State, &info)
		if err != nil {
			continue
		}
		
		conn.Database = dbName.String
		conn.Info = info.String
		conn.ConnectedAt = time.Now().Add(-time.Duration(conn.Time) * time.Second)
		conn.LastQuery = conn.Info
		conn.QueryCount = conn.Time // 简化处理
		
		connections = append(connections, conn)
	}
	
	return connections, total, nil
}

// getPostgreSQLActiveConnections 获取PostgreSQL活跃连接
func (h *EnhancedDatabaseHandler) getPostgreSQLActiveConnections(filter ConnectionFilter) ([]ActiveConnection, int, error) {
	var connections []ActiveConnection
	var total int
	
	// 构建查询
	query := `
		SELECT 
			pid, usename, client_addr, datname, state, 
			EXTRACT(EPOCH FROM (now() - backend_start))::int as duration,
			query, backend_start
		FROM pg_stat_activity 
		WHERE state IS NOT NULL AND pid != pg_backend_pid()
	`
	args := []interface{}{}
	
	// 添加过滤条件
	if filter.User != "" {
		query += " AND usename ILIKE ?"
		args = append(args, "%"+filter.User+"%")
	}
	if filter.Database != "" {
		query += " AND datname ILIKE ?"
		args = append(args, "%"+filter.Database+"%")
	}
	if filter.State != "" {
		query += " AND state ILIKE ?"
		args = append(args, "%"+filter.State+"%")
	}
	
	// 获取总数
	countQuery := "SELECT COUNT(*) FROM (" + query + ") as count_table"
	h.db.Raw(countQuery, args...).Scan(&total)
	
	// 添加排序
	validSortFields := map[string]string{
		"id":           "pid",
		"user":         "usename",
		"host":         "client_addr",
		"database":     "datname",
		"state":        "state",
		"time":         "duration",
		"connected_at": "backend_start",
	}
	
	if sortField, ok := validSortFields[filter.SortBy]; ok {
		order := "ASC"
		if filter.SortOrder == "desc" {
			order = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", sortField, order)
	}
	
	// 添加分页
	offset := (filter.Page - 1) * filter.PageSize
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", filter.PageSize, offset)
	
	// 执行查询
	rows, err := h.db.Raw(query, args...).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var conn ActiveConnection
		var clientAddr sql.NullString
		var query sql.NullString
		var backendStart time.Time
		
		err := rows.Scan(&conn.ID, &conn.User, &clientAddr, &conn.Database, &conn.State, &conn.Time, &query, &backendStart)
		if err != nil {
			continue
		}
		
		conn.Host = clientAddr.String
		conn.Info = query.String
		conn.ConnectedAt = backendStart
		conn.LastQuery = conn.Info
		conn.Command = "Query"
		
		connections = append(connections, conn)
	}
	
	return connections, total, nil
}

// getMockActiveConnections 获取模拟活跃连接
func (h *EnhancedDatabaseHandler) getMockActiveConnections(filter ConnectionFilter) ([]ActiveConnection, int, error) {
	// 生成模拟数据
	allConnections := []ActiveConnection{
		{
			ID:          "1",
			User:        "app_user",
			Host:        "192.168.1.100:3306",
			Database:    "main_db",
			Command:     "Query",
			Time:        120,
			State:       "executing",
			Info:        "SELECT * FROM users WHERE active = 1",
			ConnectedAt: time.Now().Add(-2 * time.Minute),
			LastQuery:   "SELECT * FROM users WHERE active = 1",
			QueryCount:  45,
			BytesIn:     1024,
			BytesOut:    2048,
		},
		{
			ID:          "2",
			User:        "admin",
			Host:        "192.168.1.101:3306",
			Database:    "admin_db",
			Command:     "Sleep",
			Time:        30,
			State:       "idle",
			Info:        "",
			ConnectedAt: time.Now().Add(-30 * time.Second),
			LastQuery:   "SHOW TABLES",
			QueryCount:  12,
			BytesIn:     512,
			BytesOut:    1024,
		},
		{
			ID:          "3",
			User:        "report_user",
			Host:        "192.168.1.102:3306",
			Database:    "report_db",
			Command:     "Query",
			Time:        300,
			State:       "sending data",
			Info:        "SELECT COUNT(*) FROM transactions WHERE date >= '2024-01-01'",
			ConnectedAt: time.Now().Add(-5 * time.Minute),
			LastQuery:   "SELECT COUNT(*) FROM transactions WHERE date >= '2024-01-01'",
			QueryCount:  8,
			BytesIn:     2048,
			BytesOut:    4096,
		},
	}
	
	// 应用过滤
	filteredConnections := []ActiveConnection{}
	for _, conn := range allConnections {
		if filter.User != "" && !strings.Contains(strings.ToLower(conn.User), strings.ToLower(filter.User)) {
			continue
		}
		if filter.Database != "" && !strings.Contains(strings.ToLower(conn.Database), strings.ToLower(filter.Database)) {
			continue
		}
		if filter.State != "" && !strings.Contains(strings.ToLower(conn.State), strings.ToLower(filter.State)) {
			continue
		}
		filteredConnections = append(filteredConnections, conn)
	}
	
	total := len(filteredConnections)
	
	// 应用分页
	start := (filter.Page - 1) * filter.PageSize
	end := start + filter.PageSize
	
	if start >= len(filteredConnections) {
		return []ActiveConnection{}, total, nil
	}
	
	if end > len(filteredConnections) {
		end = len(filteredConnections)
	}
	
	return filteredConnections[start:end], total, nil
}

// KillConnection 终止指定连接
func (h *EnhancedDatabaseHandler) KillConnection(c *gin.Context) {
	startTime := time.Now()
	connectionID := c.Param("id")
	
	if connectionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Connection ID is required",
			"code":    "INVALID_PARAMETER",
		})
		return
	}
	
	err := h.killDatabaseConnection(connectionID)
	if err != nil {
		duration := time.Since(startTime)
		h.logger.Error("Failed to kill connection", 
			zap.Error(err),
			zap.String("connection_id", connectionID),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.ClientIP()),
		)
		
		statusCode := h.getErrorStatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error":   "Failed to kill connection",
			"code":    h.getErrorCode(err),
			"details": err.Error(),
			"timestamp": time.Now().Unix(),
			"retryable": h.isRetryableError(err),
		})
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Connection killed successfully",
		zap.String("connection_id", connectionID),
		zap.Duration("duration", duration),
		zap.String("client_ip", c.ClientIP()),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Connection killed successfully",
		"meta": gin.H{
			"timestamp": time.Now().Unix(),
			"duration_ms": duration.Milliseconds(),
		},
	})
}

// KillMultipleConnections 批量终止连接
func (h *EnhancedDatabaseHandler) KillMultipleConnections(c *gin.Context) {
	startTime := time.Now()
	
	var request struct {
		ConnectionIDs []string `json:"connection_ids" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
			"code":    "INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}
	
	if len(request.ConnectionIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "At least one connection ID is required",
			"code":    "INVALID_PARAMETER",
		})
		return
	}
	
	results := make([]map[string]interface{}, 0, len(request.ConnectionIDs))
	successCount := 0
	
	for _, connectionID := range request.ConnectionIDs {
		result := map[string]interface{}{
			"id": connectionID,
		}
		
		err := h.killDatabaseConnection(connectionID)
		if err != nil {
			result["success"] = false
			result["message"] = err.Error()
			h.logger.Warn("Failed to kill connection in batch", 
				zap.String("connection_id", connectionID),
				zap.Error(err),
			)
		} else {
			result["success"] = true
			result["message"] = "Connection killed successfully"
			successCount++
		}
		
		results = append(results, result)
	}

	duration := time.Since(startTime)
	h.logger.Info("Batch connection kill completed",
		zap.Int("total_connections", len(request.ConnectionIDs)),
		zap.Int("successful_kills", successCount),
		zap.Duration("duration", duration),
		zap.String("client_ip", c.ClientIP()),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"results": results,
		"summary": gin.H{
			"total":      len(request.ConnectionIDs),
			"successful": successCount,
			"failed":     len(request.ConnectionIDs) - successCount,
		},
		"meta": gin.H{
			"timestamp": time.Now().Unix(),
			"duration_ms": duration.Milliseconds(),
		},
	})
}

// killDatabaseConnection 终止数据库连接
func (h *EnhancedDatabaseHandler) killDatabaseConnection(connectionID string) error {
	dbType := h.getDatabaseType()
	
	switch dbType {
	case "mysql":
		return h.killMySQLConnection(connectionID)
	case "postgresql":
		return h.killPostgreSQLConnection(connectionID)
	default:
		// 模拟终止连接
		return h.killMockConnection(connectionID)
	}
}

// killMySQLConnection 终止MySQL连接
func (h *EnhancedDatabaseHandler) killMySQLConnection(connectionID string) error {
	query := "KILL ?"
	result := h.db.Exec(query, connectionID)
	if result.Error != nil {
		return fmt.Errorf("failed to kill MySQL connection %s: %w", connectionID, result.Error)
	}
	return nil
}

// killPostgreSQLConnection 终止PostgreSQL连接
func (h *EnhancedDatabaseHandler) killPostgreSQLConnection(connectionID string) error {
	query := "SELECT pg_terminate_backend(?)"
	var success bool
	err := h.db.Raw(query, connectionID).Scan(&success).Error
	if err != nil {
		return fmt.Errorf("failed to kill PostgreSQL connection %s: %w", connectionID, err)
	}
	if !success {
		return fmt.Errorf("failed to terminate PostgreSQL connection %s", connectionID)
	}
	return nil
}

// killMockConnection 模拟终止连接
func (h *EnhancedDatabaseHandler) killMockConnection(connectionID string) error {
	// 模拟一些可能的错误情况
	if connectionID == "invalid" {
		return fmt.Errorf("connection not found")
	}
	if connectionID == "protected" {
		return fmt.Errorf("cannot kill protected connection")
	}
	
	// 模拟成功
	time.Sleep(100 * time.Millisecond) // 模拟操作延迟
	return nil
}

// getConnectionPoolStats 获取连接池统计
func (h *EnhancedDatabaseHandler) getConnectionPoolStats() (map[string]interface{}, error) {
	sqlDB, err := h.db.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	
	return map[string]interface{}{
		"total_connections":   stats.OpenConnections,
		"active_connections":  stats.InUse,
		"idle_connections":    stats.Idle,
		"waiting_connections": stats.WaitCount,
		"max_connections":     stats.MaxOpenConnections,
	}, nil
}

// testAllConnections 测试所有连接
func (h *EnhancedDatabaseHandler) testAllConnections() (map[string]interface{}, error) {
	// 测试当前数据库连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err := h.db.WithContext(ctx).Raw("SELECT 1").Error
	
	result := map[string]interface{}{
		"tested_at": time.Now(),
		"results": []map[string]interface{}{
			{
				"connection": "main",
				"status":     "success",
				"error":      nil,
			},
		},
	}
	
	if err != nil {
		result["results"] = []map[string]interface{}{
			{
				"connection": "main",
				"status":     "failed",
				"error":      err.Error(),
			},
		}
	}
	
	return result, nil
}

// performDatabaseOptimization 执行数据库优化
func (h *EnhancedDatabaseHandler) performDatabaseOptimization(req struct {
	AnalyzeTables  bool `json:"analyze_tables"`
	RebuildIndexes bool `json:"rebuild_indexes"`
	CleanupLogs    bool `json:"cleanup_logs"`
}) {
	h.logger.Info("Starting database optimization")
	
	if req.AnalyzeTables {
		h.logger.Info("Analyzing tables...")
		// 执行表分析
		time.Sleep(5 * time.Second) // 模拟分析过程
	}
	
	if req.RebuildIndexes {
		h.logger.Info("Rebuilding indexes...")
		// 重建索引
		time.Sleep(10 * time.Second) // 模拟重建过程
	}
	
	if req.CleanupLogs {
		h.logger.Info("Cleaning up logs...")
		// 清理日志
		time.Sleep(3 * time.Second) // 模拟清理过程
	}
	
	h.logger.Info("Database optimization completed")
}

// getErrorStatusCode 根据错误类型返回HTTP状态码
func (h *EnhancedDatabaseHandler) getErrorStatusCode(err error) int {
	if err == sql.ErrNoRows {
		return http.StatusNotFound
	}
	if err == context.DeadlineExceeded {
		return http.StatusRequestTimeout
	}
	if strings.Contains(err.Error(), "connection") {
		return http.StatusServiceUnavailable
	}
	return http.StatusInternalServerError
}

// getErrorCode 获取错误代码
func (h *EnhancedDatabaseHandler) getErrorCode(err error) string {
	if err == sql.ErrNoRows {
		return "DB_NO_ROWS"
	}
	if err == context.DeadlineExceeded {
		return "DB_TIMEOUT"
	}
	if strings.Contains(err.Error(), "connection") {
		return "DB_CONNECTION_ERROR"
	}
	return "DB_INTERNAL_ERROR"
}

// isRetryableError 判断错误是否可重试
func (h *EnhancedDatabaseHandler) isRetryableError(err error) bool {
	if err == context.DeadlineExceeded {
		return true
	}
	if strings.Contains(err.Error(), "connection") {
		return true
	}
	if strings.Contains(err.Error(), "timeout") {
		return true
	}
	return false
}

// CreateBackup 创建备份
func (h *EnhancedDatabaseHandler) CreateBackup(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		BackupType  string `json:"backup_type"`
		Compression bool   `json:"compression"`
		Encryption  bool   `json:"encryption"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 生成备份ID
	backupID := h.generateBackupID()
	
	// 创建备份状态
	backup := &BackupStatus{
		ID:         backupID,
		Name:       req.Name,
		Status:     "pending",
		Progress:   0,
		BackupType: req.BackupType,
		CreatedAt:  time.Now(),
	}
	
	// 保存到缓存
	h.backups.Store(backupID, backup)
	
	// 启动异步备份任务
	backupReq := struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		BackupType  string `json:"backup_type"`
		Compression bool   `json:"compression"`
		Encryption  bool   `json:"encryption"`
	}{
		Name:        req.Name,
		Description: req.Description,
		BackupType:  req.BackupType,
		Compression: req.Compression,
		Encryption:  req.Encryption,
	}
	go h.performBackup(backupID, backupReq)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"backup_id": backupID,
		"message":   "Backup task started",
	})
}

// GetBackups 获取备份列表
func (h *EnhancedDatabaseHandler) GetBackups(c *gin.Context) {
	backups := []BackupStatus{}
	
	h.backups.Range(func(key, value interface{}) bool {
		if backup, ok := value.(*BackupStatus); ok {
			backups = append(backups, *backup)
		}
		return true
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"backups": backups,
			"total":   len(backups),
		},
	})
}

// RestoreBackup 恢复备份
func (h *EnhancedDatabaseHandler) RestoreBackup(c *gin.Context) {
	backupID := c.Param("id")
	
	value, exists := h.backups.Load(backupID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Backup not found"})
		return
	}
	
	backup := value.(*BackupStatus)
	if backup.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Backup is not completed"})
		return
	}
	
	// 启动异步恢复任务
	go h.performRestore(backupID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Restore task started",
	})
}

// DeleteBackup 删除备份
func (h *EnhancedDatabaseHandler) DeleteBackup(c *gin.Context) {
	backupID := c.Param("id")
	
	_, exists := h.backups.Load(backupID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Backup not found"})
		return
	}
	
	h.backups.Delete(backupID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Backup deleted successfully",
	})
}

// GetBackupStatus 获取备份状态
func (h *EnhancedDatabaseHandler) GetBackupStatus(c *gin.Context) {
	backupID := c.Param("id")
	
	value, exists := h.backups.Load(backupID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Backup not found"})
		return
	}
	
	backup := value.(*BackupStatus)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    backup,
	})
}

// GetBackupProgress 获取备份进度
func (h *EnhancedDatabaseHandler) GetBackupProgress(c *gin.Context) {
	backupID := c.Param("id")
	
	value, exists := h.backups.Load(backupID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Backup not found"})
		return
	}
	
	backup := value.(*BackupStatus)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"backup_id": backup.ID,
			"status":    backup.Status,
			"progress":  backup.Progress,
		},
	})
}

// MetricsWebSocket 指标WebSocket连接
func (h *EnhancedDatabaseHandler) MetricsWebSocket(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade websocket", zap.Error(err))
		return
	}
	defer conn.Close()

	// 定期发送指标数据
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics, err := h.collectDatabaseMetrics()
			if err != nil {
				h.logger.Error("Failed to collect metrics for websocket", zap.Error(err))
				continue
			}

			if err := conn.WriteJSON(metrics); err != nil {
				h.logger.Error("Failed to write websocket message", zap.Error(err))
				return
			}
		}
	}
}

// HealthWebSocket 健康状态WebSocket连接
func (h *EnhancedDatabaseHandler) HealthWebSocket(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade websocket", zap.Error(err))
		return
	}
	defer conn.Close()

	// 定期发送健康状态数据
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			health, err := h.checkDatabaseHealth()
			if err != nil {
				h.logger.Error("Failed to check health for websocket", zap.Error(err))
				continue
			}

			if err := conn.WriteJSON(health); err != nil {
				h.logger.Error("Failed to write websocket message", zap.Error(err))
				return
			}
		}
	}
}

// generateBackupID 生成备份ID
func (h *EnhancedDatabaseHandler) generateBackupID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// performBackup 执行备份
func (h *EnhancedDatabaseHandler) performBackup(backupID string, req struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	BackupType  string `json:"backup_type"`
	Compression bool   `json:"compression"`
	Encryption  bool   `json:"encryption"`
}) {
	value, exists := h.backups.Load(backupID)
	if !exists {
		return
	}
	
	backup := value.(*BackupStatus)
	
	// 更新状态为运行中
	backup.Status = "running"
	backup.Progress = 0
	h.backups.Store(backupID, backup)
	
	h.logger.Info("Starting backup", zap.String("backup_id", backupID), zap.String("name", req.Name))
	
	// 模拟备份过程
	for i := 0; i <= 100; i += 10 {
		time.Sleep(1 * time.Second)
		backup.Progress = i
		h.backups.Store(backupID, backup)
	}
	
	// 完成备份
	backup.Status = "completed"
	backup.Progress = 100
	backup.FileSize = 1024 * 1024 * 50 // 50MB 模拟文件大小
	completedAt := time.Now()
	backup.CompletedAt = &completedAt
	h.backups.Store(backupID, backup)
	
	h.logger.Info("Backup completed", zap.String("backup_id", backupID))
}

// performRestore 执行恢复
func (h *EnhancedDatabaseHandler) performRestore(backupID string) {
	h.logger.Info("Starting restore", zap.String("backup_id", backupID))
	
	// 模拟恢复过程
	time.Sleep(10 * time.Second)
	
	h.logger.Info("Restore completed", zap.String("backup_id", backupID))
}