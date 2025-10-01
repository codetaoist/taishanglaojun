package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"api-gateway/internal/config"
	"api-gateway/internal/gateway"
	"api-gateway/internal/logger"
	"api-gateway/internal/monitoring"
	"api-gateway/internal/proxy"
	"api-gateway/internal/registry"
)

// 测试配置
var testConfig = &config.Config{
	Server: config.ServerConfig{
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	},
	Logging: config.LoggingConfig{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	},
	Monitoring: config.MonitoringConfig{
		Enabled: true,
		Port:    9090,
		Path:    "/metrics",
	},
	Registry: config.RegistryConfig{
		Type: "static",
	},
	Proxy: config.ProxyConfig{
		Timeout:         30 * time.Second,
		MaxIdleConns:    100,
		IdleConnTimeout: 90 * time.Second,
	},
	Security: config.SecurityConfig{
		JWT: config.JWTConfig{
			Secret:     "test-secret",
			Expiration: 24 * time.Hour,
		},
		RateLimit: config.RateLimitConfig{
			Enabled: true,
			Global: config.RateLimitRule{
				Rate:   100,
				Burst:  200,
				Window: time.Minute,
			},
		},
		CORS: config.CORSConfig{
			Enabled:      true,
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders: []string{"*"},
		},
	},
	Services: map[string]config.ServiceConfig{
		"auth-service": {
			Name: "auth-service",
			Routes: []config.RouteConfig{
				{
					Path:    "/api/auth/*",
					Method:  "ANY",
					Service: "auth-service",
				},
			},
			LoadBalancer: config.LoadBalancerConfig{
				Type: "round_robin",
			},
			HealthCheck: config.HealthCheckConfig{
				Enabled:  true,
				Path:     "/health",
				Interval: 30 * time.Second,
				Timeout:  5 * time.Second,
			},
		},
		"user-service": {
			Name: "user-service",
			Routes: []config.RouteConfig{
				{
					Path:    "/api/users/*",
					Method:  "ANY",
					Service: "user-service",
				},
			},
			LoadBalancer: config.LoadBalancerConfig{
				Type: "round_robin",
			},
			HealthCheck: config.HealthCheckConfig{
				Enabled:  true,
				Path:     "/health",
				Interval: 30 * time.Second,
				Timeout:  5 * time.Second,
			},
		},
	},
}

// 模拟后端服务
func createMockBackendServer(name string, port int) *httptest.Server {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 健康检查端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": name,
			"time":    time.Now().Unix(),
		})
	})

	// 认证服务端点
	if name == "auth-service" {
		router.POST("/api/auth/login", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"token":   "mock-jwt-token",
				"expires": time.Now().Add(24 * time.Hour).Unix(),
			})
		})

		router.POST("/api/auth/validate", func(c *gin.Context) {
			token := c.GetHeader("Authorization")
			if token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"valid":  true,
				"userId": "test-user-123",
			})
		})
	}

	// 用户服务端点
	if name == "user-service" {
		router.GET("/api/users/:id", func(c *gin.Context) {
			id := c.Param("id")
			c.JSON(http.StatusOK, gin.H{
				"id":   id,
				"name": fmt.Sprintf("User %s", id),
				"email": fmt.Sprintf("user%s@example.com", id),
			})
		})

		router.POST("/api/users", func(c *gin.Context) {
			var user map[string]interface{}
			if err := c.ShouldBindJSON(&user); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			user["id"] = "new-user-123"
			c.JSON(http.StatusCreated, user)
		})
	}

	return httptest.NewServer(router)
}

// 测试套件
type GatewayTestSuite struct {
	gateway     *gateway.Gateway
	authServer  *httptest.Server
	userServer  *httptest.Server
	testServer  *httptest.Server
}

func setupTestSuite(t *testing.T) *GatewayTestSuite {
	// 创建模拟后端服务
	authServer := createMockBackendServer("auth-service", 8081)
	userServer := createMockBackendServer("user-service", 8082)

	// 创建服务注册表
	reg := registry.NewStaticRegistry()
	
	// 注册服务实例
	authInstance := &registry.ServiceInstance{
		ID:      "auth-1",
		Name:    "auth-service",
		Address: "127.0.0.1",
		Port:    8081,
		Health:  true,
		Weight:  1,
	}
	userInstance := &registry.ServiceInstance{
		ID:      "user-1",
		Name:    "user-service",
		Address: "127.0.0.1",
		Port:    8082,
		Health:  true,
		Weight:  1,
	}

	err := reg.Register(authInstance)
	require.NoError(t, err)
	err = reg.Register(userInstance)
	require.NoError(t, err)

	// 创建依赖组件
	log := logger.NewNoOpLogger()
	metrics := monitoring.NewNoOpMetrics()
	proxyManager := proxy.NewProxyManager(testConfig, reg, metrics, log)

	// 创建网关
	gw := gateway.New(testConfig, log, metrics, reg, proxyManager)

	// 创建测试服务器
	testServer := httptest.NewServer(gw.GetRouter())

	return &GatewayTestSuite{
		gateway:    gw,
		authServer: authServer,
		userServer: userServer,
		testServer: testServer,
	}
}

func (suite *GatewayTestSuite) tearDown() {
	suite.authServer.Close()
	suite.userServer.Close()
	suite.testServer.Close()
}

// 测试健康检查
func TestHealthCheck(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.tearDown()

	resp, err := http.Get(suite.testServer.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "healthy", result["status"])
}

// 测试就绪检查
func TestReadinessCheck(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.tearDown()

	resp, err := http.Get(suite.testServer.URL + "/ready")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "ready", result["status"])
}

// 测试路由代理
func TestRouteProxy(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.tearDown()

	// 测试认证服务路由
	loginData := map[string]string{
		"username": "testuser",
		"password": "testpass",
	}
	loginJSON, _ := json.Marshal(loginData)

	resp, err := http.Post(
		suite.testServer.URL+"/api/auth/login",
		"application/json",
		bytes.NewBuffer(loginJSON),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result, "token")
	assert.Contains(t, result, "expires")
}

// 测试用户服务路由
func TestUserServiceProxy(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.tearDown()

	// 测试获取用户
	resp, err := http.Get(suite.testServer.URL + "/api/users/123")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "123", result["id"])
	assert.Equal(t, "User 123", result["name"])
}

// 测试CORS
func TestCORS(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.tearDown()

	// 创建OPTIONS请求
	req, err := http.NewRequest("OPTIONS", suite.testServer.URL+"/api/users/123", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), "GET")
}

// 测试限流
func TestRateLimit(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.tearDown()

	// 发送大量请求测试限流
	client := &http.Client{}
	var successCount, rateLimitCount int

	for i := 0; i < 150; i++ {
		resp, err := client.Get(suite.testServer.URL + "/api/users/123")
		require.NoError(t, err)
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			successCount++
		} else if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitCount++
		}
	}

	// 应该有一些请求被限流
	assert.Greater(t, rateLimitCount, 0, "应该有请求被限流")
	assert.Greater(t, successCount, 0, "应该有请求成功")
}

// 测试监控指标
func TestMetrics(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.tearDown()

	// 先发送一些请求生成指标
	http.Get(suite.testServer.URL + "/api/users/123")
	http.Get(suite.testServer.URL + "/health")

	// 获取指标
	resp, err := http.Get(suite.testServer.URL + "/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/plain; version=0.0.4; charset=utf-8", resp.Header.Get("Content-Type"))
}

// 测试管理API
func TestAdminAPI(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.tearDown()

	// 测试获取路由列表
	resp, err := http.Get(suite.testServer.URL + "/admin/routes")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var routes []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&routes)
	require.NoError(t, err)

	assert.Greater(t, len(routes), 0, "应该有路由配置")
}

// 测试服务发现
func TestServiceDiscovery(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.tearDown()

	// 测试获取服务列表
	resp, err := http.Get(suite.testServer.URL + "/admin/services")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var services []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&services)
	require.NoError(t, err)

	assert.Greater(t, len(services), 0, "应该有服务实例")
}

// 基准测试
func BenchmarkGatewayProxy(b *testing.B) {
	suite := setupTestSuite(&testing.T{})
	defer suite.tearDown()

	client := &http.Client{}
	url := suite.testServer.URL + "/api/users/123"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(url)
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()
		}
	})
}

// 压力测试
func TestGatewayUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过压力测试")
	}

	suite := setupTestSuite(t)
	defer suite.tearDown()

	client := &http.Client{}
	url := suite.testServer.URL + "/api/users/123"

	// 并发测试
	concurrency := 50
	requests := 1000
	done := make(chan bool, concurrency)

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < requests/concurrency; j++ {
				resp, err := client.Get(url)
				if err != nil {
					t.Errorf("请求失败: %v", err)
					return
				}
				resp.Body.Close()
			}
		}()
	}

	// 等待所有goroutine完成
	for i := 0; i < concurrency; i++ {
		<-done
	}

	duration := time.Since(start)
	rps := float64(requests) / duration.Seconds()

	t.Logf("完成 %d 个请求，耗时 %v，RPS: %.2f", requests, duration, rps)

	// 验证性能指标
	assert.Greater(t, rps, 100.0, "RPS应该大于100")
}