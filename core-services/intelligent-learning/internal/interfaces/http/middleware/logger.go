package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger 日志中间�?
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 获取请求ID
		requestID := param.Request.Header.Get("X-Request-ID")

		// 记录请求日志
		logrus.WithFields(logrus.Fields{
			"request_id":    requestID,
			"method":        param.Method,
			"path":          param.Path,
			"status":        param.StatusCode,
			"latency":       param.Latency,
			"client_ip":     param.ClientIP,
			"user_agent":    param.Request.UserAgent(),
			"response_size": param.BodySize,
		}).Info("HTTP Request")

		return ""
	})
}

// RequestLogger 详细请求日志中间�?
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时�?
		start := time.Now()

		// 获取请求ID
		requestID := requestid.Get(c)

		// 读取请求�?
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 创建响应写入器包装器
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:          &bytes.Buffer{},
		}
		c.Writer = writer

		// 记录请求开�?
		logrus.WithFields(logrus.Fields{
			"request_id":   requestID,
			"method":       c.Request.Method,
			"path":         c.Request.URL.Path,
			"query":        c.Request.URL.RawQuery,
			"client_ip":    c.ClientIP(),
			"user_agent":   c.Request.UserAgent(),
			"content_type": c.Request.Header.Get("Content-Type"),
			"request_body": string(requestBody),
		}).Info("Request started")

		// 处理请求
		c.Next()

		// 计算延迟
		latency := time.Since(start)

		// 记录请求结束
		logrus.WithFields(logrus.Fields{
			"request_id":     requestID,
			"method":         c.Request.Method,
			"path":           c.Request.URL.Path,
			"status":         c.Writer.Status(),
			"latency":        latency,
			"response_size":  c.Writer.Size(),
			"response_body":  writer.body.String(),
		}).Info("Request completed")
	}
}

// responseWriter 响应写入器包装器
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
