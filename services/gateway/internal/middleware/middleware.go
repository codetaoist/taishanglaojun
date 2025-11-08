package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// Logging logs each request
func Logging(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get client IP
		clientIP := c.ClientIP()

		// Get status code
		statusCode := c.Writer.Status()

		// Get request ID if available
		requestID, _ := c.Get("RequestID")

		// Build log fields
		fields := []zap.Field{
			zap.String("request_id", requestID.(string)),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", raw),
			zap.String("ip", clientIP),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		// Add error if there is one
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("error", c.Errors.String()))
		}

		// Log based on status code
		if statusCode >= 500 {
			logger.Error("Server error", fields...)
		} else if statusCode >= 400 {
			logger.Warn("Client error", fields...)
		} else {
			logger.Info("Request", fields...)
		}
	}
}

// Metrics collects metrics for each request
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record start time
		start := time.Now()

		// Process request
		c.Next()

		// Record metrics
		// This is a placeholder implementation
		// In a real implementation, we would use Prometheus or another metrics system
		statusCode := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		path := c.Request.URL.Path

		// Increment request counter
		requestCounter.WithLabelValues(method, path, statusCode).Inc()

		// Record request duration
		duration := time.Since(start).Seconds()
		requestDuration.WithLabelValues(method, path).Observe(duration)
	}
}

// Authentication validates JWT tokens
func Authentication(jwtConfig JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Check if the header has the correct format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header format must be Bearer {token}",
			})
			c.Abort()
			return
		}

		// Parse and validate token
		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtConfig.Secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", claims["user_id"])
			c.Set("username", claims["username"])
			c.Set("roles", claims["roles"])
		}

		c.Next()
	}
}

// RateLimit implements rate limiting
func RateLimit(config RateLimitConfig) gin.HandlerFunc {
	// Create a new rate limiter
	limiter := NewRateLimiter(config.Requests, config.Window)

	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()

		// Check if the client has exceeded the rate limit
		if !limiter.Allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// PrometheusHandler returns a Prometheus metrics handler
func PrometheusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This is a placeholder implementation
		// In a real implementation, we would use the Prometheus client library
		c.String(http.StatusOK, "# HELP http_requests_total Total number of HTTP requests\n# TYPE http_requests_total counter\n")
	}
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

// RateLimitConfig holds rate limit configuration
type RateLimitConfig struct {
	Requests int           `mapstructure:"requests"`
	Window   time.Duration `mapstructure:"window"`
}

// RateLimiter implements a simple rate limiter
type RateLimiter struct {
	requests int
	window   time.Duration
	clients  map[string]*ClientLimiter
	mutex    sync.RWMutex
}

// ClientLimiter holds rate limit information for a client
type ClientLimiter struct {
	count     int
	lastReset time.Time
	mutex     sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: requests,
		window:   window,
		clients:  make(map[string]*ClientLimiter),
	}
}

// Allow checks if a client is allowed to make a request
func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.mutex.RLock()
	limiter, exists := rl.clients[clientIP]
	rl.mutex.RUnlock()

	if !exists {
		rl.mutex.Lock()
		// Double-check after acquiring write lock
		limiter, exists = rl.clients[clientIP]
		if !exists {
			limiter = &ClientLimiter{}
			rl.clients[clientIP] = limiter
		}
		rl.mutex.Unlock()
	}

	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	now := time.Now()

	// Reset the counter if the window has passed
	if now.Sub(limiter.lastReset) > rl.window {
		limiter.count = 0
		limiter.lastReset = now
	}

	// Check if the client has exceeded the rate limit
	if limiter.count >= rl.requests {
		return false
	}

	// Increment the counter
	limiter.count++
	return true
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// This is a simple implementation
	// In a real implementation, we would use a more sophisticated method
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Metrics counters and histograms
var (
	requestCounter = NewCounter("http_requests_total", "Total number of HTTP requests", "method", "path", "status")
	requestDuration = NewHistogram("http_request_duration_seconds", "HTTP request duration in seconds", "method", "path")
)

// Counter is a simple counter implementation
type Counter struct {
	name  string
	help  string
	value int64
}

// NewCounter creates a new counter
func NewCounter(name, help string, labels ...string) *Counter {
	return &Counter{
		name: name,
		help: help,
	}
}

// Inc increments the counter
func (c *Counter) Inc() {
	c.value++
}

// WithLabelValues returns a counter with label values
func (c *Counter) WithLabelValues(values ...string) *Counter {
	// This is a placeholder implementation
	// In a real implementation, we would handle labels properly
	return c
}

// Histogram is a simple histogram implementation
type Histogram struct {
	name  string
	help  string
	value float64
}

// NewHistogram creates a new histogram
func NewHistogram(name, help string, labels ...string) *Histogram {
	return &Histogram{
		name: name,
		help: help,
	}
}

// Observe records a value
func (h *Histogram) Observe(value float64) {
	h.value = value
}

// WithLabelValues returns a histogram with label values
func (h *Histogram) WithLabelValues(values ...string) *Histogram {
	// This is a placeholder implementation
	// In a real implementation, we would handle labels properly
	return h
}