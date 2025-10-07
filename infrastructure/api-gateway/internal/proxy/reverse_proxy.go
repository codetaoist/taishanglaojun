package proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/logger"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/registry"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/types"
)

// ProxyManager 代理管理器接口
type ProxyManager interface {
	// HandleRequest 处理代理请求
	HandleRequest(w http.ResponseWriter, r *http.Request, serviceName string) error
	
	// GetStats 获取统计信息
	GetStats(serviceName string) (map[string]interface{}, error)
	
	// SetLoadBalancerStrategy 设置负载均衡策略
	SetLoadBalancerStrategy(serviceName string, strategy types.LoadBalancerType) error
	
	// Close 关闭代理管理器
	Close() error
	
	// GetLoadBalancerManager 获取负载均衡管理器
	GetLoadBalancerManager() LoadBalancerManager
}

// proxyManager 代理管理器实现
type proxyManager struct {
	registry   registry.Registry
	loadBalancerMgr  LoadBalancerManager
	proxies          map[string]*httputil.ReverseProxy
	mu               sync.RWMutex
	logger           logger.Logger
	timeout          time.Duration
}

// New 创建代理管理器
func New(registry registry.Registry, loadBalancerMgr LoadBalancerManager, logger logger.Logger, timeout time.Duration) ProxyManager {
	return &proxyManager{
		registry:        registry,
		loadBalancerMgr: loadBalancerMgr,
		proxies:         make(map[string]*httputil.ReverseProxy),
		logger:          logger,
		timeout:         timeout,
	}
}

// HandleRequest 处理代理请求
func (pm *proxyManager) HandleRequest(w http.ResponseWriter, r *http.Request, serviceName string) error {
	// 获取客户端IP
	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP = r.Header.Get("X-Real-IP")
	}
	if clientIP == "" {
		clientIP = strings.Split(r.RemoteAddr, ":")[0]
	}
	
	// 发现服务实例
	instances, err := pm.registry.Discover(r.Context(), serviceName)
	if err != nil {
		pm.logger.WithFields(map[string]interface{}{
			"service": serviceName,
			"error":   err.Error(),
		}).Error("Failed to discover service instances")
		http.Error(w, "Service discovery failed", http.StatusServiceUnavailable)
		return err
	}

	if len(instances) == 0 {
		pm.logger.WithFields(map[string]interface{}{
			"service": serviceName,
		}).Warn("No available instances for service")
		http.Error(w, "No service instances available", http.StatusServiceUnavailable)
		return fmt.Errorf("no available instances for service: %s", serviceName)
	}

	// 使用负载均衡管理器选择实例
	instance, err := pm.loadBalancerMgr.SelectInstance(serviceName, instances, clientIP)
	if err != nil {
		pm.logger.WithFields(map[string]interface{}{
			"service": serviceName,
			"error":   err.Error(),
		}).Error("Failed to select service instance")
		http.Error(w, "Load balancer failed", http.StatusInternalServerError)
		return err
	}

	// 构建目标URL
	targetURL := fmt.Sprintf("http://%s:%d", instance.Address, instance.Port)
	
	// 增加连接计数
	pm.loadBalancerMgr.IncrementConnections(serviceName, instance.ID)
	defer pm.loadBalancerMgr.DecrementConnections(serviceName, instance.ID)

	// 获取或创建代理
	proxy, err := pm.getOrCreateProxy(targetURL, nil)
	if err != nil {
		pm.logger.WithFields(map[string]interface{}{
			"service":    serviceName,
			"target_url": targetURL,
			"error":      err.Error(),
		}).Error("Failed to create proxy")
		http.Error(w, "Proxy creation failed", http.StatusInternalServerError)
		return err
	}

	// 执行代理请求
	proxy.ServeHTTP(w, r)
	return nil
}

// getOrCreateProxy 获取或创建反向代理
func (pm *proxyManager) getOrCreateProxy(targetURL string, pathRewrite map[string]string) (*httputil.ReverseProxy, error) {
	pm.mu.RLock()
	proxy, exists := pm.proxies[targetURL]
	pm.mu.RUnlock()

	if exists {
		return proxy, nil
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 双重检查
	if proxy, exists := pm.proxies[targetURL]; exists {
		return proxy, nil
	}

	// 解析目标URL
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid target URL: %w", err)
	}

	// 创建反向代理
	proxy = httputil.NewSingleHostReverseProxy(target)

	// 自定义Director函数
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		
		// 应用路径重写规则
		originalPath := req.URL.Path
		for pattern, replacement := range pathRewrite {
			if strings.HasPrefix(originalPath, pattern) {
				newPath := strings.Replace(originalPath, pattern, replacement, 1)
				req.URL.Path = newPath
				pm.logger.WithFields(map[string]interface{}{
					"original_path": originalPath,
					"new_path":      newPath,
					"pattern":       pattern,
					"replacement":   replacement,
				}).Debug("Applied path rewrite")
				break
			}
		}
		
		// 设置必要的请求头
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		req.Header.Set("X-Forwarded-Proto", "http")
		if req.Header.Get("X-Forwarded-For") == "" {
			req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		}
	}

	// 自定义错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		pm.logger.WithFields(map[string]interface{}{
			"target_url": targetURL,
			"path":       r.URL.Path,
			"error":      err.Error(),
		}).Error("Proxy request failed")
		
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Bad Gateway"))
	}

	// 缓存代理
	pm.proxies[targetURL] = proxy
	return proxy, nil
}

// setProxyHeaders 设置代理请求头
func (p *proxyManager) setProxyHeaders(w http.ResponseWriter, r *http.Request, instance *registry.ServiceInstance) {
	// 设置X-Forwarded-* headers
	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP = r.Header.Get("X-Real-IP")
	}
	if clientIP == "" {
		clientIP = strings.Split(r.RemoteAddr, ":")[0]
	}
	
	if r.Header.Get("X-Forwarded-For") == "" {
		r.Header.Set("X-Forwarded-For", clientIP)
	} else {
		r.Header.Set("X-Forwarded-For", r.Header.Get("X-Forwarded-For")+", "+clientIP)
	}
	
	if r.Header.Get("X-Forwarded-Host") == "" {
		r.Header.Set("X-Forwarded-Host", r.Host)
	}
	
	if r.Header.Get("X-Forwarded-Proto") == "" {
		if r.TLS != nil {
			r.Header.Set("X-Forwarded-Proto", "https")
		} else {
			r.Header.Set("X-Forwarded-Proto", "http")
		}
	}
	
	// 设置服务相关headers
	r.Header.Set("X-Service-Name", instance.Name)
	r.Header.Set("X-Service-Instance", instance.ID)
	r.Header.Set("X-Gateway-Time", time.Now().Format(time.RFC3339))
}

// removeHopByHopHeaders 移除逐跳headers
func (p *proxyManager) removeHopByHopHeaders(header http.Header) {
	// RFC 2616定义的逐跳headers
	hopByHopHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}
	
	for _, h := range hopByHopHeaders {
		header.Del(h)
	}
	
	// 处理Connection header中指定的headers
	if connections := header.Get("Connection"); connections != "" {
		for _, connection := range strings.Split(connections, ",") {
			header.Del(strings.TrimSpace(connection))
		}
	}
}

// GetStats 获取代理统计信息
func (pm *proxyManager) GetStats(serviceName string) (map[string]interface{}, error) {
	stats := pm.loadBalancerMgr.GetStats(serviceName)
	
	pm.mu.RLock()
	stats["cached_proxies"] = len(pm.proxies)
	pm.mu.RUnlock()
	
	return stats, nil
}

// SetLoadBalancerStrategy 设置负载均衡策略
func (pm *proxyManager) SetLoadBalancerStrategy(serviceName string, strategy types.LoadBalancerType) error {
	return pm.loadBalancerMgr.SetLoadBalancerStrategy(serviceName, strategy)
}

// GetLoadBalancerManager 获取负载均衡管理器
func (pm *proxyManager) GetLoadBalancerManager() LoadBalancerManager {
	return pm.loadBalancerMgr
}

// Close 关闭代理管理器
func (pm *proxyManager) Close() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	// 清理代理缓存
	pm.proxies = make(map[string]*httputil.ReverseProxy)
	
	// 关闭负载均衡管理器
	if err := pm.loadBalancerMgr.Close(); err != nil {
		pm.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Failed to close load balancer manager")
		return err
	}
	
	pm.logger.Info("Proxy manager closed")
	return nil
}

// StreamingProxy 流式代理（用于WebSocket等长连接）
type StreamingProxy struct {
	target *url.URL
	logger logger.Logger
}

// NewStreamingProxy 创建流式代理
func NewStreamingProxy(targetURL string, log logger.Logger) (*StreamingProxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid target URL: %w", err)
	}
	
	return &StreamingProxy{
		target: target,
		logger: log,
	}, nil
}

// ServeHTTP 处理流式代理请求
func (s *StreamingProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 创建到后端的连接
	backendURL := *s.target
	backendURL.Path = r.URL.Path
	backendURL.RawQuery = r.URL.RawQuery
	
	// 创建后端请求
	backendReq, err := http.NewRequestWithContext(r.Context(), r.Method, backendURL.String(), r.Body)
	if err != nil {
		s.logger.Errorf("Failed to create backend request: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	// 复制headers
	for key, values := range r.Header {
		for _, value := range values {
			backendReq.Header.Add(key, value)
		}
	}
	
	// 执行请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	resp, err := client.Do(backendReq)
	if err != nil {
		s.logger.Errorf("Backend request failed: %v", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	
	// 复制响应headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	
	// 设置状态码
	w.WriteHeader(resp.StatusCode)
	
	// 流式复制响应体
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		s.logger.Errorf("Failed to copy response body: %v", err)
	}
}