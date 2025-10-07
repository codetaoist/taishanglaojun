package monitoring

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Metrics Prometheus指标收集器
type Metrics struct {
	// HTTP请求指标
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestSize      *prometheus.HistogramVec
	httpResponseSize     *prometheus.HistogramVec

	// 认证指标
	authAttemptsTotal    *prometheus.CounterVec
	authSuccessTotal     *prometheus.CounterVec
	authFailuresTotal    *prometheus.CounterVec

	// 用户指标
	usersTotal           prometheus.Gauge
	activeUsersTotal     prometheus.Gauge
	userRegistrations    *prometheus.CounterVec

	// 会话指标
	sessionsTotal        prometheus.Gauge
	activeSessionsTotal  prometheus.Gauge
	sessionCreations     *prometheus.CounterVec
	sessionDuration      *prometheus.HistogramVec

	// 系统指标
	goroutinesTotal      prometheus.Gauge
	memoryUsage          prometheus.Gauge
	databaseConnections  prometheus.Gauge
	gcDuration           *prometheus.HistogramVec

	// 权限指标
	permissionChecks     *prometheus.CounterVec
	permissionDenials    *prometheus.CounterVec

	logger *zap.Logger
}

// NewMetrics 创建新的指标收集器
func NewMetrics(logger *zap.Logger) *Metrics {
	m := &Metrics{
		logger: logger,
	}

	// HTTP请求指标
	m.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	m.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	m.httpRequestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "path"},
	)

	m.httpResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "path"},
	)

	// 认证指标
	m.authAttemptsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"method", "result"},
	)

	m.authSuccessTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_success_total",
			Help: "Total number of successful authentications",
		},
		[]string{"method"},
	)

	m.authFailuresTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_failures_total",
			Help: "Total number of failed authentications",
		},
		[]string{"method", "reason"},
	)

	// 用户指标
	m.usersTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_users_total",
			Help: "Total number of users",
		},
	)

	m.activeUsersTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_active_users_total",
			Help: "Total number of active users",
		},
	)

	m.userRegistrations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_user_registrations_total",
			Help: "Total number of user registrations",
		},
		[]string{"role"},
	)

	// 会话指标
	m.sessionsTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_sessions_total",
			Help: "Total number of sessions",
		},
	)

	m.activeSessionsTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_active_sessions_total",
			Help: "Total number of active sessions",
		},
	)

	m.sessionCreations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_session_creations_total",
			Help: "Total number of session creations",
		},
		[]string{"user_role"},
	)

	m.sessionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_session_duration_seconds",
			Help:    "Session duration in seconds",
			Buckets: []float64{60, 300, 900, 1800, 3600, 7200, 14400, 28800, 86400},
		},
		[]string{"user_role"},
	)

	// 系统指标
	m.goroutinesTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_goroutines_total",
			Help: "Total number of goroutines",
		},
	)

	m.memoryUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_memory_usage_bytes",
			Help: "Memory usage in bytes",
		},
	)

	m.databaseConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_database_connections_total",
			Help: "Total number of database connections",
		},
	)

	m.gcDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_gc_duration_seconds",
			Help:    "Garbage collection duration in seconds",
			Buckets: []float64{0.001, 0.01, 0.1, 1, 10},
		},
		[]string{"gc_type"},
	)

	// 权限指标
	m.permissionChecks = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_permission_checks_total",
			Help: "Total number of permission checks",
		},
		[]string{"resource", "action", "result"},
	)

	m.permissionDenials = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_permission_denials_total",
			Help: "Total number of permission denials",
		},
		[]string{"resource", "action", "reason"},
	)

	// 注册所有指标
	m.registerMetrics()

	return m
}

// registerMetrics 注册所有指标到Prometheus
func (m *Metrics) registerMetrics() {
	prometheus.MustRegister(
		// HTTP指标
		m.httpRequestsTotal,
		m.httpRequestDuration,
		m.httpRequestSize,
		m.httpResponseSize,

		// 认证指标
		m.authAttemptsTotal,
		m.authSuccessTotal,
		m.authFailuresTotal,

		// 用户指标
		m.usersTotal,
		m.activeUsersTotal,
		m.userRegistrations,

		// 会话指标
		m.sessionsTotal,
		m.activeSessionsTotal,
		m.sessionCreations,
		m.sessionDuration,

		// 系统指标
		m.goroutinesTotal,
		m.memoryUsage,
		m.databaseConnections,
		m.gcDuration,

		// 权限指标
		m.permissionChecks,
		m.permissionDenials,
	)
}

// RecordHTTPRequest 记录HTTP请求指标
func (m *Metrics) RecordHTTPRequest(method, path, status string, duration time.Duration, requestSize, responseSize int64) {
	m.httpRequestsTotal.WithLabelValues(method, path, status).Inc()
	m.httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
	
	if requestSize > 0 {
		m.httpRequestSize.WithLabelValues(method, path).Observe(float64(requestSize))
	}
	if responseSize > 0 {
		m.httpResponseSize.WithLabelValues(method, path).Observe(float64(responseSize))
	}
}

// RecordAuthAttempt 记录认证尝试
func (m *Metrics) RecordAuthAttempt(method, result string) {
	m.authAttemptsTotal.WithLabelValues(method, result).Inc()
}

// RecordAuthSuccess 记录认证成功
func (m *Metrics) RecordAuthSuccess(method string) {
	m.authSuccessTotal.WithLabelValues(method).Inc()
}

// RecordAuthFailure 记录认证失败
func (m *Metrics) RecordAuthFailure(method, reason string) {
	m.authFailuresTotal.WithLabelValues(method, reason).Inc()
}

// UpdateUserMetrics 更新用户指标
func (m *Metrics) UpdateUserMetrics(total, active int64) {
	m.usersTotal.Set(float64(total))
	m.activeUsersTotal.Set(float64(active))
}

// RecordUserRegistration 记录用户注册
func (m *Metrics) RecordUserRegistration(role string) {
	m.userRegistrations.WithLabelValues(role).Inc()
}

// UpdateSessionMetrics 更新会话指标
func (m *Metrics) UpdateSessionMetrics(total, active int64) {
	m.sessionsTotal.Set(float64(total))
	m.activeSessionsTotal.Set(float64(active))
}

// RecordSessionCreation 记录会话创建
func (m *Metrics) RecordSessionCreation(userRole string) {
	m.sessionCreations.WithLabelValues(userRole).Inc()
}

// RecordSessionDuration 记录会话持续时间
func (m *Metrics) RecordSessionDuration(userRole string, duration time.Duration) {
	m.sessionDuration.WithLabelValues(userRole).Observe(duration.Seconds())
}

// UpdateSystemMetrics 更新系统指标
func (m *Metrics) UpdateSystemMetrics(goroutines int, memoryUsage uint64, dbConnections int) {
	m.goroutinesTotal.Set(float64(goroutines))
	m.memoryUsage.Set(float64(memoryUsage))
	m.databaseConnections.Set(float64(dbConnections))
}

// RecordGCDuration 记录垃圾回收时间
func (m *Metrics) RecordGCDuration(gcType string, duration time.Duration) {
	m.gcDuration.WithLabelValues(gcType).Observe(duration.Seconds())
}

// RecordPermissionCheck 记录权限检查
func (m *Metrics) RecordPermissionCheck(resource, action, result string) {
	m.permissionChecks.WithLabelValues(resource, action, result).Inc()
}

// RecordPermissionDenial 记录权限拒绝
func (m *Metrics) RecordPermissionDenial(resource, action, reason string) {
	m.permissionDenials.WithLabelValues(resource, action, reason).Inc()
}

// RecordSystemCleanup 记录系统清理操作
func (m *Metrics) RecordSystemCleanup(resource, operation, result string) {
	// 使用现有的指标或创建一个通用的系统操作计数器
	// 这里我们可以复用权限检查的指标结构，或者记录到日志
	m.logger.Info("System cleanup operation",
		zap.String("resource", resource),
		zap.String("operation", operation),
		zap.String("result", result),
	)
	
	// 如果需要专门的清理指标，可以在 NewMetrics 中添加相应的 prometheus 指标
	// 这里暂时使用日志记录
}

// GinMiddleware 返回Gin中间件用于自动收集HTTP指标
func (m *Metrics) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 获取请求大小
		requestSize := c.Request.ContentLength
		if requestSize < 0 {
			requestSize = 0
		}

		// 处理请求
		c.Next()

		// 计算响应时间
		duration := time.Since(start)
		
		// 获取响应大小
		responseSize := int64(c.Writer.Size())
		if responseSize < 0 {
			responseSize = 0
		}

		// 记录指标
		m.RecordHTTPRequest(
			c.Request.Method,
			c.FullPath(),
			strconv.Itoa(c.Writer.Status()),
			duration,
			requestSize,
			responseSize,
		)
	}
}

// Handler 返回Prometheus指标处理器
func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}