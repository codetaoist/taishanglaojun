package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/response"
)

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
		response.Error(c, statusCode, 500, "Internal server error", message)
	})
}

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