package router

import (
	"net/http"
	"time"

	"github.com/codetaoist/taishanglaojun/gateway/internal/config"
	"github.com/codetaoist/taishanglaojun/gateway/internal/discovery"
	"github.com/codetaoist/taishanglaojun/gateway/internal/middleware"
	"github.com/codetaoist/taishanglaojun/gateway/internal/proxy"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Setup sets up the router with all routes and middleware
func Setup(cfg *config.Config, proxyManager *proxy.ProxyManager) *gin.Engine {
	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create router
	r := gin.New()

	// Add middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware(cfg))

	// Add custom middleware
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	r.Use(middleware.RequestID())
	r.Use(middleware.Logging(logger))
	r.Use(middleware.Metrics())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "gateway",
		})
	})

	// API routes
	api := r.Group("/api")
	{
		// Service discovery endpoints
		api.GET("/services", listServices(proxyManager))
		api.GET("/services/:name", getService(proxyManager))

		// Gateway configuration endpoints
		api.GET("/config", getGatewayConfig(cfg))
		api.PUT("/config", updateGatewayConfig(cfg))

		// Metrics endpoint
		api.GET("/metrics", middleware.PrometheusHandler())
	}

	// Proxy routes
	setupProxyRoutes(r, cfg, proxyManager)

	// 404 handler
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Endpoint not found",
			"path":  c.Request.URL.Path,
		})
	})

	return r
}

// setupProxyRoutes sets up proxy routes for all configured services
func setupProxyRoutes(r *gin.Engine, cfg *config.Config, proxyManager *proxy.ProxyManager) {
	// Watch for service changes
	serviceNames := make([]string, 0, len(cfg.Services))
	for name := range cfg.Services {
		serviceNames = append(serviceNames, name)
	}
	proxyManager.WatchServices(serviceNames)

	// Create proxy routes for each service
	for name, serviceConfig := range cfg.Services {
		// Create a route group for the service
		group := r.Group(serviceConfig.PathPrefix)

		// Add service-specific middleware
		if serviceConfig.AuthRequired {
			group.Use(middleware.Authentication(cfg.JWT))
		}

		if serviceConfig.RateLimitEnabled {
			group.Use(middleware.RateLimit(serviceConfig.RateLimit))
		}

		// Add the proxy handler for all methods
		group.Any("/*path", proxyManager.ProxyHandler(name))
	}
}

// corsMiddleware returns CORS middleware based on configuration
func corsMiddleware(cfg *config.Config) gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.CORS.AllowedOrigins
	corsConfig.AllowMethods = cfg.CORS.AllowedMethods
	corsConfig.AllowHeaders = cfg.CORS.AllowedHeaders
	corsConfig.AllowCredentials = cfg.CORS.AllowCredentials
	corsConfig.MaxAge = cfg.CORS.MaxAge

	return cors.New(corsConfig)
}

// listServices returns a list of all registered services
func listServices(proxyManager *proxy.ProxyManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// This is a placeholder implementation
		// In a real implementation, we would get the list from the discovery client
		c.JSON(http.StatusOK, gin.H{
			"services": []string{
				"api",
				"auth",
				"notification",
			},
		})
	}
}

// getService returns details about a specific service
func getService(proxyManager *proxy.ProxyManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceName := c.Param("name")

		// This is a placeholder implementation
		// In a real implementation, we would get the service details from the discovery client
		c.JSON(http.StatusOK, gin.H{
			"name":      serviceName,
			"instances": 1,
			"status":    "healthy",
		})
	}
}

// getGatewayConfig returns the current gateway configuration
func getGatewayConfig(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Return a sanitized version of the configuration
		c.JSON(http.StatusOK, gin.H{
			"port":       cfg.Port,
			"environment": cfg.Environment,
			"log_level":  cfg.LogLevel,
			"services":   cfg.Services,
			"cors":       cfg.CORS,
		})
	}
}

// updateGatewayConfig updates the gateway configuration
func updateGatewayConfig(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// This is a placeholder implementation
		// In a real implementation, we would update the configuration
		c.JSON(http.StatusOK, gin.H{
			"message": "Configuration updated successfully",
		})
	}
}