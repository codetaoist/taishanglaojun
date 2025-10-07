package gateway

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/api"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/config"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/health"
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
	healthChecker health.HealthChecker
	
	// 中间件
	authMiddleware      *middleware.AuthMiddleware
	rateLimitMiddleware *middleware.RateLimitMiddleware
	corsMiddleware      *middleware.CORSMiddleware
	
	// 路由器
	router *gin.Engine
	
	// 动态路由管理器
	dynamicRouter *DynamicRouter
	routeManager  *RouteManager
	
	// 路由配置 (保留用于兼容性)
	routes     map[string]*RouteConfig
	routesMu   sync.RWMutex
	
	// 停止信号
	stopCh chan struct{}
	wg     sync.WaitGroup
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
	healthChecker health.HealthChecker,
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
		healthChecker:       healthChecker,
		authMiddleware:      authMiddleware,
		rateLimitMiddleware: rateLimitMiddleware,
		corsMiddleware:      corsMiddleware,
		routes:              make(map[string]*RouteConfig),
	}
	
	// 创建路由管理器
	gateway.routeManager = NewRouteManager(log)
	
	// 创建动态路由管理器
	gateway.dynamicRouter = NewDynamicRouter(gateway, gateway.routeManager, log)
	
	// 初始化路由器
	if err := gateway.initRouter(); err != nil {
		return nil, fmt.Errorf("failed to initialize router: %w", err)
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
	g.router.Use(g.corsMiddleware.Handler())
	
	// 健康检查端点
	g.router.GET("/health", g.healthHandler)
	g.router.GET("/ready", g.readyHandler)
	
	// 注册负载均衡管理API
	if lbManager, ok := g.proxyManager.(interface{ GetLoadBalancerManager() proxy.LoadBalancerManager }); ok {
		lbAPI := api.NewLoadBalancerAPI(lbManager.GetLoadBalancerManager(), g.healthChecker, g.logger.GetLogrusLogger())
		lbAPI.RegisterRoutes(g.router)
		g.logger.Info("Load balancer API routes registered")
	} else {
		g.logger.Warn("ProxyManager does not implement GetLoadBalancerManager interface")
	}
	
	// 管理端点
	admin := g.router.Group("/gateway/admin")
	{
		admin.GET("/routes", g.listRoutesHandler)
		admin.POST("/routes", g.addRouteHandler)
		admin.PUT("/routes/:method/:path", g.updateRouteHandler)
		admin.DELETE("/routes/:method/:path", g.deleteRouteHandler)
		admin.GET("/services", g.listServicesHandler)
		admin.GET("/metrics", g.metricsHandler)
	}
	
	return nil
}

// Start 启动网关
func (g *Gateway) Start(ctx context.Context) error {
	// 启动健康检查
	if g.healthChecker != nil {
		g.healthChecker.StartHealthChecks(ctx)
	}
	
	// 加载路由配置
	if err := g.LoadRoutes(); err != nil {
		return fmt.Errorf("failed to load routes: %w", err)
	}
	
	g.logger.Info("Gateway started successfully")
	return nil
}

// LoadRoutes 加载路由配置
func (g *Gateway) LoadRoutes() error {
	routes := make([]*RouteConfig, 0)
	
	for _, serviceConfig := range g.config.Services {
		for _, route := range serviceConfig.Routes {
			routeConfig := &RouteConfig{
				ID:          fmt.Sprintf("%s:%s", route.Method, route.Path),
				Path:        route.Path,
				Method:      route.Method,
				Service:     serviceConfig.Name,
				Rewrite:     route.Rewrite,
				StripPrefix: route.StripPrefix,
				Timeout:     time.Duration(route.Timeout) * time.Second,
				Middleware:  route.Middleware,
				Headers:     route.Headers,
				Enabled:     true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			
			routes = append(routes, routeConfig)
		}
	}
	
	// 加载路由到RouteManager
	if err := g.routeManager.LoadRoutes(routes); err != nil {
		return fmt.Errorf("failed to load routes to manager: %w", err)
	}
	
	// 同步到旧的routes map以保持兼容性
	g.routesMu.Lock()
	g.routes = make(map[string]*RouteConfig)
	for _, route := range routes {
		routeKey := fmt.Sprintf("%s:%s", route.Method, route.Path)
		g.routes[routeKey] = route
	}
	g.routesMu.Unlock()
	
	// 注册路由到Gin引擎
	for _, route := range routes {
		if err := g.registerRoute(route); err != nil {
			return fmt.Errorf("failed to register route %s %s: %w", route.Method, route.Path, err)
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
	handlers := g.createRouteHandlers(route)
	
	// 注册路由
	switch strings.ToUpper(route.Method) {
	case "GET":
		g.router.GET(route.Path, handlers...)
	case "POST":
		g.router.POST(route.Path, handlers...)
	case "PUT":
		g.router.PUT(route.Path, handlers...)
	case "DELETE":
		g.router.DELETE(route.Path, handlers...)
	case "PATCH":
		g.router.PATCH(route.Path, handlers...)
	case "HEAD":
		g.router.HEAD(route.Path, handlers...)
	case "OPTIONS":
		g.router.OPTIONS(route.Path, handlers...)
	case "ANY":
		g.router.Any(route.Path, handlers...)
	default:
		return fmt.Errorf("unsupported HTTP method: %s", route.Method)
	}
	
	// 保存路由配置
	g.routes[routeKey] = route
	
	g.logger.Infof("Added route: %s %s -> %s", route.Method, route.Path, route.Service)
	return nil
}

// createRouteHandlers 创建路由处理器链
func (g *Gateway) createRouteHandlers(route *RouteConfig) []gin.HandlerFunc {
	var handlers []gin.HandlerFunc
	
	// 添加中间件
	for _, middlewareName := range route.Middleware {
		switch middlewareName {
		case "auth":
			handlers = append(handlers, g.authMiddleware.Handler())
		case "rate_limit":
			handlers = append(handlers, g.rateLimitMiddleware.Handler())
		case "cors":
			handlers = append(handlers, g.corsMiddleware.Handler())
		default:
			g.logger.WithFields(map[string]interface{}{
				"middleware": middlewareName,
			}).Warn("Unknown middleware")
		}
	}
	
	// 添加主处理器
	handlers = append(handlers, g.createRouteHandler(route))
	
	return handlers
}

// createRouteHandler 创建路由处理器
func (g *Gateway) createRouteHandler(route *RouteConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置超时
		if route.Timeout > 0 {
			c.Header("X-Timeout", route.Timeout.String())
		}
		
		// 添加自定义头部
		for key, value := range route.Headers {
			c.Header(key, value)
		}
		
		// 计算目标路径
		targetPath := route.Path
		if route.Rewrite != "" {
			targetPath = route.Rewrite
		}
		
		if route.StripPrefix {
			// 移除路径前缀
			originalPath := c.Request.URL.Path
			if strings.HasPrefix(originalPath, route.Path) {
				targetPath = strings.TrimPrefix(originalPath, route.Path)
				if !strings.HasPrefix(targetPath, "/") {
					targetPath = "/" + targetPath
				}
			}
		}
		
		// 使用ProxyManager处理请求（已集成负载均衡）
		err := g.proxyManager.HandleRequest(c.Writer, c.Request, route.Service)
		if err != nil {
			g.logger.WithFields(map[string]interface{}{
				"service":    route.Service,
				"targetPath": targetPath,
				"error":      err.Error(),
			}).Error("Proxy request failed")
			
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "Bad gateway",
				"message": "Failed to proxy request",
			})
			return
		}
	}
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
	routes := g.routeManager.ListRoutes()
	
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
	
	if err := g.routeManager.AddRoute(&route); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to add route",
			"message": err.Error(),
		})
		return
	}
	
	// 重新加载路由到Gin引擎
	if err := g.reloadRoutes(); err != nil {
		g.logger.WithFields(map[string]interface{}{
			"error": err,
		}).Error("Failed to reload routes after adding")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Route added but failed to reload",
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
	method := c.Param("method")
	path := c.Param("path")
	
	if method == "" || path == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Method and path are required",
		})
		return
	}
	
	var updatedRoute RouteConfig
	if err := c.ShouldBindJSON(&updatedRoute); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}
	
	if err := g.routeManager.UpdateRoute(method, path, &updatedRoute); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to update route",
			"message": err.Error(),
		})
		return
	}
	
	// 重新加载路由到Gin引擎
	if err := g.reloadRoutes(); err != nil {
		g.logger.WithFields(map[string]interface{}{
			"error": err,
		}).Error("Failed to reload routes after updating")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Route updated but failed to reload",
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Route updated successfully",
		"route":   updatedRoute,
	})
}

// deleteRouteHandler 删除路由处理器
func (g *Gateway) deleteRouteHandler(c *gin.Context) {
	method := c.Param("method")
	path := c.Param("path")
	
	if method == "" || path == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Method and path are required",
		})
		return
	}
	
	deletedRoute, err := g.routeManager.DeleteRoute(method, path)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Route not found",
			"message": err.Error(),
		})
		return
	}
	
	// 重新加载路由到Gin引擎
	if err := g.reloadRoutes(); err != nil {
		g.logger.WithFields(map[string]interface{}{
			"error": err,
		}).Error("Failed to reload routes after deleting")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Route deleted but failed to reload",
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Route deleted successfully",
		"route":   deletedRoute,
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
// Stop 停止网关
func (g *Gateway) Stop() error {
	g.logger.Info("Stopping API Gateway...")
	
	// 停止动态路由管理器
	if err := g.dynamicRouter.Stop(); err != nil {
		g.logger.Errorf("Failed to stop dynamic router: %v", err)
	}
	
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
	if err := g.LoadRoutes(); err != nil {
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