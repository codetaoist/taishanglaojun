package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/auth/internal/config"
)

// RequestLogger is a middleware that logs requests
func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
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

// CORS is a middleware that handles CORS
func CORS(cfg config.Config) gin.HandlerFunc {
	allowed := cfg.AllowedOrigins
	allowAll := false
	for _, o := range allowed {
		if o == "*" {
			allowAll = true
			break
		}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			// Non-CORS request
			c.Next()
			return
		}
		allowedOrigin := ""
		if allowAll {
			allowedOrigin = origin
		} else {
			for _, o := range allowed {
				if strings.EqualFold(o, origin) {
					allowedOrigin = origin
					break
				}
			}
		}
		if allowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Methods", "GET,POST,DELETE,PUT,PATCH,OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,Accept,Origin,User-Agent,Cache-Control,X-Requested-With,Referer,"+cfg.TraceHeader)
			c.Header("Access-Control-Expose-Headers", cfg.TraceHeader)
			c.Header("Access-Control-Allow-Credentials", "false")
			if c.Request.Method == http.MethodOptions {
				c.Status(http.StatusNoContent)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// ErrorHandler is a middleware that handles errors and returns standardized responses
func ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		var err error
		var message string
		var statusCode int

		switch e := recovered.(type) {
		case string:
			err = nil
			message = e
			statusCode = http.StatusInternalServerError
		case error:
			err = e
			message = err.Error()
			statusCode = http.StatusInternalServerError
		default:
			err = nil
			message = "Unknown error"
			statusCode = http.StatusInternalServerError
		}

		// Log the error with stack trace
		log.Printf("Panic recovered: %v\n%s", message, debug.Stack())

		// Return error response
		c.JSON(statusCode, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Internal server error",
			"details": message,
		})
	})
}