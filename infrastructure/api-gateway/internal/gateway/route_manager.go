package gateway

import (
	"fmt"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/logger"
)

// RouteManager 路由管理器
type RouteManager struct {
	routes map[string]*RouteConfig
	mu     sync.RWMutex
	logger logger.Logger
}

// NewRouteManager 创建路由管理器
func NewRouteManager(log logger.Logger) *RouteManager {
	return &RouteManager{
		routes: make(map[string]*RouteConfig),
		logger: log,
	}
}

// AddRoute 添加路由
func (rm *RouteManager) AddRoute(route *RouteConfig) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	routeKey := fmt.Sprintf("%s:%s", route.Method, route.Path)
	
	// 检查路由是否已存在
	if _, exists := rm.routes[routeKey]; exists {
		return fmt.Errorf("route already exists: %s %s", route.Method, route.Path)
	}

	// 验证路由配置
	if err := rm.validateRoute(route); err != nil {
		return fmt.Errorf("invalid route configuration: %w", err)
	}

	rm.routes[routeKey] = route
	rm.logger.WithFields(map[string]interface{}{
		"method":  route.Method,
		"path":    route.Path,
		"service": route.Service,
	}).Info("Route added")

	return nil
}

// UpdateRoute 更新路由
func (rm *RouteManager) UpdateRoute(method, path string, updatedRoute *RouteConfig) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	routeKey := fmt.Sprintf("%s:%s", method, path)
	
	// 检查路由是否存在
	if _, exists := rm.routes[routeKey]; !exists {
		return fmt.Errorf("route not found: %s %s", method, path)
	}

	// 验证更新的路由配置
	if err := rm.validateRoute(updatedRoute); err != nil {
		return fmt.Errorf("invalid route configuration: %w", err)
	}

	// 设置路由的方法和路径
	updatedRoute.Method = method
	updatedRoute.Path = path

	rm.routes[routeKey] = updatedRoute
	rm.logger.WithFields(map[string]interface{}{
		"method":  method,
		"path":    path,
		"service": updatedRoute.Service,
	}).Info("Route updated")

	return nil
}

// DeleteRoute 删除路由
func (rm *RouteManager) DeleteRoute(method, path string) (*RouteConfig, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	routeKey := fmt.Sprintf("%s:%s", method, path)
	
	// 检查路由是否存在
	route, exists := rm.routes[routeKey]
	if !exists {
		return nil, fmt.Errorf("route not found: %s %s", method, path)
	}

	delete(rm.routes, routeKey)
	rm.logger.WithFields(map[string]interface{}{
		"method":  method,
		"path":    path,
		"service": route.Service,
	}).Info("Route deleted")

	return route, nil
}

// GetRoute 获取路由
func (rm *RouteManager) GetRoute(method, path string) (*RouteConfig, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	routeKey := fmt.Sprintf("%s:%s", method, path)
	route, exists := rm.routes[routeKey]
	return route, exists
}

// ListRoutes 列出所有路由
func (rm *RouteManager) ListRoutes() []*RouteConfig {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	routes := make([]*RouteConfig, 0, len(rm.routes))
	for _, route := range rm.routes {
		routes = append(routes, route)
	}

	return routes
}

// GetRouteCount 获取路由数量
func (rm *RouteManager) GetRouteCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return len(rm.routes)
}

// validateRoute 验证路由配置
func (rm *RouteManager) validateRoute(route *RouteConfig) error {
	if route.Path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	if route.Method == "" {
		return fmt.Errorf("method cannot be empty")
	}

	if route.Service == "" {
		return fmt.Errorf("service cannot be empty")
	}

	// 验证HTTP方法
	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "DELETE": true,
		"PATCH": true, "HEAD": true, "OPTIONS": true, "ANY": true,
	}

	if !validMethods[route.Method] {
		return fmt.Errorf("invalid HTTP method: %s", route.Method)
	}

	// 验证超时时间
	if route.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	// 如果没有设置超时时间，使用默认值
	if route.Timeout == 0 {
		route.Timeout = 30 * time.Second
	}

	return nil
}

// ClearRoutes 清空所有路由
func (rm *RouteManager) ClearRoutes() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.routes = make(map[string]*RouteConfig)
	rm.logger.Info("All routes cleared")
}

// LoadRoutes 批量加载路由
func (rm *RouteManager) LoadRoutes(routes []*RouteConfig) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// 清空现有路由
	rm.routes = make(map[string]*RouteConfig)

	// 加载新路由
	for _, route := range routes {
		routeKey := fmt.Sprintf("%s:%s", route.Method, route.Path)
		
		// 验证路由配置
		if err := rm.validateRoute(route); err != nil {
			rm.logger.Errorf("Invalid route configuration for %s %s: %v", route.Method, route.Path, err)
			continue
		}

		rm.routes[routeKey] = route
	}

	rm.logger.WithFields(map[string]interface{}{
		"count": len(rm.routes),
	}).Info("Routes loaded")

	return nil
}