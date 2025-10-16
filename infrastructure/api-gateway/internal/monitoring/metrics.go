package monitoring

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/config"
)

// Metrics 监控指标接口
type Metrics interface {
	// HTTP指标
	RecordHTTPRequest(method, path, status string, duration time.Duration)
	RecordHTTPRequestSize(method, path string, size int64)
	RecordHTTPResponseSize(method, path string, size int64)
	
	// 代理指标
	RecordProxyRequest(service, method, status string, duration time.Duration)
	RecordProxyError(service, errorType string)
	
	// 服务指标
	RecordServiceHealth(service string, healthy bool)
	RecordServiceLatency(service string, duration time.Duration)
	
	// 限流指标
	RecordRateLimitHit(service, path string)
	RecordRateLimitPass(service, path string)
	
	// 熔断器指标
	RecordCircuitBreakerState(service, state string)
	RecordCircuitBreakerRequest(service, result string)
	
	// 启动监控服务器
	StartServer(port int) error
	
	// 获取Gin中间件
	GinMiddleware() gin.HandlerFunc
}

// prometheusMetrics Prometheus实现
type prometheusMetrics struct {
	// HTTP指标
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestSize      *prometheus.HistogramVec
	httpResponseSize     *prometheus.HistogramVec
	
	// 代理指标
	proxyRequestsTotal   *prometheus.CounterVec
	proxyRequestDuration *prometheus.HistogramVec
	proxyErrorsTotal     *prometheus.CounterVec
	
	// 服务指标
	serviceHealthGauge   *prometheus.GaugeVec
	serviceLatency       *prometheus.HistogramVec
	
	// 限流指标
	rateLimitHitsTotal   *prometheus.CounterVec
	rateLimitPassTotal   *prometheus.CounterVec
	
	// 熔断器指标
	circuitBreakerState  *prometheus.GaugeVec
	circuitBreakerTotal  *prometheus.CounterVec
	
	registry *prometheus.Registry
}

// New 创建新的监控实例
func New(cfg config.MonitoringConfig) Metrics {
	registry := prometheus.NewRegistry()
	
	m := &prometheusMetrics{
		// HTTP指标
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gateway_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		httpRequestSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gateway_http_request_size_bytes",
				Help:    "HTTP request size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "path"},
		),
		httpResponseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gateway_http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "path"},
		),
		
		// 代理指标
		proxyRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_proxy_requests_total",
				Help: "Total number of proxy requests",
			},
			[]string{"service", "method", "status"},
		),
		proxyRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gateway_proxy_request_duration_seconds",
				Help:    "Proxy request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method"},
		),
		proxyErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_proxy_errors_total",
				Help: "Total number of proxy errors",
			},
			[]string{"service", "error_type"},
		),
		
		// 服务指标
		serviceHealthGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gateway_service_health",
				Help: "Service health status (1=healthy, 0=unhealthy)",
			},
			[]string{"service"},
		),
		serviceLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gateway_service_latency_seconds",
				Help:    "Service latency in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service"},
		),
		
		// 限流指标
		rateLimitHitsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_rate_limit_hits_total",
				Help: "Total number of rate limit hits",
			},
			[]string{"service", "path"},
		),
		rateLimitPassTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_rate_limit_pass_total",
				Help: "Total number of rate limit passes",
			},
			[]string{"service", "path"},
		),
		
		// 熔断器指标
		circuitBreakerState: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gateway_circuit_breaker_state",
				Help: "Circuit breaker state (0=closed, 1=open, 2=half-open)",
			},
			[]string{"service"},
		),
		circuitBreakerTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_circuit_breaker_requests_total",
				Help: "Total number of circuit breaker requests",
			},
			[]string{"service", "result"},
		),
		
		registry: registry,
	}
	
	// 注册指标
	registry.MustRegister(
		m.httpRequestsTotal,
		m.httpRequestDuration,
		m.httpRequestSize,
		m.httpResponseSize,
		m.proxyRequestsTotal,
		m.proxyRequestDuration,
		m.proxyErrorsTotal,
		m.serviceHealthGauge,
		m.serviceLatency,
		m.rateLimitHitsTotal,
		m.rateLimitPassTotal,
		m.circuitBreakerState,
		m.circuitBreakerTotal,
	)
	
	return m
}

// RecordHTTPRequest 记录HTTP请求
func (m *prometheusMetrics) RecordHTTPRequest(method, path, status string, duration time.Duration) {
	m.httpRequestsTotal.WithLabelValues(method, path, status).Inc()
	m.httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// RecordHTTPRequestSize 记录HTTP请求大小
func (m *prometheusMetrics) RecordHTTPRequestSize(method, path string, size int64) {
	m.httpRequestSize.WithLabelValues(method, path).Observe(float64(size))
}

// RecordHTTPResponseSize 记录HTTP响应大小
func (m *prometheusMetrics) RecordHTTPResponseSize(method, path string, size int64) {
	m.httpResponseSize.WithLabelValues(method, path).Observe(float64(size))
}

// RecordProxyRequest 记录代理请求
func (m *prometheusMetrics) RecordProxyRequest(service, method, status string, duration time.Duration) {
	m.proxyRequestsTotal.WithLabelValues(service, method, status).Inc()
	m.proxyRequestDuration.WithLabelValues(service, method).Observe(duration.Seconds())
}

// RecordProxyError 记录代理错误
func (m *prometheusMetrics) RecordProxyError(service, errorType string) {
	m.proxyErrorsTotal.WithLabelValues(service, errorType).Inc()
}

// RecordServiceHealth 记录服务健康状态
func (m *prometheusMetrics) RecordServiceHealth(service string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	m.serviceHealthGauge.WithLabelValues(service).Set(value)
}

// RecordServiceLatency 记录服务延迟
func (m *prometheusMetrics) RecordServiceLatency(service string, duration time.Duration) {
	m.serviceLatency.WithLabelValues(service).Observe(duration.Seconds())
}

// RecordRateLimitHit 记录限流命中
func (m *prometheusMetrics) RecordRateLimitHit(service, path string) {
	m.rateLimitHitsTotal.WithLabelValues(service, path).Inc()
}

// RecordRateLimitPass 记录限流通过
func (m *prometheusMetrics) RecordRateLimitPass(service, path string) {
	m.rateLimitPassTotal.WithLabelValues(service, path).Inc()
}

// RecordCircuitBreakerState 记录熔断器状态
func (m *prometheusMetrics) RecordCircuitBreakerState(service, state string) {
	var value float64
	switch state {
	case "closed":
		value = 0
	case "open":
		value = 1
	case "half-open":
		value = 2
	}
	m.circuitBreakerState.WithLabelValues(service).Set(value)
}

// RecordCircuitBreakerRequest 记录熔断器请求
func (m *prometheusMetrics) RecordCircuitBreakerRequest(service, result string) {
	m.circuitBreakerTotal.WithLabelValues(service, result).Inc()
}

// StartServer 启动监控服务器
func (m *prometheusMetrics) StartServer(port int) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	
	return server.ListenAndServe()
}

// GinMiddleware 获取Gin中间件
func (m *prometheusMetrics) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 记录请求大小
		if c.Request.ContentLength > 0 {
			m.RecordHTTPRequestSize(c.Request.Method, c.FullPath(), c.Request.ContentLength)
		}
		
		c.Next()
		
		// 记录请求指标
		duration := time.Since(start)
		status := strconv.Itoa(c.Writer.Status())
		
		m.RecordHTTPRequest(c.Request.Method, c.FullPath(), status, duration)
		
		// 记录响应大小
		if c.Writer.Size() > 0 {
			m.RecordHTTPResponseSize(c.Request.Method, c.FullPath(), int64(c.Writer.Size()))
		}
	}
}

// NewNop 创建空监控实例（用于测试）
func NewNop() Metrics {
	return &nopMetrics{}
}

// nopMetrics 空实现
type nopMetrics struct{}

func (n *nopMetrics) RecordHTTPRequest(method, path, status string, duration time.Duration) {}
func (n *nopMetrics) RecordHTTPRequestSize(method, path string, size int64) {}
func (n *nopMetrics) RecordHTTPResponseSize(method, path string, size int64) {}
func (n *nopMetrics) RecordProxyRequest(service, method, status string, duration time.Duration) {}
func (n *nopMetrics) RecordProxyError(service, errorType string) {}
func (n *nopMetrics) RecordServiceHealth(service string, healthy bool) {}
func (n *nopMetrics) RecordServiceLatency(service string, duration time.Duration) {}
func (n *nopMetrics) RecordRateLimitHit(service, path string) {}
func (n *nopMetrics) RecordRateLimitPass(service, path string) {}
func (n *nopMetrics) RecordCircuitBreakerState(service, state string) {}
func (n *nopMetrics) RecordCircuitBreakerRequest(service, result string) {}
func (n *nopMetrics) StartServer(port int) error { return nil }
func (n *nopMetrics) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}