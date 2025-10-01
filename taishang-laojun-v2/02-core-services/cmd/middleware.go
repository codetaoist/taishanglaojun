package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/taishanglaojun/core-services/internal/config"
)

// corsMiddleware 配置CORS中间件
func corsMiddleware(corsConfig config.CORSConfig) gin.HandlerFunc {
	config := cors.Config{
		AllowOrigins:     corsConfig.AllowedOrigins,
		AllowMethods:     corsConfig.AllowedMethods,
		AllowHeaders:     corsConfig.AllowedHeaders,
		ExposeHeaders:    corsConfig.ExposedHeaders,
		AllowCredentials: corsConfig.AllowCredentials,
		MaxAge:           corsConfig.MaxAge,
	}

	return cors.New(config)
}

// requestIDMiddleware 添加请求ID中间件
func requestIDMiddleware() gin.HandlerFunc {
	return requestid.New()
}
