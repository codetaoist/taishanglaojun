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

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/config"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/logger"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/monitoring"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/registry"
)

// ProxyManager 代理管理器接口
type ProxyManager interface {
	// 处理代理请求
	HandleRequest(c *gin.Context, serviceName string) error
	
	// 健康检查
	HealthCheck()
	
	// 关闭代理管理器
	Close() error
}

// proxyManager 代理管理器实现
type proxyManager struct {
	config          config.ProxyConfig
	registry        registry.Registry
	loadBalancer    LoadBalancer
	healthChecker   *HealthChecker
	metrics         monitoring.Metrics
	logger          logger.Logger
	
	// 代理缓存
	proxies map[string]*httputil.ReverseProxy
	mu      sync.RWMutex
	
	// HTTP客户端
	client *http.Client
}

// New 创建代理管理器
func New(cfg config.ProxyConfig, reg registry.Registry, metrics monitoring.Metrics, log logger.Logger) ProxyManager {
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.MaxIdleConns,
			IdleConnTimeout:     time.Duration(cfg.IdleConnTimeout) * time.Second,
			DisableCompression:  false,
			DisableKeepAlives:   false,
		},
	}
	
	// 创建负载均衡器
	lb := NewLoadBalancer(LoadBalancerType(cfg.LoadBalancer))
	healthAwareLB := NewHealthAwareLoadBalancer(lb)
	
	pm := &proxyManager{
		config:        cfg,
		registry:      reg,
		loadBalancer:  healthAwareLB,
		metrics:       metrics,
		logger:        log,
		proxies:       make(map[string]*httputil.ReverseProxy),
		client:        client,
	}
	
	// 创建健康检查器
	pm.healthChecker = NewHealthChecker(reg, client, log)
	
	return pm
}

// HandleRequest 处理代理请求
func (p *proxyManager) HandleRequest(c *gin.Context, serviceName string) error {
	start := time.Now()
	
	// 发现服务实例
	instances, err := p.registry.Discover(c.Request.Context(), serviceName)
	if err != nil {
		p.metrics.RecordProxyError(serviceName, "service_discovery_failed")
		return fmt.Errorf("service discovery failed: %w", err)
	}
	
	if len(instances) == 0 {
		p.metrics.RecordProxyError(serviceName, "no_instances")
		return fmt.Errorf("no available instances for service: %s", serviceName)
	}
	
	// 选择服务实例
	instance, err := p.loadBalancer.Select(instances)
	if err != nil {
		p.metrics.RecordProxyError(serviceName, "load_balancer_failed")
		return fmt.Errorf("load balancer failed: %w", err)
	}
	
	// 构建目标URL
	targetURL := instance.Meta["url"]
	if targetURL == "" {
		targetURL = fmt.Sprintf("http://%s:%d", instance.Address, instance.Port)
	}
	
	// 获取或创建反向代理
	proxy, err := p.getOrCreateProxy(targetURL)
	if err != nil {
		p.metrics.RecordProxyError(serviceName, "proxy_creation_failed")
		return fmt.Errorf("failed to create proxy: %w", err)
	}
	
	// 设置请求头
	p.setProxyHeaders(c, instance)
	
	// 记录请求开始
	p.logger.WithFields(map[string]interface{}{
		"service":  serviceName,
		"instance": instance.ID,
		"target":   targetURL,
		"path":     c.Request.URL.Path,
		"method":   c.Request.Method,
	}).Debug("Proxying request")
	
	// 执行代理请求
	proxy.ServeHTTP(c.Writer, c.Request)
	
	// 记录指标
	duration := time.Since(start)
	status := fmt.Sprintf("%d", c.Writer.Status())
	p.metrics.RecordProxyRequest(serviceName, c.Request.Method, status, duration)
	
	return nil
}

// getOrCreateProxy 获取或创建反向代理
func (p *proxyManager) getOrCreateProxy(targetURL string) (*httputil.ReverseProxy, error) {
	p.mu.RLock()
	proxy, exists := p.proxies[targetURL]
	p.mu.RUnlock()
	
	if exists {
		return proxy, nil
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// 双重检查
	if proxy, exists := p.proxies[targetURL]; exists {
		return proxy, nil
	}
	
	// 解析目标URL
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid target URL: %w", err)
	}
	
	// 创建反向代理
	proxy = httputil.NewSingleHostReverseProxy(target)
	
	// 自定义Director
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		
		// 修改请求头
		req.Header.Set("X-Forwarded-Proto", "http")
		if req.Header.Get("X-Forwarded-For") == "" {
			req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		}
		
		// 移除Hop-by-hop headers
		p.removeHopByHopHeaders(req.Header)
	}
	
	// 自定义错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		p.logger.WithFields(map[string]interface{}{
			"target": targetURL,
			"error":  err.Error(),
		}).Error("Proxy request failed")
		
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Bad Gateway"))
	}
	
	// 自定义响应修改
	proxy.ModifyResponse = func(resp *http.Response) error {
		// 移除Hop-by-hop headers
		p.removeHopByHopHeaders(resp.Header)
		
		// 添加自定义响应头
		resp.Header.Set("X-Gateway", "taishang-gateway")
		
		return nil
	}
	
	// 缓存代理
	p.proxies[targetURL] = proxy
	
	return proxy, nil
}

// setProxyHeaders 设置代理请求头
func (p *proxyManager) setProxyHeaders(c *gin.Context, instance *registry.ServiceInstance) {
	req := c.Request
	
	// 设置X-Forwarded-* headers
	if req.Header.Get("X-Forwarded-For") == "" {
		req.Header.Set("X-Forwarded-For", c.ClientIP())
	} else {
		req.Header.Set("X-Forwarded-For", req.Header.Get("X-Forwarded-For")+", "+c.ClientIP())
	}
	
	if req.Header.Get("X-Forwarded-Host") == "" {
		req.Header.Set("X-Forwarded-Host", req.Host)
	}
	
	if req.Header.Get("X-Forwarded-Proto") == "" {
		if req.TLS != nil {
			req.Header.Set("X-Forwarded-Proto", "https")
		} else {
			req.Header.Set("X-Forwarded-Proto", "http")
		}
	}
	
	// 设置服务相关headers
	req.Header.Set("X-Service-Name", instance.Name)
	req.Header.Set("X-Service-Instance", instance.ID)
	req.Header.Set("X-Gateway-Time", time.Now().Format(time.RFC3339))
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

// HealthCheck 执行健康检查
func (p *proxyManager) HealthCheck() {
	if p.healthChecker != nil {
		p.healthChecker.CheckAll()
	}
}

// Close 关闭代理管理器
func (p *proxyManager) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// 清空代理缓存
	p.proxies = make(map[string]*httputil.ReverseProxy)
	
	// 关闭HTTP客户端
	if transport, ok := p.client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	
	// 关闭健康检查器
	if p.healthChecker != nil {
		p.healthChecker.Close()
	}
	
	p.logger.Info("Proxy manager closed")
	
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