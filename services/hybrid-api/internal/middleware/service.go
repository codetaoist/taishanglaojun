package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"codetaoist/api/internal/service"
)

// RequestContext represents the context of a request
type RequestContext struct {
	RequestID     string
	UserID        string
	TraceID       string
	StartTime     time.Time
	Headers       map[string]string
	ServiceName   string
	OperationName string
}

// NewRequestContext creates a new request context
func NewRequestContext(serviceName, operationName string) *RequestContext {
	return &RequestContext{
		RequestID:     uuid.New().String(),
		TraceID:       uuid.New().String(),
		StartTime:     time.Now(),
		Headers:       make(map[string]string),
		ServiceName:   serviceName,
		OperationName: operationName,
	}
}

// ServiceMiddleware provides middleware for service communication
type ServiceMiddleware struct {
	serviceManager *service.ServiceManager
	logger         *zap.Logger
}

// NewServiceMiddleware creates a new service middleware
func NewServiceMiddleware(serviceManager *service.ServiceManager, logger *zap.Logger) *ServiceMiddleware {
	return &ServiceMiddleware{
		serviceManager: serviceManager,
		logger:         logger,
	}
}

// RequestIDMiddleware adds a request ID to the context
func (m *ServiceMiddleware) RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("requestID", requestID)
		c.Next()
	}
}

// TraceMiddleware adds trace information to the context
func (m *ServiceMiddleware) TraceMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}

		c.Header("X-Trace-ID", traceID)
		c.Set("traceID", traceID)
		c.Set("serviceName", serviceName)
		c.Set("startTime", time.Now())
		c.Next()
	}
}

// LoggingMiddleware logs requests and responses
func (m *ServiceMiddleware) LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// MetricsMiddleware collects metrics for requests
func (m *ServiceMiddleware) MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		// Calculate metrics
		latency := time.Since(startTime)
		statusCode := c.Writer.Status()
		path := c.FullPath()
		method := c.Request.Method

		// Log metrics
		m.logger.Info("Request metrics",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
		)

		// In a real implementation, you would send these metrics to a metrics system
		// like Prometheus, InfluxDB, etc.
	}
}

// ServiceDiscoveryMiddleware injects service discovery information
func (m *ServiceMiddleware) ServiceDiscoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get service discovery information
		services, err := m.serviceManager.ListServices(c.Request.Context())
		if err != nil {
			m.logger.Error("Failed to list services", zap.Error(err))
			c.Next()
			return
		}

		// Add service discovery information to the context
		c.Set("services", services)
		c.Next()
	}
}

// CircuitBreakerMiddleware provides circuit breaker functionality
func (m *ServiceMiddleware) CircuitBreakerMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the service
		svc, err := m.serviceManager.GetService(c.Request.Context(), serviceName)
		if err != nil {
			m.logger.Error("Service not found", zap.String("service", serviceName), zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": fmt.Sprintf("Service %s not available", serviceName),
			})
			c.Abort()
			return
		}

		// Check service health
		health, err := m.serviceManager.HealthCheck(c.Request.Context(), svc.ID)
		if err != nil {
			m.logger.Error("Failed to check service health", zap.String("service", serviceName), zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": fmt.Sprintf("Service %s health check failed", serviceName),
			})
			c.Abort()
			return
		}

		// If service is unhealthy, return an error
		if health.Status != "healthy" {
			m.logger.Warn("Service is unhealthy", 
				zap.String("service", serviceName), 
				zap.String("status", health.Status),
				zap.String("message", health.Message),
			)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": fmt.Sprintf("Service %s is not healthy: %s", serviceName, health.Message),
			})
			c.Abort()
			return
		}

		// Service is healthy, continue with the request
		c.Next()
	}
}

// RateLimitMiddleware provides rate limiting functionality
func (m *ServiceMiddleware) RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	// This is a simple in-memory rate limiter
	// In a real implementation, you would use a more sophisticated rate limiter
	// like Redis or a dedicated rate limiting service
	requestCounts := make(map[string][]time.Time)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// Clean up old requests
		if timestamps, exists := requestCounts[clientIP]; exists {
			var validTimestamps []time.Time
			for _, timestamp := range timestamps {
				if now.Sub(timestamp) < time.Minute {
					validTimestamps = append(validTimestamps, timestamp)
				}
			}
			requestCounts[clientIP] = validTimestamps
		}

		// Check if the client has exceeded the rate limit
		if len(requestCounts[clientIP]) >= requestsPerMinute {
			m.logger.Warn("Rate limit exceeded", zap.String("clientIP", clientIP))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		// Add the current request
		requestCounts[clientIP] = append(requestCounts[clientIP], now)
		c.Next()
	}
}

// AuthMiddleware provides authentication functionality
func (m *ServiceMiddleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Check if the header is in the correct format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header format must be Bearer {token}",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// In a real implementation, you would validate the token
		// For now, we'll just check if it's not empty
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token is required",
			})
			c.Abort()
			return
		}

		// Extract user information from the token
		// In a real implementation, you would decode the JWT token
		userID := "user123" // This would come from the token

		c.Set("userID", userID)
		c.Next()
	}
}

// CORSMiddleware provides CORS functionality
func (m *ServiceMiddleware) CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Request-ID, X-Trace-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// TimeoutMiddleware provides timeout functionality
func (m *ServiceMiddleware) TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// RecoveryMiddleware recovers from panics
func (m *ServiceMiddleware) RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		m.logger.Error("Panic recovered",
			zap.Any("panic", recovered),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	})
}

// GetRequestContext gets the request context from the gin context
func GetRequestContext(c *gin.Context) *RequestContext {
	requestID, _ := c.Get("requestID")
	traceID, _ := c.Get("traceID")
	userID, _ := c.Get("userID")
	serviceName, _ := c.Get("serviceName")
	startTime, _ := c.Get("startTime")

	operationName := c.FullPath()
	if operationName == "" {
		operationName = c.Request.URL.Path
	}

	return &RequestContext{
		RequestID:     requestID.(string),
		TraceID:       traceID.(string),
		UserID:        userID.(string),
		StartTime:     startTime.(time.Time),
		Headers:       make(map[string]string),
		ServiceName:   serviceName.(string),
		OperationName: operationName,
	}
}

// LogRequest logs the request context
func LogRequest(logger *zap.Logger, ctx *RequestContext) {
	logger.Info("Request",
		zap.String("requestID", ctx.RequestID),
		zap.String("traceID", ctx.TraceID),
		zap.String("userID", ctx.UserID),
		zap.String("serviceName", ctx.ServiceName),
		zap.String("operationName", ctx.OperationName),
		zap.Time("startTime", ctx.StartTime),
	)
}