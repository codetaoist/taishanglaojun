package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestContext holds request context information
type RequestContext struct {
	RequestID string
	TraceID   string
	UserID    string
	StartTime time.Time
}

// RequestIDMiddleware generates and adds a request ID to the context
type RequestIDMiddleware struct{}

// NewRequestIDMiddleware creates a new RequestIDMiddleware
func NewRequestIDMiddleware() *RequestIDMiddleware {
	return &RequestIDMiddleware{}
}

// Handle returns the Gin middleware handler
func (m *RequestIDMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID if not present
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add request ID to response header
		c.Header("X-Request-ID", requestID)

		// Add request ID to context
		c.Set("requestID", requestID)

		c.Next()
	}
}

// TracingMiddleware handles distributed tracing
type TracingMiddleware struct{}

// NewTracingMiddleware creates a new TracingMiddleware
func NewTracingMiddleware() *TracingMiddleware {
	return &TracingMiddleware{}
}

// Handle returns the Gin middleware handler
func (m *TracingMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get or generate trace ID
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// Add trace ID to response header
		c.Header("X-Trace-ID", traceID)

		// Add trace ID to context
		c.Set("traceID", traceID)

		// Create request context
		requestID, _ := c.Get("requestID")
		ctx := &RequestContext{
			RequestID: requestID.(string),
			TraceID:   traceID,
			StartTime: time.Now(),
		}

		// Add request context to Gin context
		c.Set("requestContext", ctx)

		c.Next()
	}
}

// LoggingMiddleware handles request logging
type LoggingMiddleware struct {
	logger *Logger
}

// NewLoggingMiddleware creates a new LoggingMiddleware
func NewLoggingMiddleware(logger *Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// Handle returns the Gin middleware handler
func (m *LoggingMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request context
		requestCtx, exists := c.Get("requestContext")
		if !exists {
			c.Next()
			return
		}

		ctx := requestCtx.(*RequestContext)

		// Log request
		m.logger.Infof("Request started: %s %s", c.Request.Method, c.Request.URL.Path)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(ctx.StartTime)

		// Log response
		m.logger.Infof("Request completed: %s %s - Status: %d - Duration: %v",
			c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration)
	}
}

// MetricsMiddleware handles request metrics
type MetricsMiddleware struct{}

// NewMetricsMiddleware creates a new MetricsMiddleware
func NewMetricsMiddleware() *MetricsMiddleware {
	return &MetricsMiddleware{}
}

// Handle returns the Gin middleware handler
func (m *MetricsMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request context
		requestCtx, exists := c.Get("requestContext")
		if !exists {
			c.Next()
			return
		}

		ctx := requestCtx.(*RequestContext)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(ctx.StartTime)

		// In a real implementation, you would record metrics to a metrics system
		// like Prometheus, InfluxDB, etc.
		// For now, we'll just log the metrics
		fmt.Printf("Metric: method=%s path=%s status=%d duration_ms=%d\n",
			c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration.Milliseconds())
	}
}

// RateLimitMiddleware handles rate limiting
type RateLimitMiddleware struct {
	// In a real implementation, you would use a rate limiter like
	// go-redis-rate-limit or golang.org/x/time/rate
}

// NewRateLimitMiddleware creates a new RateLimitMiddleware
func NewRateLimitMiddleware(config interface{}) *RateLimitMiddleware {
	return &RateLimitMiddleware{}
}

// Handle returns the Gin middleware handler
func (m *RateLimitMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// In a real implementation, you would check the rate limit
		// based on the client IP, user ID, or API key
		// For now, we'll just pass through
		c.Next()
	}
}

// CircuitBreakerMiddleware handles circuit breaking
type CircuitBreakerMiddleware struct {
	// In a real implementation, you would use a circuit breaker library
	// like github.com/sony/gobreaker
}

// NewCircuitBreakerMiddleware creates a new CircuitBreakerMiddleware
func NewCircuitBreakerMiddleware(config interface{}) *CircuitBreakerMiddleware {
	return &CircuitBreakerMiddleware{}
}

// Handle returns the Gin middleware handler
func (m *CircuitBreakerMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// In a real implementation, you would check the circuit breaker state
		// before allowing the request to proceed
		// For now, we'll just pass through
		c.Next()
	}
}

// TimeoutMiddleware handles request timeouts
type TimeoutMiddleware struct {
	timeout time.Duration
}

// NewTimeoutMiddleware creates a new TimeoutMiddleware
func NewTimeoutMiddleware(timeout time.Duration) *TimeoutMiddleware {
	return &TimeoutMiddleware{
		timeout: timeout,
	}
}

// Handle returns the Gin middleware handler
func (m *TimeoutMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), m.timeout)
		defer cancel()

		// Replace the request context
		c.Request = c.Request.WithContext(ctx)

		// Create a channel to signal completion
		finished := make(chan struct{})

		go func() {
			defer close(finished)
			c.Next()
		}()

		select {
		case <-finished:
			// Request completed normally
		case <-ctx.Done():
			// Request timed out
			c.JSON(http.StatusRequestTimeout, gin.H{
				"error": "Request timeout",
			})
			c.Abort()
		}
	}
}

// CORSMiddleware handles CORS
type CORSMiddleware struct{}

// NewCORSMiddleware creates a new CORSMiddleware
func NewCORSMiddleware() *CORSMiddleware {
	return &CORSMiddleware{}
}

// Handle returns the Gin middleware handler
func (m *CORSMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set CORS headers
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Request-ID, X-Trace-ID")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RecoveryMiddleware handles panic recovery
type RecoveryMiddleware struct {
	logger *Logger
}

// NewRecoveryMiddleware creates a new RecoveryMiddleware
func NewRecoveryMiddleware(logger *Logger) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		logger: logger,
	}
}

// Handle returns the Gin middleware handler
func (m *RecoveryMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				m.logger.Errorf("Panic recovered: %v", err)

				// Return error response
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})

				// Abort the request
				c.Abort()
			}
		}()

		c.Next()
	}
}

// AuthMiddleware handles authentication
type AuthMiddleware struct {
	// In a real implementation, you would use a JWT library or OAuth2
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

// Handle returns the Gin middleware handler
func (m *AuthMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// In a real implementation, you would validate the JWT token
		// For now, we'll just check if the header exists
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		// Extract token
		token := authHeader[7:]

		// In a real implementation, you would validate the token and extract user info
		// For now, we'll just add a placeholder user ID
		c.Set("userID", "user123")

		// Update request context
		if requestCtx, exists := c.Get("requestContext"); exists {
			ctx := requestCtx.(*RequestContext)
			ctx.UserID = "user123"
			c.Set("requestContext", ctx)
		}

		c.Next()
	}
}