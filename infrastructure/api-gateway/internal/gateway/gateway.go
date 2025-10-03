package gateway

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/config"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/logger"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/middleware"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/monitoring"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/proxy"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/registry"
)

// Gateway API网关
type Gateway struct {
	config       *config.Config
	logger       logger.Logger
	metrics      monitoring.Metrics
	registry     registry.Registry
	proxyManager proxy.ProxyManager
	
	// 中间件
	authMiddleware      *middleware.AuthMiddleware
	rateLimitMiddleware *middleware.RateLimitMiddleware
	corsMiddleware      *middleware.CORSMiddleware
	
	// 路由器
	router *gin.Engine
	
	// 路由配置
	routes     map[string]*RouteConfig
	routesMu   sync.RWMutex
	
	// 停止信号
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// RouteConfig 路由配置
type RouteConfig struct {
	Path        string            `yaml:"path"`
	Method      string            `yaml:"method"`
	Service     string            `yaml:"service"`
	Rewrite     string            `yaml:"rewrite"`
	StripPrefix bool              `yaml:"strip_prefix"`
	Timeout     time.Duration     `yaml:"timeout"`
	Middleware  []string          `yaml:"middleware"`
	Headers     map[string]string `yaml:"headers"`
	
	// 负载均衡配置
	LoadBalancer string `yaml:"load_balancer"`
	
	// 熔断配置
	CircuitBreaker *CircuitBreakerConfig `yaml:"circuit_breaker"`
	
	// 重试配置
	Retry *RetryConfig `yaml:"retry"`
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	Enabled           bool          `yaml:"enabled"`
	RequestThreshold  int           `yaml:"request_threshold"`
	ErrorThreshold    float64       `yaml:"error_threshold"`
	Timeout           time.Duration `yaml:"timeout"`
	MaxConcurrent     int           `yaml:"max_concurrent"`
	SleepWindow       time.Duration `yaml:"sleep_window"`
}

// RetryConfig 重试配置
type RetryConfig struct {
	Enabled     bool          `yaml:"enabled"`
	MaxAttempts int           `yaml:"max_attempts"`
	Delay       time.Duration `yaml:"delay"`
	BackoffRate float64       `yaml:"backoff_rate"`
}

// NewGateway 创建API网关
func NewGateway(
	cfg *config.Config,
	log logger.Logger,
	metrics monitoring.Metrics,
	reg registry.Registry,
	pm proxy.ProxyManager,
) (*Gateway, error) {
	
	// 创建中间件
	authConfig := &middleware.AuthConfig{
		JWTSecret:     cfg.Security.Auth.JWTSecret,
		TokenExpiry:   cfg.Security.Auth.TokenExpiry,
		RefreshExpiry: cfg.Security.Auth.RefreshExpiry,
		RedisAddr:     cfg.Security.Auth.RedisAddr,
		RedisPassword: cfg.Security.Auth.RedisPassword,
		RedisDB:       cfg.Security.Auth.RedisDB,
		SkipPaths:     cfg.Security.Auth.SkipPaths,
		OptionalPaths: cfg.Security.Auth.OptionalPaths,
	}
	
	// 如果嵌套配置为空，使用向后兼容的配置
	if authConfig.JWTSecret == "" {
		authConfig.JWTSecret = cfg.Security.JWTSecret
	}
	
	authMiddleware, err := middleware.NewAuthMiddleware(authConfig, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth middleware: %w", err)
	}
	
	rateLimitMiddleware, err := middleware.NewRateLimitMiddleware(&middleware.RateLimitConfig{
		GlobalRPS:   cfg.RateLimit.DefaultRate,
		GlobalBurst: cfg.RateLimit.DefaultBurst,
		IPRPS:       cfg.RateLimit.DefaultRate,
		IPBurst:     cfg.RateLimit.DefaultBurst,
	}, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create rate limit middleware: %w", err)
	}
	
	corsMiddleware := middleware.NewCORSMiddleware(&middleware.CORSConfig{
		AllowOrigins: cfg.Security.CORSOrigins,
		AllowMethods: cfg.Security.CORSMethods,
		AllowHeaders: cfg.Security.CORSHeaders,
	})
	
	gateway := &Gateway{
		config:              cfg,
		logger:              log,
		metrics:             metrics,
		registry:            reg,
		proxyManager:        pm,
		authMiddleware:      authMiddleware,
		rateLimitMiddleware: rateLimitMiddleware,
		corsMiddleware:      corsMiddleware,
		routes:              make(map[string]*RouteConfig),
		stopCh:              make(chan struct{}),
	}
	
	// 初始化路由器
	if err := gateway.initRouter(); err != nil {
		return nil, fmt.Errorf("failed to initialize router: %w", err)
	}
	
	// 加载路由配置
	if err := gateway.loadRoutes(); err != nil {
		return nil, fmt.Errorf("failed to load routes: %w", err)
	}
	
	return gateway, nil
}

// initRouter 初始化路由器
func (g *Gateway) initRouter() error {
	// 设置Gin模式
	if g.config.Server.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	
	g.router = gin.New()
	
	// 添加全局中间件
	g.router.Use(gin.Recovery())
	g.router.Use(g.loggingMiddleware())
	g.router.Use(g.metricsMiddleware())
	g.router.Use(g.corsMiddleware.Handler())
	
	// 健康检查端点
	g.router.GET("/health", g.healthHandler)
	g.router.GET("/ready", g.readyHandler)
	
	// 管理端点
	admin := g.router.Group("/gateway/admin")
	{
		admin.GET("/routes", g.listRoutesHandler)
		admin.POST("/routes", g.addRouteHandler)
		admin.PUT("/routes/:id", g.updateRouteHandler)
		admin.DELETE("/routes/:id", g.deleteRouteHandler)
		admin.GET("/services", g.listServicesHandler)
		admin.GET("/metrics", g.metricsHandler)
	}
	
	return nil
}

// loadRoutes 加载路由配置
func (g *Gateway) loadRoutes() error {
	for _, serviceConfig := range g.config.Services {
		for _, route := range serviceConfig.Routes {
			routeConfig := &RouteConfig{
				Path:           route.Path,
				Method:         route.Method,
				Service:        serviceConfig.Name,
				Rewrite:        route.Rewrite,
				StripPrefix:    route.StripPrefix,
				Timeout:        time.Duration(route.Timeout) * time.Second,
				Middleware:     route.Middleware,
				Headers:        route.Headers,
				LoadBalancer:   serviceConfig.LoadBalancer,
				CircuitBreaker: nil,
				Retry:          nil,
			}
			
			if err := g.addRoute(routeConfig); err != nil {
				return fmt.Errorf("failed to add route %s: %w", route.Path, err)
			}
		}
	}
	
	return nil
}

// addRoute 添加路由
func (g *Gateway) addRoute(route *RouteConfig) error {
	g.routesMu.Lock()
	defer g.routesMu.Unlock()
	
	// 生成路由键
	routeKey := fmt.Sprintf("%s:%s", route.Method, route.Path)
	
	// 检查路由是否已存在
	if _, exists := g.routes[routeKey]; exists {
		g.logger.Warnf("Route already exists, skipping: %s %s", route.Method, route.Path)
		return nil
	}
	
	// 创建处理器
	handler := g.createRouteHandler(route)
	
	// 注册路由
	switch strings.ToUpper(route.Method) {
	case "GET":
		g.router.GET(route.Path, handler...)
	case "POST":
		g.router.POST(route.Path, handler...)
	case "PUT":
		g.router.PUT(route.Path, handler...)
	case "DELETE":
		g.router.DELETE(route.Path, handler...)
	case "PATCH":
		g.router.PATCH(route.Path, handler...)
	case "HEAD":
		g.router.HEAD(route.Path, handler...)
	case "OPTIONS":
		g.router.OPTIONS(route.Path, handler...)
	case "ANY":
		g.router.Any(route.Path, handler...)
	default:
		return fmt.Errorf("unsupported HTTP method: %s", route.Method)
	}
	
	// 保存路由配置
	g.routes[routeKey] = route
	
	g.logger.Infof("Added route: %s %s -> %s", route.Method, route.Path, route.Service)
	return nil
}

// createRouteHandler 创建路由处理器
func (g *Gateway) createRouteHandler(route *RouteConfig) []gin.HandlerFunc {
	var handlers []gin.HandlerFunc
	
	// 调试日志：确认中间件配置
	g.logger.Infof("Creating route handler for %s %s with middleware: %v", route.Method, route.Path, route.Middleware)
	
	// 添加中间件
	for _, middlewareName := range route.Middleware {
		switch middlewareName {
		case "auth":
			g.logger.Infof("Adding auth middleware for %s %s", route.Method, route.Path)
			handlers = append(handlers, g.authMiddleware.Handler())
		case "rate_limit":
			handlers = append(handlers, g.rateLimitMiddleware.Handler())
		case "admin":
			handlers = append(handlers, middleware.RequireRole("admin"))
		}
	}
	
	// 添加代理处理器
	handlers = append(handlers, g.proxyHandler(route))
	
	return handlers
}

// proxyHandler 代理处理器
func (g *Gateway) proxyHandler(route *RouteConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 设置超时
		ctx := c.Request.Context()
		if route.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, route.Timeout)
			defer cancel()
		}
		
		// 添加自定义头部
		for key, value := range route.Headers {
			c.Request.Header.Set(key, value)
		}
		
		// 路径重写
		originalPath := c.Request.URL.Path
		if route.Rewrite != "" {
			// 处理路径参数重写
			rewrittenPath := route.Rewrite
			
			// 如果路由路径包含参数（如 :id），需要将参数值传递到重写路径
			if strings.Contains(route.Path, ":") && strings.Contains(route.Rewrite, ":") {
				// 提取路径参数
				pathParts := strings.Split(strings.Trim(route.Path, "/"), "/")
				originalParts := strings.Split(strings.Trim(originalPath, "/"), "/")
				rewriteParts := strings.Split(strings.Trim(route.Rewrite, "/"), "/")
				
				// 替换重写路径中的参数
				for i, part := range pathParts {
					if strings.HasPrefix(part, ":") && i < len(originalParts) && i < len(rewriteParts) {
						paramName := part[1:] // 移除 : 前缀
						for j, rewritePart := range rewriteParts {
							if rewritePart == ":"+paramName {
								rewriteParts[j] = originalParts[i]
							}
						}
					}
				}
				rewrittenPath = "/" + strings.Join(rewriteParts, "/")
			}
			
			c.Request.URL.Path = rewrittenPath
		} else if route.StripPrefix {
			// 移除路由前缀
			if strings.HasPrefix(originalPath, route.Path) {
				c.Request.URL.Path = strings.TrimPrefix(originalPath, strings.TrimSuffix(route.Path, "*"))
			}
		}
		
		// 获取服务实例
		instances, err := g.registry.Discover(c.Request.Context(), route.Service)
		if err != nil {
			g.logger.WithFields(map[string]interface{}{
				"service": route.Service,
				"error":   err.Error(),
			}).Error("Failed to get service instances")
			
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service unavailable",
				"message": "Failed to get service instances",
			})
			return
		}
		
		if len(instances) == 0 {
			g.logger.WithFields(map[string]interface{}{
				"service": route.Service,
			}).Error("No available service instances")
			
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service unavailable",
				"message": "No available service instances",
			})
			return
		}
		
		// 使用ProxyManager处理请求
		err = g.proxyManager.HandleRequest(c, route.Service)
		if err != nil {
			g.logger.WithFields(map[string]interface{}{
				"service": route.Service,
				"error":   err.Error(),
			}).Error("Failed to proxy request")
			
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "Bad gateway",
				"message": "Failed to proxy request",
			})
			return
		}
		
		// 记录指标
		duration := time.Since(start)
		status := fmt.Sprintf("%d", c.Writer.Status())
		g.metrics.RecordProxyRequest(route.Service, c.Request.Method, status, duration)
		
		g.logger.WithFields(map[string]interface{}{
			"service":  route.Service,
			"path":     route.Path,
			"method":   c.Request.Method,
			"status":   status,
			"duration": duration.String(),
		}).Info("Request proxied")
	}
}

// loggingMiddleware 日志中间件
func (g *Gateway) loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		g.logger.WithFields(map[string]interface{}{
			"timestamp":   param.TimeStamp.Format(time.RFC3339),
			"status":      param.StatusCode,
			"latency":     param.Latency,
			"client_ip":   param.ClientIP,
			"method":      param.Method,
			"path":        param.Path,
			"user_agent":  param.Request.UserAgent(),
			"error":       param.ErrorMessage,
		}).Info("HTTP request")
		
		return ""
	})
}

// metricsMiddleware 指标中间件
func (g *Gateway) metricsMiddleware() gin.HandlerFunc {
	return g.metrics.GinMiddleware()
}

// healthHandler 健康检查处理器
func (g *Gateway) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	})
}

// readyHandler 就绪检查处理器
func (g *Gateway) readyHandler(c *gin.Context) {
	// 检查依赖服务状态
	ctx := c.Request.Context()
	services := make(map[string]string)
	ready := true
	
	// 检查配置的服务
	for _, serviceConfig := range g.config.Services {
		instances, err := g.registry.Discover(ctx, serviceConfig.Name)
		if err != nil || len(instances) == 0 {
			ready = false
			services[serviceConfig.Name] = "unavailable"
		} else {
			services[serviceConfig.Name] = "available"
		}
	}
	
	status := http.StatusOK
	if !ready {
		status = http.StatusServiceUnavailable
	}
	
	c.JSON(status, gin.H{
		"status":    map[bool]string{true: "ready", false: "not ready"}[ready],
		"services":  services,
		"timestamp": time.Now().Unix(),
	})
}

// listRoutesHandler 列出路由处理器
func (g *Gateway) listRoutesHandler(c *gin.Context) {
	g.routesMu.RLock()
	defer g.routesMu.RUnlock()
	
	routes := make([]*RouteConfig, 0, len(g.routes))
	for _, route := range g.routes {
		routes = append(routes, route)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"routes": routes,
		"count":  len(routes),
	})
}

// addRouteHandler 添加路由处理器
func (g *Gateway) addRouteHandler(c *gin.Context) {
	var route RouteConfig
	if err := c.ShouldBindJSON(&route); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}
	
	if err := g.addRoute(&route); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to add route",
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "Route added successfully",
		"route":   route,
	})
}

// updateRouteHandler 更新路由处理器
func (g *Gateway) updateRouteHandler(c *gin.Context) {
	// 实现路由更新逻辑
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Route update not implemented yet",
	})
}

// deleteRouteHandler 删除路由处理器
func (g *Gateway) deleteRouteHandler(c *gin.Context) {
	// 实现路由删除逻辑
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Route deletion not implemented yet",
	})
}

// listServicesHandler 列出服务处理器
func (g *Gateway) listServicesHandler(c *gin.Context) {
	ctx := c.Request.Context()
	services, err := g.registry.ListServices(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list services",
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"services": services,
	})
}

// metricsHandler 指标处理器
func (g *Gateway) metricsHandler(c *gin.Context) {
	// 返回Prometheus格式的指标
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, "# Metrics endpoint - implement Prometheus handler")
}

// GetRouter 获取路由器
func (g *Gateway) GetRouter() *gin.Engine {
	return g.router
}

// Start 启动网关
func (g *Gateway) Start() error {
	g.logger.Info("Starting API Gateway...")
	return nil
}

// Stop 停止网关
func (g *Gateway) Stop() error {
	g.logger.Info("Stopping API Gateway...")
	
	close(g.stopCh)
	g.wg.Wait()
	
	// 关闭中间件资源
	if err := g.authMiddleware.Close(); err != nil {
		g.logger.Errorf("Failed to close auth middleware: %v", err)
	}
	
	if err := g.rateLimitMiddleware.Close(); err != nil {
		g.logger.Errorf("Failed to close rate limit middleware: %v", err)
	}
	
	g.logger.Info("API Gateway stopped")
	return nil
}

// ReloadRoutes 重新加载路由
func (g *Gateway) ReloadRoutes() error {
	g.logger.Info("Reloading routes...")
	
	// 清空现有路由
	g.routesMu.Lock()
	g.routes = make(map[string]*RouteConfig)
	g.routesMu.Unlock()
	
	// 重新初始化路由器
	if err := g.initRouter(); err != nil {
		return fmt.Errorf("failed to reinitialize router: %w", err)
	}
	
	// 重新加载路由配置
	if err := g.loadRoutes(); err != nil {
		return fmt.Errorf("failed to reload routes: %w", err)
	}
	
	g.logger.Info("Routes reloaded successfully")
	return nil
}

// GetRouteCount 获取路由数量
func (g *Gateway) GetRouteCount() int {
	g.routesMu.RLock()
	defer g.routesMu.RUnlock()
	return len(g.routes)
}

// GetRoute 获取路由配置
func (g *Gateway) GetRoute(method, path string) (*RouteConfig, bool) {
	g.routesMu.RLock()
	defer g.routesMu.RUnlock()
	
	routeKey := fmt.Sprintf("%s:%s", method, path)
	route, exists := g.routes[routeKey]
	return route, exists
}