package main

import (
	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/core-services/internal/config"
)

// corsMiddleware 配置CORS中间件
func corsMiddleware(corsConfig config.CORSConfig) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 设置CORS头
		for _, origin := range corsConfig.AllowedOrigins {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}

// requestIDMiddleware 添加请求ID中间件
func requestIDMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 简单的请求ID生成
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = "req-" + c.Request.Header.Get("X-Forwarded-For")
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	})
}
