package gateway

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// reloadRoutes 重新加载路由到Gin引擎
func (g *Gateway) reloadRoutes() error {
	// 获取所有路由
	routes := g.routeManager.ListRoutes()
	
	// 重新初始化路由器
	if err := g.initRouter(); err != nil {
		return fmt.Errorf("failed to reinitialize router: %w", err)
	}
	
	// 注册所有路由
	for _, route := range routes {
		if err := g.registerRoute(route); err != nil {
			g.logger.WithFields(map[string]interface{}{
			"method": route.Method,
			"path":   route.Path,
			"error":  err,
		}).Error("Failed to register route during reload")
			continue
		}
	}
	
	g.logger.WithFields(map[string]interface{}{
		"count": len(routes),
	}).Info("Routes reloaded successfully")
	return nil
}

// registerRoute 注册单个路由到Gin引擎
func (g *Gateway) registerRoute(route *RouteConfig) error {
	// 创建处理器
	handlers := g.createRouteHandlers(route)
	
	// 根据HTTP方法注册路由
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
	
	g.logger.WithFields(map[string]interface{}{
		"method":  route.Method,
		"path":    route.Path,
		"service": route.Service,
	}).Info("Route registered")
	
	return nil
}

// proxyRequest 代理请求到目标服务
func (g *Gateway) proxyRequest(c *gin.Context, route *RouteConfig) error {
	// 获取服务实例
	ctx := c.Request.Context()
	instances, err := g.registry.Discover(ctx, route.Service)
	if err != nil {
		return fmt.Errorf("failed to discover service %s: %w", route.Service, err)
	}
	
	if len(instances) == 0 {
		return fmt.Errorf("no available instances for service %s", route.Service)
	}
	
	// 构建目标URL
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
	
	// 使用代理管理器转发请求
	return g.proxyManager.HandleRequest(c.Writer, c.Request, route.Service)
}

// syncRoutesToManager 同步旧的路由配置到RouteManager
func (g *Gateway) syncRoutesToManager() error {
	g.routesMu.RLock()
	defer g.routesMu.RUnlock()
	
	routes := make([]*RouteConfig, 0, len(g.routes))
	for _, route := range g.routes {
		routes = append(routes, route)
	}
	
	return g.routeManager.LoadRoutes(routes)
}