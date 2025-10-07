package middleware

import (
	"context"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

// RequestIDKey 请求ID上下文键
type RequestIDKey struct{}

// LoggingMiddleware 日志中间件
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 创建响应写入器包装器来捕获状态码
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// 处理请求
		next.ServeHTTP(wrapper, r)

		// 记录日志
		duration := time.Since(start)
		log.Printf(
			"[%s] %s %s %d %v",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			wrapper.statusCode,
			duration,
		)
	})
}

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v\n%s", err, debug.Stack())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// RequestIDMiddleware 请求ID中间件
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 生成请求ID
		requestID := uuid.New().String()

		// 添加到上下文
		ctx := context.WithValue(r.Context(), RequestIDKey{}, requestID)
		r = r.WithContext(ctx)

		// 添加到响应头
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware 认证中间件
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 获取Authorization头
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// 这里应该验证JWT token或其他认证方式
		// 为了演示，我们简单检查是否以"Bearer "开头
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		// 在实际应用中，这里应该解析和验证token
		// token := authHeader[7:]
		// userID, err := validateToken(token)
		// if err != nil {
		//     http.Error(w, "Invalid token", http.StatusUnauthorized)
		//     return
		// }

		// 将用户ID添加到上下文（这里使用模拟值）
		// ctx := context.WithValue(r.Context(), "userID", userID)
		// r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware CORS中间件
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置CORS头
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// 处理预检请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
	// 简单的内存限流器（生产环境应使用Redis等）
	clients := make(map[string][]time.Time)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := r.RemoteAddr

			now := time.Now()
			windowStart := now.Add(-time.Minute)

			// 清理过期的请求记录
			if requests, exists := clients[clientIP]; exists {
				validRequests := make([]time.Time, 0)
				for _, reqTime := range requests {
					if reqTime.After(windowStart) {
						validRequests = append(validRequests, reqTime)
					}
				}
				clients[clientIP] = validRequests
			}

			// 检查是否超过限制
			if len(clients[clientIP]) >= requestsPerMinute {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			// 记录当前请求
			clients[clientIP] = append(clients[clientIP], now)

			next.ServeHTTP(w, r)
		})
	}
}

// ValidationMiddleware 验证中间件
func ValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查Content-Type
		if r.Method == "POST" || r.Method == "PUT" {
			contentType := r.Header.Get("Content-Type")
			if contentType != "application/json" {
				http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// responseWriter 响应写入器包装器
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader 写入状态码
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// GetRequestID 从上下文获取请求ID
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey{}).(string); ok {
		return requestID
	}
	return ""
}