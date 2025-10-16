package gateway

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/logger"
	"github.com/gin-gonic/gin"
)

// DynamicRouter 动态路由管理器
type DynamicRouter struct {
	gateway      *Gateway
	routeManager *RouteManager
	logger       logger.Logger
	mu           sync.RWMutex
	
	// 路由配置源
	configSources []RouteConfigSource
	
	// 路由变更通知
	changeNotifiers []RouteChangeNotifier
	
	// 停止信号
	stopCh chan struct{}
}

// RouteConfigSource 路由配置源接口
type RouteConfigSource interface {
	// 获取路由配置
	GetRoutes(ctx context.Context) ([]*RouteConfig, error)
	
	// 监听路由配置变化
	Watch(ctx context.Context) (<-chan []*RouteConfig, error)
	
	// 关闭配置源
	Close() error
}

// RouteChangeNotifier 路由变更通知器接口
type RouteChangeNotifier interface {
	// 通知路由变更
	NotifyChange(ctx context.Context, changeType RouteChangeType, route *RouteConfig) error
}

// RouteChangeType 路由变更类型
type RouteChangeType string

const (
	RouteChangeTypeAdd    RouteChangeType = "add"
	RouteChangeTypeUpdate RouteChangeType = "update"
	RouteChangeTypeDelete RouteChangeType = "delete"
)

// RouteConfig 路由配置
type RouteConfig struct {
	ID          string            `json:"id" yaml:"id"`
	Method      string            `json:"method" yaml:"method"`
	Path        string            `json:"path" yaml:"path"`
	Service     string            `json:"service" yaml:"service"`
	Rewrite     string            `json:"rewrite,omitempty" yaml:"rewrite,omitempty"`
	StripPrefix bool              `json:"strip_prefix,omitempty" yaml:"strip_prefix,omitempty"`
	Timeout     time.Duration     `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Middleware  []string          `json:"middleware,omitempty" yaml:"middleware,omitempty"`
	Headers     map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	RateLimit   *RateLimitConfig  `json:"rate_limit,omitempty" yaml:"rate_limit,omitempty"`
	Auth        *AuthConfig       `json:"auth,omitempty" yaml:"auth,omitempty"`
	Enabled     bool              `json:"enabled" yaml:"enabled"`
	CreatedAt   time.Time         `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" yaml:"updated_at"`
}

// RateLimitConfig 路由级别限流配置
type RateLimitConfig struct {
	Rate  int `json:"rate" yaml:"rate"`
	Burst int `json:"burst" yaml:"burst"`
}

// AuthConfig 路由级别认证配置
type AuthConfig struct {
	Required bool     `json:"required" yaml:"required"`
	Roles    []string `json:"roles,omitempty" yaml:"roles,omitempty"`
	Scopes   []string `json:"scopes,omitempty" yaml:"scopes,omitempty"`
}

// NewDynamicRouter 创建动态路由管理器
func NewDynamicRouter(gateway *Gateway, routeManager *RouteManager, log logger.Logger) *DynamicRouter {
	return &DynamicRouter{
		gateway:         gateway,
		routeManager:    routeManager,
		logger:          log,
		configSources:   make([]RouteConfigSource, 0),
		changeNotifiers: make([]RouteChangeNotifier, 0),
		stopCh:          make(chan struct{}),
	}
}

// AddConfigSource 添加路由配置源
func (dr *DynamicRouter) AddConfigSource(source RouteConfigSource) {
	dr.mu.Lock()
	defer dr.mu.Unlock()
	
	dr.configSources = append(dr.configSources, source)
	dr.logger.Info("Route config source added")
}

// AddChangeNotifier 添加路由变更通知器
func (dr *DynamicRouter) AddChangeNotifier(notifier RouteChangeNotifier) {
	dr.mu.Lock()
	defer dr.mu.Unlock()
	
	dr.changeNotifiers = append(dr.changeNotifiers, notifier)
	dr.logger.Info("Route change notifier added")
}

// Start 启动动态路由管理器
func (dr *DynamicRouter) Start(ctx context.Context) error {
	// 初始加载路由配置
	if err := dr.loadInitialRoutes(ctx); err != nil {
		return fmt.Errorf("failed to load initial routes: %w", err)
	}
	
	// 启动配置监听
	go dr.watchConfigChanges(ctx)
	
	dr.logger.Info("Dynamic router started")
	return nil
}

// Stop 停止动态路由管理器
func (dr *DynamicRouter) Stop() error {
	close(dr.stopCh)
	
	// 关闭所有配置源
	for _, source := range dr.configSources {
		if err := source.Close(); err != nil {
			dr.logger.Errorf("Failed to close config source: %v", err)
		}
	}
	
	dr.logger.Info("Dynamic router stopped")
	return nil
}

// loadInitialRoutes 加载初始路由配置
func (dr *DynamicRouter) loadInitialRoutes(ctx context.Context) error {
	dr.mu.RLock()
	sources := make([]RouteConfigSource, len(dr.configSources))
	copy(sources, dr.configSources)
	dr.mu.RUnlock()
	
	var allRoutes []*RouteConfig
	
	// 从所有配置源加载路由
	for _, source := range sources {
		routes, err := source.GetRoutes(ctx)
		if err != nil {
			dr.logger.Errorf("Failed to get routes from config source: %v", err)
			continue
		}
		allRoutes = append(allRoutes, routes...)
	}
	
	// 加载到路由管理器
	if err := dr.routeManager.LoadRoutes(allRoutes); err != nil {
		return fmt.Errorf("failed to load routes to manager: %w", err)
	}
	
	// 重新加载网关路由
	if err := dr.gateway.reloadRoutes(); err != nil {
		return fmt.Errorf("failed to reload gateway routes: %w", err)
	}
	
	dr.logger.WithFields(map[string]interface{}{
		"count": len(allRoutes),
	}).Info("Initial routes loaded")
	
	return nil
}

// watchConfigChanges 监听配置变化
func (dr *DynamicRouter) watchConfigChanges(ctx context.Context) {
	dr.mu.RLock()
	sources := make([]RouteConfigSource, len(dr.configSources))
	copy(sources, dr.configSources)
	dr.mu.RUnlock()
	
	// 为每个配置源启动监听
	for _, source := range sources {
		go dr.watchSingleSource(ctx, source)
	}
}

// watchSingleSource 监听单个配置源的变化
func (dr *DynamicRouter) watchSingleSource(ctx context.Context, source RouteConfigSource) {
	watchCh, err := source.Watch(ctx)
	if err != nil {
		dr.logger.Errorf("Failed to watch config source: %v", err)
		return
	}
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-dr.stopCh:
			return
		case routes := <-watchCh:
			if routes == nil {
				continue
			}
			
			if err := dr.handleRouteChanges(ctx, routes); err != nil {
				dr.logger.Errorf("Failed to handle route changes: %v", err)
			}
		}
	}
}

// handleRouteChanges 处理路由变化
func (dr *DynamicRouter) handleRouteChanges(ctx context.Context, newRoutes []*RouteConfig) error {
	// 获取当前路由
	currentRoutes := dr.routeManager.ListRoutes()
	
	// 构建路由映射
	currentMap := make(map[string]*RouteConfig)
	for _, route := range currentRoutes {
		key := fmt.Sprintf("%s:%s", route.Method, route.Path)
		currentMap[key] = route
	}
	
	newMap := make(map[string]*RouteConfig)
	for _, route := range newRoutes {
		key := fmt.Sprintf("%s:%s", route.Method, route.Path)
		newMap[key] = route
	}
	
	// 检测变化
	var changes []RouteChange
	
	// 检测新增和更新
	for key, newRoute := range newMap {
		if currentRoute, exists := currentMap[key]; exists {
			// 检查是否有更新
			if !dr.routesEqual(currentRoute, newRoute) {
				changes = append(changes, RouteChange{
					Type:  RouteChangeTypeUpdate,
					Route: newRoute,
				})
			}
		} else {
			// 新增路由
			changes = append(changes, RouteChange{
				Type:  RouteChangeTypeAdd,
				Route: newRoute,
			})
		}
	}
	
	// 检测删除
	for key, currentRoute := range currentMap {
		if _, exists := newMap[key]; !exists {
			changes = append(changes, RouteChange{
				Type:  RouteChangeTypeDelete,
				Route: currentRoute,
			})
		}
	}
	
	// 应用变化
	if len(changes) > 0 {
		if err := dr.applyRouteChanges(ctx, changes); err != nil {
			return fmt.Errorf("failed to apply route changes: %w", err)
		}
	}
	
	return nil
}

// RouteChange 路由变化
type RouteChange struct {
	Type  RouteChangeType
	Route *RouteConfig
}

// applyRouteChanges 应用路由变化
func (dr *DynamicRouter) applyRouteChanges(ctx context.Context, changes []RouteChange) error {
	for _, change := range changes {
		switch change.Type {
		case RouteChangeTypeAdd:
			if err := dr.routeManager.AddRoute(change.Route); err != nil {
				dr.logger.Errorf("Failed to add route %s %s: %v", 
					change.Route.Method, change.Route.Path, err)
				continue
			}
			
		case RouteChangeTypeUpdate:
			if err := dr.routeManager.UpdateRoute(change.Route.Method, change.Route.Path, change.Route); err != nil {
				dr.logger.Errorf("Failed to update route %s %s: %v", 
					change.Route.Method, change.Route.Path, err)
				continue
			}
			
		case RouteChangeTypeDelete:
			if _, err := dr.routeManager.DeleteRoute(change.Route.Method, change.Route.Path); err != nil {
				dr.logger.Errorf("Failed to delete route %s %s: %v", 
					change.Route.Method, change.Route.Path, err)
				continue
			}
		}
		
		// 通知变更
		dr.notifyChange(ctx, change.Type, change.Route)
	}
	
	// 重新加载网关路由
	if err := dr.gateway.reloadRoutes(); err != nil {
		return fmt.Errorf("failed to reload gateway routes: %w", err)
	}
	
	dr.logger.WithFields(map[string]interface{}{
		"changes": len(changes),
	}).Info("Route changes applied")
	
	return nil
}

// notifyChange 通知路由变更
func (dr *DynamicRouter) notifyChange(ctx context.Context, changeType RouteChangeType, route *RouteConfig) {
	dr.mu.RLock()
	notifiers := make([]RouteChangeNotifier, len(dr.changeNotifiers))
	copy(notifiers, dr.changeNotifiers)
	dr.mu.RUnlock()
	
	for _, notifier := range notifiers {
		if err := notifier.NotifyChange(ctx, changeType, route); err != nil {
			dr.logger.Errorf("Failed to notify route change: %v", err)
		}
	}
}

// routesEqual 比较两个路由是否相等
func (dr *DynamicRouter) routesEqual(route1, route2 *RouteConfig) bool {
	// 简单比较，实际应该更全面
	return route1.Method == route2.Method &&
		route1.Path == route2.Path &&
		route1.Service == route2.Service &&
		route1.Rewrite == route2.Rewrite &&
		route1.StripPrefix == route2.StripPrefix &&
		route1.Timeout == route2.Timeout &&
		route1.Enabled == route2.Enabled
}

// GetRouteStats 获取路由统计信息
func (dr *DynamicRouter) GetRouteStats() map[string]interface{} {
	routes := dr.routeManager.ListRoutes()
	
	stats := map[string]interface{}{
		"total_routes": len(routes),
		"enabled_routes": 0,
		"disabled_routes": 0,
		"methods": make(map[string]int),
		"services": make(map[string]int),
	}
	
	methods := make(map[string]int)
	services := make(map[string]int)
	
	for _, route := range routes {
		if route.Enabled {
			stats["enabled_routes"] = stats["enabled_routes"].(int) + 1
		} else {
			stats["disabled_routes"] = stats["disabled_routes"].(int) + 1
		}
		
		methods[route.Method]++
		services[route.Service]++
	}
	
	stats["methods"] = methods
	stats["services"] = services
	
	return stats
}

// RegisterManagementRoutes 注册路由管理API
func (dr *DynamicRouter) RegisterManagementRoutes(router *gin.Engine) {
	api := router.Group("/api/v1/routes")
	
	// 获取所有路由
	api.GET("", dr.handleListRoutes)
	
	// 获取单个路由
	api.GET("/:method/*path", dr.handleGetRoute)
	
	// 添加路由
	api.POST("", dr.handleAddRoute)
	
	// 更新路由
	api.PUT("/:method/*path", dr.handleUpdateRoute)
	
	// 删除路由
	api.DELETE("/:method/*path", dr.handleDeleteRoute)
	
	// 获取路由统计
	api.GET("/stats", dr.handleGetRouteStats)
	
	// 重新加载路由
	api.POST("/reload", dr.handleReloadRoutes)
}

// handleListRoutes 处理获取所有路由请求
func (dr *DynamicRouter) handleListRoutes(c *gin.Context) {
	routes := dr.routeManager.ListRoutes()
	c.JSON(http.StatusOK, gin.H{
		"routes": routes,
		"count":  len(routes),
	})
}

// handleGetRoute 处理获取单个路由请求
func (dr *DynamicRouter) handleGetRoute(c *gin.Context) {
	method := c.Param("method")
	path := c.Param("path")
	
	route, exists := dr.routeManager.GetRoute(method, path)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Route not found",
		})
		return
	}
	
	c.JSON(http.StatusOK, route)
}

// handleAddRoute 处理添加路由请求
func (dr *DynamicRouter) handleAddRoute(c *gin.Context) {
	var route RouteConfig
	if err := c.ShouldBindJSON(&route); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}
	
	// 设置时间戳
	route.CreatedAt = time.Now()
	route.UpdatedAt = time.Now()
	
	if err := dr.routeManager.AddRoute(&route); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": fmt.Sprintf("Failed to add route: %v", err),
		})
		return
	}
	
	// 重新加载路由
	if err := dr.gateway.reloadRoutes(); err != nil {
		dr.logger.Errorf("Failed to reload routes after adding: %v", err)
	}
	
	// 通知变更
	dr.notifyChange(c.Request.Context(), RouteChangeTypeAdd, &route)
	
	c.JSON(http.StatusCreated, route)
}

// handleUpdateRoute 处理更新路由请求
func (dr *DynamicRouter) handleUpdateRoute(c *gin.Context) {
	method := c.Param("method")
	path := c.Param("path")
	
	var route RouteConfig
	if err := c.ShouldBindJSON(&route); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}
	
	// 设置更新时间
	route.UpdatedAt = time.Now()
	
	if err := dr.routeManager.UpdateRoute(method, path, &route); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Failed to update route: %v", err),
		})
		return
	}
	
	// 重新加载路由
	if err := dr.gateway.reloadRoutes(); err != nil {
		dr.logger.Errorf("Failed to reload routes after updating: %v", err)
	}
	
	// 通知变更
	dr.notifyChange(c.Request.Context(), RouteChangeTypeUpdate, &route)
	
	c.JSON(http.StatusOK, route)
}

// handleDeleteRoute 处理删除路由请求
func (dr *DynamicRouter) handleDeleteRoute(c *gin.Context) {
	method := c.Param("method")
	path := c.Param("path")
	
	route, err := dr.routeManager.DeleteRoute(method, path)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Failed to delete route: %v", err),
		})
		return
	}
	
	// 重新加载路由
	if err := dr.gateway.reloadRoutes(); err != nil {
		dr.logger.Errorf("Failed to reload routes after deleting: %v", err)
	}
	
	// 通知变更
	dr.notifyChange(c.Request.Context(), RouteChangeTypeDelete, route)
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Route deleted successfully",
		"route":   route,
	})
}

// handleGetRouteStats 处理获取路由统计请求
func (dr *DynamicRouter) handleGetRouteStats(c *gin.Context) {
	stats := dr.GetRouteStats()
	c.JSON(http.StatusOK, stats)
}

// handleReloadRoutes 处理重新加载路由请求
func (dr *DynamicRouter) handleReloadRoutes(c *gin.Context) {
	if err := dr.loadInitialRoutes(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to reload routes: %v", err),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Routes reloaded successfully",
	})
}