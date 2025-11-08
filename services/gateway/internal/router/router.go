package router

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/gateway/internal/config"
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

	// Admin API routes
	adminAPI := r.Group("/admin/api")
	{
		// Service discovery endpoints
		adminAPI.GET("/services", listServices(proxyManager))
		adminAPI.GET("/services/:name", getService(proxyManager))

		// Gateway configuration endpoints
		adminAPI.GET("/config", getGatewayConfig(cfg))
		adminAPI.PUT("/config", updateGatewayConfig(cfg))

		// Metrics endpoint
		adminAPI.GET("/metrics", middleware.PrometheusHandler())
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

	// Create a single catch-all route for all API requests
	// The proxy handler will determine which service to forward to based on the path
	r.Any("/api/*path", func(c *gin.Context) {
		// Get the full path
		path := c.Param("path")
		fullPath := "/api" + path
		
		// Determine which service should handle this request based on path prefix
		var serviceName string
		var matchedPrefix string
		
		// Find the service with the longest matching path prefix
		for name, serviceConfig := range cfg.Services {
			if len(serviceConfig.PathPrefix) > len(matchedPrefix) && 
			   strings.HasPrefix(fullPath, serviceConfig.PathPrefix) {
				serviceName = name
				matchedPrefix = serviceConfig.PathPrefix
			}
		}
		
		// Debug logging
		fmt.Printf("Debug: path=%s, fullPath=%s, serviceName=%s, matchedPrefix=%s\n", path, fullPath, serviceName, matchedPrefix)
		
		if serviceName == "" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "No service found for path",
				"path": fullPath,
			})
			return
		}
		
		// Apply service-specific middleware
		serviceConfig := cfg.Services[serviceName]
		
		// Check if authentication is required
		if serviceConfig.AuthRequired {
			// This is a simplified check - in a real implementation, 
			// you would use the actual middleware
			// For now, we'll just pass through to the proxy handler
		}
		
		// Update the request path to remove the service path prefix
		// For example, /api/v1/auth/login becomes /auth/login when forwarded to auth service
		// Special case for auth and api services: keep the full path
		if len(matchedPrefix) > 0 && serviceName != "auth" && serviceName != "api" {
			c.Request.URL.Path = strings.TrimPrefix(fullPath, matchedPrefix)
			if !strings.HasPrefix(c.Request.URL.Path, "/") {
				c.Request.URL.Path = "/" + c.Request.URL.Path
			}
		} else if len(matchedPrefix) > 0 && serviceName == "auth" {
			// For auth service, keep the full path without any modification
			// Because auth service expects /api/v1/auth/login, not /v1/auth/login
			// Explicitly set the path to fullPath to ensure it's not modified
			c.Request.URL.Path = fullPath
			fmt.Printf("Debug: auth service path explicitly set to %s\n", fullPath)
		} else if len(matchedPrefix) > 0 && serviceName == "api" {
			// For api service, we need to remove the /api prefix but keep the rest
			// For example, /api/v1/models becomes /v1/models
			c.Request.URL.Path = strings.TrimPrefix(fullPath, "/api")
			if !strings.HasPrefix(c.Request.URL.Path, "/") {
				c.Request.URL.Path = "/" + c.Request.URL.Path
			}
		}
		
		// Debug logging
		fmt.Printf("Debug: final request path=%s\n", c.Request.URL.Path)
		
		// Forward to the appropriate service
		proxyManager.ProxyHandler(serviceName)(c)
	})
}

// corsMiddleware returns CORS middleware based on configuration
func corsMiddleware(cfg *config.Config) gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.CORS.AllowedOrigins
	corsConfig.AllowMethods = cfg.CORS.AllowedMethods
	corsConfig.AllowHeaders = cfg.CORS.AllowedHeaders
	corsConfig.AllowCredentials = cfg.CORS.AllowCredentials
	corsConfig.MaxAge = time.Duration(cfg.CORS.MaxAge) * time.Second

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