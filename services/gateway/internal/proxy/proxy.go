package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/gateway/internal/discovery"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoadBalancer defines the load balancing strategy
type LoadBalancer interface {
	SelectService(services []*discovery.Service) *discovery.Service
}

// RoundRobinLoadBalancer implements round-robin load balancing
type RoundRobinLoadBalancer struct {
	current int
}

// NewRoundRobinLoadBalancer creates a new round-robin load balancer
func NewRoundRobinLoadBalancer() *RoundRobinLoadBalancer {
	return &RoundRobinLoadBalancer{current: 0}
}

// SelectService selects a service using round-robin strategy
func (lb *RoundRobinLoadBalancer) SelectService(services []*discovery.Service) *discovery.Service {
	if len(services) == 0 {
		return nil
	}

	service := services[lb.current%len(services)]
	lb.current++
	return service
}

// RandomLoadBalancer implements random load balancing
type RandomLoadBalancer struct{}

// NewRandomLoadBalancer creates a new random load balancer
func NewRandomLoadBalancer() *RandomLoadBalancer {
	return &RandomLoadBalancer{}
}

// SelectService selects a service using random strategy
func (lb *RandomLoadBalancer) SelectService(services []*discovery.Service) *discovery.Service {
	if len(services) == 0 {
		return nil
	}

	// In a real implementation, we would use math/rand here
	// For simplicity, we'll just return the first service
	return services[0]
}

// ProxyManager manages reverse proxies for services
type ProxyManager struct {
	discoveryClient discovery.ServiceDiscovery
	proxies         map[string]*httputil.ReverseProxy
	loadBalancer    LoadBalancer
	logger          *zap.Logger
	config          Config
}

// Config holds configuration for the proxy manager
type Config struct {
	Timeout             time.Duration `mapstructure:"timeout"`
	HealthCheckInterval time.Duration `mapstructure:"health_check_interval"`
	RetryAttempts       int           `mapstructure:"retry_attempts"`
}

// NewManager creates a new proxy manager
func NewManager(discoveryClient discovery.ServiceDiscovery, config Config) *ProxyManager {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	return &ProxyManager{
		discoveryClient: discoveryClient,
		proxies:         make(map[string]*httputil.ReverseProxy),
		loadBalancer:    NewRoundRobinLoadBalancer(),
		logger:          logger,
		config:          config,
	}
}

// SetLoadBalancer sets the load balancing strategy
func (pm *ProxyManager) SetLoadBalancer(lb LoadBalancer) {
	pm.loadBalancer = lb
}

// CreateProxy creates a reverse proxy for a service
func (pm *ProxyManager) CreateProxy(serviceName string) (*httputil.ReverseProxy, error) {
	// Get service instances
	services, err := pm.discoveryClient.GetService(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get service %s: %w", serviceName, err)
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("no instances found for service %s", serviceName)
	}

	// Select a service using the load balancer
	service := pm.loadBalancer.SelectService(services)
	if service == nil {
		return nil, fmt.Errorf("failed to select service instance for %s", serviceName)
	}

	// Create target URL
	target, err := url.Parse(fmt.Sprintf("http://%s:%d", service.Address, service.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to parse target URL: %w", err)
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Set up error handler
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		pm.logger.Error("proxy error",
			zap.Error(err),
			zap.String("service", serviceName),
			zap.String("method", req.Method),
			zap.String("path", req.URL.Path),
		)

		// Try to get a new service instance and retry
		if pm.retryRequest(serviceName, rw, req) {
			return
		}

		// If all retries failed, return an error response
		rw.WriteHeader(http.StatusBadGateway)
		rw.Write([]byte(`{"error": "Service unavailable"}`))
	}

	// Set up director to modify the request
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Add custom headers
		req.Header.Set("X-Gateway-Service", serviceName)
		req.Header.Set("X-Gateway-Timestamp", time.Now().Format(time.RFC3339))

		// Modify host header
		req.Host = target.Host
	}

	return proxy, nil
}

// retryRequest retries a request with a different service instance
func (pm *ProxyManager) retryRequest(serviceName string, rw http.ResponseWriter, req *http.Request) bool {
	for i := 0; i < pm.config.RetryAttempts; i++ {
		// Get fresh service instances
		services, err := pm.discoveryClient.GetService(serviceName)
		if err != nil {
			pm.logger.Error("failed to get service during retry",
				zap.Error(err),
				zap.String("service", serviceName),
				zap.Int("attempt", i+1),
			)
			continue
		}

		if len(services) == 0 {
			continue
		}

		// Select a different service instance
		service := pm.loadBalancer.SelectService(services)
		if service == nil {
			continue
		}

		// Create a new proxy for this instance
		proxy, err := pm.CreateProxy(serviceName)
		if err != nil {
			pm.logger.Error("failed to create proxy during retry",
				zap.Error(err),
				zap.String("service", serviceName),
				zap.Int("attempt", i+1),
			)
			continue
		}

		// Create a new request with the same body
		body, err := io.ReadAll(req.Body)
		if err != nil {
			pm.logger.Error("failed to read request body during retry",
				zap.Error(err),
				zap.String("service", serviceName),
				zap.Int("attempt", i+1),
			)
			continue
		}

		req.Body = io.NopCloser(bytes.NewBuffer(body))

		// Try the request again
		proxy.ServeHTTP(rw, req)
		return true
	}

	return false
}

// GetProxy gets or creates a proxy for a service
func (pm *ProxyManager) GetProxy(serviceName string) (*httputil.ReverseProxy, error) {
	if proxy, ok := pm.proxies[serviceName]; ok {
		return proxy, nil
	}

	proxy, err := pm.CreateProxy(serviceName)
	if err != nil {
		return nil, err
	}

	pm.proxies[serviceName] = proxy
	return proxy, nil
}

// ProxyHandler creates a Gin handler for proxying requests
func (pm *ProxyManager) ProxyHandler(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get proxy for the service
		proxy, err := pm.GetProxy(serviceName)
		if err != nil {
			pm.logger.Error("failed to get proxy",
				zap.Error(err),
				zap.String("service", serviceName),
			)

			c.JSON(http.StatusBadGateway, gin.H{
				"error": fmt.Sprintf("Service %s is unavailable", serviceName),
			})
			return
		}

		// Create a new request with the same body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			pm.logger.Error("failed to read request body",
				zap.Error(err),
				zap.String("service", serviceName),
			)

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read request body",
			})
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// Update the request path to remove the service prefix
		// For example, /api/users/users/123 becomes /users/123
		path := c.Request.URL.Path
		prefix := "/" + serviceName
		if strings.HasPrefix(path, prefix) {
			c.Request.URL.Path = strings.TrimPrefix(path, prefix)
			if !strings.HasPrefix(c.Request.URL.Path, "/") {
				c.Request.URL.Path = "/" + c.Request.URL.Path
			}
		}

		// Proxy the request
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// WatchServices watches for changes in services and updates proxies
func (pm *ProxyManager) WatchServices(serviceNames []string) {
	for _, serviceName := range serviceNames {
		go func(name string) {
			watcher, err := pm.discoveryClient.WatchService(name)
			if err != nil {
				pm.logger.Error("failed to watch service",
					zap.Error(err),
					zap.String("service", name),
				)
				return
			}

			for services := range watcher {
				pm.logger.Info("service updated",
					zap.String("service", name),
					zap.Int("instances", len(services)),
				)

				// Invalidate the proxy cache for this service
				delete(pm.proxies, name)
			}
		}(serviceName)
	}
}

// Close closes the proxy manager
func (pm *ProxyManager) Close() error {
	// Clear proxy cache
	pm.proxies = make(map[string]*httputil.ReverseProxy)
	return nil
}