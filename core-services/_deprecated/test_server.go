package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("启动简化测试服务器...")
	
	// 设置Gin为发布模式
	gin.SetMode(gin.ReleaseMode)
	
	// 创建路由
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	
	// 添加测试路由
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})
	
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "太上老君核心服务测试版本",
			"version": "test-1.0.0",
		})
	})
	
	// 启动服务器
	fmt.Println("服务器启动在 http://localhost:8080")
	fmt.Println("健康检查: http://localhost:8080/health")
	
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}