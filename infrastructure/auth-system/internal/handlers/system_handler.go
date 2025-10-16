package handlers

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/monitoring"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/repository"
)

// SystemHandler 系统监控处理器
type SystemHandler struct {
	db          *gorm.DB
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	metrics     *monitoring.Metrics
	logger      *zap.Logger
}

// NewSystemHandler 创建系统监控处理器实例
func NewSystemHandler(
	db *gorm.DB,
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	metrics *monitoring.Metrics,
	logger *zap.Logger,
) *SystemHandler {
	return &SystemHandler{
		db:          db,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		metrics:     metrics,
		logger:      logger,
	}
}

// HealthCheck 健康检查
// @Summary 健康检查
// @Description 检查系统健康状态
// @Tags system
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} ErrorResponse
// @Router /system/health [get]
func (h *SystemHandler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	health := &HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Services:  make(map[string]ServiceHealth),
	}

	// 检查数据库连接
	if err := h.checkDatabase(ctx); err != nil {
		health.Services["database"] = ServiceHealth{
			Status: "unhealthy",
			Error:  err.Error(),
		}
		health.Status = "unhealthy"
	} else {
		health.Services["database"] = ServiceHealth{
			Status: "healthy",
		}
	}

	// 检查用户仓库
	if err := h.checkUserRepository(ctx); err != nil {
		health.Services["user_repository"] = ServiceHealth{
			Status: "unhealthy",
			Error:  err.Error(),
		}
		health.Status = "unhealthy"
	} else {
		health.Services["user_repository"] = ServiceHealth{
			Status: "healthy",
		}
	}

	// 检查会话仓库
	if err := h.checkSessionRepository(ctx); err != nil {
		health.Services["session_repository"] = ServiceHealth{
			Status: "unhealthy",
			Error:  err.Error(),
		}
		health.Status = "unhealthy"
	} else {
		health.Services["session_repository"] = ServiceHealth{
			Status: "healthy",
		}
	}

	if health.Status == "unhealthy" {
		c.JSON(http.StatusServiceUnavailable, health)
	} else {
		c.JSON(http.StatusOK, health)
	}
}

// GetSystemInfo 获取系统信息
// @Summary 获取系统信息
// @Description 获取系统运行时信息
// @Tags system
// @Produce json
// @Success 200 {object} SystemInfoResponse
// @Failure 500 {object} ErrorResponse
// @Router /system/info [get]
func (h *SystemHandler) GetSystemInfo(c *gin.Context) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	info := &SystemInfoResponse{
		Version:   "1.0.0", // 可以从配置或构建信息中获取
		BuildTime: time.Now().Format("2006-01-02 15:04:05"), // 可以从构建信息中获取
		GoVersion: runtime.Version(),
		Runtime: RuntimeInfo{
			Goroutines:   runtime.NumGoroutine(),
			CPUs:         runtime.NumCPU(),
			MemoryAlloc:  memStats.Alloc,
			MemoryTotal:  memStats.TotalAlloc,
			MemorySys:    memStats.Sys,
			GCRuns:       memStats.NumGC,
			LastGCTime:   time.Unix(0, int64(memStats.LastGC)),
		},
		Uptime:    time.Since(startTime),
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, info)
}

// GetMetrics 获取系统指标
// @Summary 获取系统指标
// @Description 获取系统性能指标
// @Tags system
// @Produce json
// @Success 200 {object} MetricsResponse
// @Failure 500 {object} ErrorResponse
// @Router /system/metrics [get]
func (h *SystemHandler) GetMetrics(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取用户统计
	totalUsers, err := h.userRepo.Count(ctx)
	if err != nil {
		h.logger.Error("Failed to get user count", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get metrics",
		})
		return
	}

	// 获取会话统计
	totalSessions, err := h.sessionRepo.Count(ctx)
	if err != nil {
		h.logger.Error("Failed to get session count", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get metrics",
		})
		return
	}

	activeSessions, err := h.sessionRepo.CountByStatus(ctx, "active")
	if err != nil {
		h.logger.Error("Failed to get active session count", zap.Error(err))
		activeSessions = 0 // 继续执行，但记录错误
	}

	// 获取内存统计
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics := &MetricsResponse{
		Users: UserMetrics{
			Total:  totalUsers,
			Active: 0, // 可以根据需要实现活跃用户统计
		},
		Sessions: SessionMetrics{
			Total:  totalSessions,
			Active: activeSessions,
		},
		System: SystemMetrics{
			Goroutines:  runtime.NumGoroutine(),
			MemoryUsage: memStats.Alloc,
			CPUUsage:    0.0, // 需要实现CPU使用率监控
		},
		Database: DatabaseMetrics{
			Connections: h.getDatabaseConnections(),
		},
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, metrics)
}

// CleanupSystem 系统清理
// @Summary 系统清理
// @Description 执行系统清理操作
// @Tags system
// @Accept json
// @Produce json
// @Param request body CleanupRequest true "清理请求"
// @Success 200 {object} CleanupResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /system/cleanup [post]
func (h *SystemHandler) CleanupSystem(c *gin.Context) {
	var req CleanupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid cleanup request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	response := &CleanupResponse{
		Operations: make(map[string]CleanupResult),
		Timestamp:  time.Now(),
	}

	// 清理过期会话
	if req.CleanExpiredSessions {
		count, err := h.sessionRepo.CleanupExpiredSessions(ctx)
		if err != nil {
			h.logger.Error("Failed to cleanup expired sessions", zap.Error(err))
			response.Operations["expired_sessions"] = CleanupResult{
				Success: false,
				Error:   err.Error(),
			}
		} else {
			response.Operations["expired_sessions"] = CleanupResult{
				Success: true,
				Message: fmt.Sprintf("Cleaned %d expired sessions", count),
			}
		}
	}

	// 清理已撤销会话
	if req.CleanRevokedSessions {
		count, err := h.sessionRepo.CleanupRevokedSessions(ctx)
		if err != nil {
			h.logger.Error("Failed to cleanup revoked sessions", zap.Error(err))
			response.Operations["revoked_sessions"] = CleanupResult{
				Success: false,
				Error:   err.Error(),
			}
		} else {
			response.Operations["revoked_sessions"] = CleanupResult{
				Success: true,
				Message: fmt.Sprintf("Cleaned %d revoked sessions", count),
			}
		}
	}

	// 强制垃圾回收
	if req.ForceGC {
		runtime.GC()
		response.Operations["garbage_collection"] = CleanupResult{
			Success: true,
			Message: "Garbage collection completed",
		}
	}

	c.JSON(http.StatusOK, response)
}

// 辅助方法

func (h *SystemHandler) checkDatabase(ctx context.Context) error {
	sqlDB, err := h.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (h *SystemHandler) checkUserRepository(ctx context.Context) error {
	_, err := h.userRepo.Count(ctx)
	return err
}

func (h *SystemHandler) checkSessionRepository(ctx context.Context) error {
	_, err := h.sessionRepo.Count(ctx)
	return err
}

func (h *SystemHandler) getDatabaseConnections() int {
	sqlDB, err := h.db.DB()
	if err != nil {
		return 0
	}
	stats := sqlDB.Stats()
	return stats.OpenConnections
}

// 全局变量记录启动时间
var startTime = time.Now()

// 响应结构体

type HealthResponse struct {
	Status    string                    `json:"status"`
	Services  map[string]ServiceHealth  `json:"services"`
	Timestamp time.Time                 `json:"timestamp"`
}

type ServiceHealth struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type SystemInfoResponse struct {
	Version   string      `json:"version"`
	BuildTime string      `json:"build_time"`
	GoVersion string      `json:"go_version"`
	Runtime   RuntimeInfo `json:"runtime"`
	Uptime    time.Duration `json:"uptime"`
	Timestamp time.Time   `json:"timestamp"`
}

type RuntimeInfo struct {
	Goroutines  int       `json:"goroutines"`
	CPUs        int       `json:"cpus"`
	MemoryAlloc uint64    `json:"memory_alloc"`
	MemoryTotal uint64    `json:"memory_total"`
	MemorySys   uint64    `json:"memory_sys"`
	GCRuns      uint32    `json:"gc_runs"`
	LastGCTime  time.Time `json:"last_gc_time"`
}

type MetricsResponse struct {
	Users     UserMetrics     `json:"users"`
	Sessions  SessionMetrics  `json:"sessions"`
	System    SystemMetrics   `json:"system"`
	Database  DatabaseMetrics `json:"database"`
	Timestamp time.Time       `json:"timestamp"`
}

type UserMetrics struct {
	Total  int64 `json:"total"`
	Active int64 `json:"active"`
}

type SessionMetrics struct {
	Total  int64 `json:"total"`
	Active int64 `json:"active"`
}

type SystemMetrics struct {
	Goroutines  int     `json:"goroutines"`
	MemoryUsage uint64  `json:"memory_usage"`
	CPUUsage    float64 `json:"cpu_usage"`
}

type DatabaseMetrics struct {
	Connections int `json:"connections"`
}

type CleanupRequest struct {
	CleanExpiredSessions bool `json:"clean_expired_sessions"`
	CleanRevokedSessions bool `json:"clean_revoked_sessions"`
	ForceGC              bool `json:"force_gc"`
}

type CleanupResponse struct {
	Operations map[string]CleanupResult `json:"operations"`
	Timestamp  time.Time                `json:"timestamp"`
}

type CleanupResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}