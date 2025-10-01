package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/taishang/sequence-zero/pkg/security"
)

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name string
	URL  string
	Path string
}

// Gateway API网关结构体
type Gateway struct {
	services map[string]*ServiceConfig
	proxies  map[string]*httputil.ReverseProxy
}

// NewGateway 创建新的API网关
func NewGateway() *Gateway {
	gateway := &Gateway{
		services: make(map[string]*ServiceConfig),
		proxies:  make(map[string]*httputil.ReverseProxy),
	}

	// 注册服务
	gateway.registerServices()
	return gateway
}

// registerServices 注册微服务
func (g *Gateway) registerServices() {
	services := []ServiceConfig{
		{
			Name: "auth-service",
			URL:  getEnvOrDefault("AUTH_SERVICE_URL", "http://localhost:8081"),
			Path: "/api/v1/auth",
		},
		{
			Name: "consciousness-service",
			URL:  getEnvOrDefault("CONSCIOUSNESS_SERVICE_URL", "http://localhost:8082"),
			Path: "/api/v1/consciousness",
		},
		{
			Name: "cultural-service",
			URL:  getEnvOrDefault("CULTURAL_SERVICE_URL", "http://localhost:8083"),
			Path: "/api/v1/cultural",
		},
		{
			Name: "permission-service",
			URL:  getEnvOrDefault("PERMISSION_SERVICE_URL", "http://localhost:8084"),
			Path: "/api/v1/permission",
		},
	}

	for _, service := range services {
		g.registerService(service)
	}
}

// registerService 注册单个服务
func (g *Gateway) registerService(config ServiceConfig) {
	targetURL, err := url.Parse(config.URL)
	if err != nil {
		log.Printf("Failed to parse URL for service %s: %v", config.Name, err)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	
	// 自定义错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error for service %s: %v", config.Name, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(fmt.Sprintf(`{"error": "Service %s unavailable", "message": "%s"}`, config.Name, err.Error())))
	}

	// 修改请求
	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Set("X-Gateway", "taishang-sequence-zero")
		resp.Header.Set("X-Service", config.Name)
		return nil
	}

	g.services[config.Path] = &config
	g.proxies[config.Path] = proxy

	log.Printf("Registered service: %s -> %s", config.Path, config.URL)
}

// setupRoutes 设置路由
func (g *Gateway) setupRoutes() *gin.Engine {
	r := gin.Default()

	// 添加CORS中间件
	r.Use(corsMiddleware())

	// 添加请求日志中间件
	r.Use(requestLogMiddleware())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "healthy",
			"service":  "taishang-api-gateway",
			"version":  "1.0.0",
			"services": g.getServiceStatus(),
		})
	})

	// 服务发现端点
	r.GET("/services", func(c *gin.Context) {
		serviceList := make([]map[string]string, 0)
		for path, config := range g.services {
			serviceList = append(serviceList, map[string]string{
				"name": config.Name,
				"path": path,
				"url":  config.URL,
			})
		}
		c.JSON(http.StatusOK, gin.H{
			"services": serviceList,
		})
	})

	// 代理所有API请求
	r.Any("/api/*path", g.proxyHandler)

	return r
}

// proxyHandler 代理处理器
func (g *Gateway) proxyHandler(c *gin.Context) {
	path := c.Param("path")
	fullPath := "/api" + path

	// 查找匹配的服务
	var matchedService *ServiceConfig
	var matchedProxy *httputil.ReverseProxy
	var matchedPath string

	for servicePath, service := range g.services {
		if strings.HasPrefix(fullPath, servicePath) {
			matchedService = service
			matchedProxy = g.proxies[servicePath]
			matchedPath = servicePath
			break
		}
	}

	if matchedService == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Service not found",
			"path":  fullPath,
		})
		return
	}

	// 检查是否需要认证
	if g.requiresAuth(fullPath) {
		if err := g.validateAuth(c); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"message": err.Error(),
			})
			return
		}
	}

	// 修改请求路径
	c.Request.URL.Path = strings.TrimPrefix(fullPath, matchedPath)
	if c.Request.URL.Path == "" {
		c.Request.URL.Path = "/"
	}

	// 添加请求头
	c.Request.Header.Set("X-Gateway-Service", matchedService.Name)
	c.Request.Header.Set("X-Gateway-Path", matchedPath)
	c.Request.Header.Set("X-Original-Path", fullPath)

	// 代理请求
	matchedProxy.ServeHTTP(c.Writer, c.Request)
}

// requiresAuth 检查是否需要认证
func (g *Gateway) requiresAuth(path string) bool {
	// 公开端点，不需要认证
	publicPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/cultural/wisdom/daily",
		"/api/v1/cultural/knowledge/search",
	}

	for _, publicPath := range publicPaths {
		if strings.HasPrefix(path, publicPath) {
			return false
		}
	}

	return true
}

// validateAuth 验证认证
func (g *Gateway) validateAuth(c *gin.Context) error {
	token := c.GetHeader("Authorization")
	if token == "" {
		return fmt.Errorf("missing authorization header")
	}

	// 移除 "Bearer " 前缀
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	// 验证JWT令牌
	if err := security.ValidateJWT(token); err != nil {
		return fmt.Errorf("invalid token: %v", err)
	}

	return nil
}

// getServiceStatus 获取服务状态
func (g *Gateway) getServiceStatus() map[string]string {
	status := make(map[string]string)
	for _, service := range g.services {
		// 简单的健康检查
		if g.checkServiceHealth(service.URL) {
			status[service.Name] = "healthy"
		} else {
			status[service.Name] = "unhealthy"
		}
	}
	return status
}

// checkServiceHealth 检查服务健康状态
func (g *Gateway) checkServiceHealth(serviceURL string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(serviceURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-Gateway, X-Service")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// requestLogMiddleware 请求日志中间件
func requestLogMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\
",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// getEnvOrDefault 获取环境变量或默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// 创建API网关
	gateway := NewGateway()

	// 设置路由
	r := gateway.setupRoutes()

	// 启动服务器
	port := getEnvOrDefault("GATEWAY_PORT", "8080")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	fmt.Printf("太上序列零 - API网关启动在端口 %s\n", port)
	fmt.Println("注册的服务:")
	for path, service := range gateway.services {
		fmt.Printf("  %s -> %s\n", path, service.URL)
	}

	// 等待中断信号来优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("正在关闭API网关...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("API网关强制关闭:", err)
	}

	fmt.Println("API网关已退出")
}